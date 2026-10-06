#!/usr/bin/env python3
"""Run Redalert with an isolated loopback fixture and temporary demo config."""

import http.server
import os
import pathlib
import socket
import signal
import subprocess
import sys
import tempfile
import threading
import time
import urllib.error
import urllib.request


def request(url, timeout=1):
    with urllib.request.urlopen(url, timeout=timeout) as response:
        return response.read()


def ensure_available(port, label):
    try:
        with socket.socket() as sock:
            sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            sock.bind(("127.0.0.1", port))
    except OSError as error:
        raise RuntimeError("%s port %d is unavailable: %s" % (label, port, error))


def main():
    if len(sys.argv) != 2:
        raise SystemExit("usage: dev.py PATH_TO_REDALERT")
    web_port = int(os.environ.get("REDALERT_DEV_PORT", "8888"))
    rpc_port = int(os.environ.get("REDALERT_DEV_RPC_PORT", "8889"))
    fixture_port = int(os.environ.get("REDALERT_DEV_FIXTURE_PORT", "0"))
    if web_port == rpc_port:
        raise RuntimeError("dashboard and RPC ports must be different")
    ensure_available(web_port, "dashboard")
    ensure_available(rpc_port, "RPC")
    state = {"healthy": True}

    class Handler(http.server.BaseHTTPRequestHandler):
        def do_GET(self):
            if self.path == "/fail":
                state["healthy"] = False
            elif self.path == "/recover":
                state["healthy"] = True
            body = b"healthy" if state["healthy"] else b"failing"
            self.send_response(200)
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)

        def log_message(self, *_args):
            pass

    try:
        fixture = http.server.ThreadingHTTPServer(("127.0.0.1", fixture_port), Handler)
    except OSError as error:
        raise RuntimeError("fixture port %d is unavailable: %s" % (fixture_port, error))
    fixture_port = fixture.server_address[1]
    fixture_thread = threading.Thread(target=fixture.serve_forever, daemon=True)
    fixture_thread.start()
    child = None
    stop = threading.Event()
    for sig in (signal.SIGINT, signal.SIGTERM):
        signal.signal(sig, lambda *_: stop.set())
    try:
        template = pathlib.Path(__file__).resolve().parents[1] / "config/demo.json"
        with tempfile.TemporaryDirectory(prefix="redalert-dev-") as temp:
            config = pathlib.Path(temp) / "config.json"
            config.write_text(template.read_text().replace("__FIXTURE_PORT__", str(fixture_port)))
            child = subprocess.Popen([
                str(pathlib.Path(sys.argv[1]).resolve()), "server", "--config-file", str(config),
                "--port", str(web_port), "--rpc-port", str(rpc_port),
            ])
            base = "http://127.0.0.1:%d" % web_port
            deadline = time.monotonic() + 20
            while time.monotonic() < deadline:
                if child.poll() is not None:
                    raise RuntimeError("Redalert exited during startup with status %d" % child.returncode)
                try:
                    if request(base + "/healthcheck") == b"OK":
                        break
                except (OSError, urllib.error.URLError):
                    pass
                time.sleep(.2)
            else:
                raise RuntimeError("timed out waiting for dashboard; ports may be occupied")
            print("Dashboard: %s/" % base, flush=True)
            print("Fixture:   http://127.0.0.1:%d/" % fixture_port, flush=True)
            print("Open /fail on the fixture URL to trigger an alert; open /recover to recover.", flush=True)
            print("Press Ctrl-C to stop both processes.", flush=True)
            while not stop.wait(.2):
                if child.poll() is not None:
                    raise RuntimeError("Redalert exited unexpectedly with status %d" % child.returncode)
    finally:
        if child is not None and child.poll() is None:
            child.send_signal(signal.SIGINT)
            try:
                child.wait(timeout=4)
            except subprocess.TimeoutExpired:
                child.terminate()
                try:
                    child.wait(timeout=2)
                except subprocess.TimeoutExpired:
                    child.kill()
                    child.wait()
        fixture.shutdown()
        fixture.server_close()
        fixture_thread.join(timeout=2)


if __name__ == "__main__":
    try:
        main()
    except Exception as error:
        print("redalert dev: %s" % error, file=sys.stderr)
        raise SystemExit(1)
