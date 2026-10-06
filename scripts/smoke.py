#!/usr/bin/env python3
"""Build and exercise the standalone server from outside the source tree."""

import http.server
import json
import os
import pathlib
import socket
import subprocess
import sys
import tempfile
import threading
import time
import urllib.error
import urllib.request


def free_port():
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        return sock.getsockname()[1]


def request(url, method="GET"):
    req = urllib.request.Request(url, method=method)
    with urllib.request.urlopen(req, timeout=2) as response:
        return response.status, response.read(), response.headers


def wait_for(description, action, timeout=25, process=None, logs=None):
    deadline = time.monotonic() + timeout
    last = None
    while time.monotonic() < deadline:
        if process is not None and process.poll() is not None:
            details = ""
            if logs:
                details = "\nprocess logs:\n" + "\n".join(
                    path + ":\n" + pathlib.Path(path).read_text()
                    for path in logs)
            raise RuntimeError("server exited early with status %d while waiting for %s%s" %
                               (process.returncode, description, details))
        try:
            last = action()
            if last:
                return last
        except (OSError, ValueError, urllib.error.URLError):
            pass
        time.sleep(0.2)
    raise RuntimeError("timed out waiting for " + description + "; last result: " + repr(last))


def main():
    binary = pathlib.Path(sys.argv[1]).resolve()
    mode = {"body": "healthy"}

    class Target(http.server.BaseHTTPRequestHandler):
        def do_GET(self):
            body = mode["body"].encode()
            self.send_response(200)
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)

        def log_message(self, *_args):
            pass

    target = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Target)
    threading.Thread(target=target.serve_forever, daemon=True).start()
    web_port, rpc_port = free_port(), free_port()
    with tempfile.TemporaryDirectory(prefix="redalert-smoke-") as temp:
        work = pathlib.Path(temp)
        config = {
            "checks": [{
                "id": "smoke", "name": "smoke", "type": "web-ping",
                "send_alerts": ["stderr"],
                "backoff": {"type": "constant", "interval": 1},
                "config": {"address": "http://127.0.0.1:%d/" % target.server_port},
                "assertions": [{"source": "text", "comparison": "==", "target": "healthy"}],
            }],
            "notifications": [],
            "preferences": {"notifications": {"fail_count_alert_threshold": 1}},
        }
        config_path = work / "config.json"
        config_path.write_text(json.dumps(config))
        version = subprocess.run([str(binary), "version"], cwd=work, check=True,
                                 text=True, capture_output=True).stdout
        if "Redalert v" not in version:
            raise RuntimeError("unexpected version output: " + version)
        stdout_log = work / "server.stdout"
        stderr_log = work / "server.stderr"
        process = subprocess.Popen(
            [str(binary), "server", "--config-file", str(config_path),
             "--port", str(web_port), "--rpc-port", str(rpc_port)],
            cwd=work, stdout=stdout_log.open("w"), stderr=stderr_log.open("w"), text=True)
        base = "http://127.0.0.1:%d" % web_port
        try:
            wait_for("healthcheck", lambda: request(base + "/healthcheck")[1] == b"OK",
                     process=process, logs=[str(stdout_log), str(stderr_log)])
            status, html, headers = request(base + "/")
            assert status == 200 and b"<html" in html.lower() and headers.get_content_type() == "text/html", "dashboard HTML was not served"
            status, js, headers = request(base + "/assets/app.bundle.js")
            assert status == 200 and len(js) > 1000 and js.startswith(b"!function"), "embedded JavaScript bundle was not served"
            assert headers.get_content_type() in ("application/javascript", "text/javascript"), "JavaScript content type was not preserved"

            def stats():
                status, body, headers = request(base + "/v1/stats")
                parsed = json.loads(body)
                return parsed if status == 200 and headers.get_content_type() == "application/json" and len(parsed) == 1 else None

            def current_stats():
                result = stats()
                return result[0] if result else None

            current = wait_for("initial healthy check", lambda: current_stats()
                               if current_stats() and current_stats()["status"] == "SUCCESSFUL" else None)
            before_trigger = current["stats"]["successful_total"]
            request(base + "/v1/checks/smoke/trigger", "POST")
            wait_for("triggered check", lambda: stats()[0]["stats"]["successful_total"] > before_trigger)

            mode["body"] = "failing"
            request(base + "/v1/checks/smoke/trigger", "POST")
            wait_for("failure", lambda: stats()[0]["status"] == "FAILING")
            mode["body"] = "healthy"
            request(base + "/v1/checks/smoke/trigger", "POST")
            wait_for("recovery", lambda: stats()[0]["status"] == "SUCCESSFUL")

            request(base + "/v1/checks/smoke/disable", "POST")
            wait_for("disabled check", lambda: stats()[0]["status"] == "DISABLED")
            request(base + "/v1/checks/smoke/enable", "POST")
            wait_for("enabled check", lambda: stats()[0]["status"] == "SUCCESSFUL")
            process.terminate()
            process.wait(timeout=3)
            stderr = stderr_log.read_text()
            if "fail:" not in stderr or "recovery - check is now successful" not in stderr:
                raise RuntimeError("failure/recovery alerts were not logged: " + stderr)
            print("smoke passed: version, embedded dashboard, health, stats, failure/recovery alerts, trigger, disable, enable")
        finally:
            process.terminate()
            try:
                process.wait(timeout=3)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait()
            target.shutdown()
            target.server_close()


if __name__ == "__main__":
    main()
