## Redalert

[![CI](https://github.com/jonog/redalert/actions/workflows/ci.yml/badge.svg)](https://github.com/jonog/redalert/actions/workflows/ci.yml)

[![Launch Stack](https://cdn.rawgit.com/buildkite/cloudformation-launch-stack-button-svg/master/launch-stack.svg)](https://console.aws.amazon.com/cloudformation/home#/stacks/new?stackName=redalert&templateURL=https://s3-ap-southeast-2.amazonaws.com/redalert-cloudformation/redalert.yml)

For monitoring your infrastructure and sending notifications if stuff is not ok.
(e.g. pinging your websites/APIs via HTTP GET at specified intervals, and alerting you if there is downtime).

<img src="https://cloud.githubusercontent.com/assets/1314353/23970218/34e46d0a-0a1d-11e7-8af0-6db94f69f0a9.png" width="500">

### Features

#### Checks
* *Website monitoring* & latency measurement (check type: `web-ping`)
* *Server metrics* from local machine (check type: `scollector`)
* *Docker container metrics* (check type: `docker-stats`)
* *Docker container metrics* from remote host via SSH (check type: `remote-docker`)
* *Postgres counts/stats* via SQL queries (check type: `postgres`)
* *TCP connectivity monitoring* & latency measurement (check type: `tcp`)
* *Execute local commands* & capture output (check type: `command`)
* *Execute remote commands via SSH* & capture output (check type: `remote-command`)
* *Run test suite and capture report metrics* via `JUnit XML` format (check type: `test-report`)

Checks will happen at specified intervals or explicit trigger (i.e. trigger check API endpoint).

#### Dashboard and Alerts
* Alert notifications available on several channels:
  * sending email (`gmail`)
  * sending SMS (`twilio`)
  * posting a message to Slack (`slack`)
  * unix stream (`stderr`)
* Provides ping status & latency info to `stdout`.
* Adjustable back-off after a check fails (constant, linear, exponential - see notes below).
* Includes a web UI as indicated by the screenshot above. (visit localhost:8888/, configure port via cli flag)
* Triggers a failure alert (`redalert`) when a check is failing, and a recovery alert (`greenalert`) when the check has recovered (e.g. a successful ping, following a failing ping).
* Triggers an alert when specified metric is above/below threshold.

#### Assertions
* Assertions are used to define criteria for checks to pass or fail:
* Assert on metrics
  * source: `metric`
  * `>` or `greater than`
  * `>=` or `greater than or equal`
  * `<` or `less than`
  * `<=` or `less than or equal`
  * `==` or `=` or `equals`
* Assert on metadata
  * source: `metadata`
  * `web-ping` returns `status_code`
* Assert on response
  * source: `text`
  * source: `json`

#### API

| Endpoint | Description |
| --- | --- |
| `GET /v1/stats` | Retrieve stats for all checks |
| `POST /v1/checks/{check_id}/disable` | Disable check |
| `POST /v1/checks/{check_id}/enable` | Enable check |
| `POST /v1/checks/{check_id}/trigger` | Trigger check |


### Design

```

         ┌──────────────────────────────┐
         │                              │
   ┌────▶│     Redalert Check Flow      │
   │     │                              │
   │     └──────────────────────────────┘
   │                    │
   │          @interval or ->trigger   ┌──────────────────────┐
   │                    │            ┌▶│  error during check  │
   │                    ▼            │ └──────────────────────┘
   │        ┌──────────────────────┐ │ ┌──────────────────────┐
   │        │  is check failing?   │─┤ │  failing assertions  │
   │        └──────────────────────┘ │ │     * metrics *      │
   │                    │            └▶│     * metadata *     │
   │          ┌───YES───┴───NO────┐    │     * response *     │
   │          │                   │    └──────────────────────┘
   │          ▼                   ▼
   │  ┌───────────────┐   ┌───────────────┐
   │  │send alerts via│   │   is check    │
   │  │   notifiers   │   │  recovering?  │
   │  └───────────────┘   └───────────────┘
   │  ┌───────────────┐          YES
   │  │adjust backoff │           │
   │  └───────────────┘           ▼
   │          │           ┌───────────────┐
   │          │           │send alerts via│
   │          │           │   notifiers   │
   │          │           └───────────────┘
   │          │           ┌───────────────┐
   │          │           │ reset backoff │
   │          │           └───────────────┘
   │          │                   │
   │          ▼                   ▼
   │         ┌──────────────────────┐
   └─────────│    Event Storage     │
             └──────────────────────┘
```

#### Screenshots
![](https://cloud.githubusercontent.com/assets/1314353/5157264/edb21476-733a-11e4-8452-4b96b443f7ee.jpg)

### Getting started
Run via Docker:
```
docker run -d -P -v /path/to/config.json:/config.json jonog/redalert
```
Quick bootstrap example:
```
curl https://gist.githubusercontent.com/jonog/32c953aedf03edf71acaef53d89ce785/raw/e87f7e933165574e1d441781465223bfe6c3f1aa/sample_redalert_config.json > /tmp/sample_redalert_config.json && \
    docker run -d -P -v /tmp/sample_redalert_config.json:/config.json --name test_redalert jonog/redalert && \
    open "http://$(docker port test_redalert 8888)"
```



#### Usage
Get started with the `redalert` command:
```
Usage:
  redalert [command]

Available Commands:
  checks      List checks
  events      List retained events for a check
  config-sync Sync file and database configurations
  server      Run checks and server stats
  version     Print the version number of Redalert

Flags:
  -d, --config-db string     config database url
  -f, --config-file string   config file (default "config.json")
  -s, --config-s3 string     config S3
  -u, --config-url string    config url
  -h, --help                 help for redalert
  -p, --port int             port to run web server (default 8888)
  -r, --rpc-port int         port to run RPC server (default 8889)

Use "redalert [command] --help" for more information about a command.
```

Use `redalert events <check-id>` to display the events currently retained by the running server for one check. The in-memory event store is bounded, so older evicted events and events from previous server processes are unavailable. The command uses the configured `--rpc-port` (default 8889).

Use `redalert check-add --input checks.json` to append one or more checks to the running server's file configuration and start them immediately. The input is a nonempty JSON array using the `checks` entries from `config.json`; omitted IDs are generated and the command prints the IDs that were added. Use `--input -` to read the array from standard input. For example:

```json
[
  {
    "name": "Local HTTP",
    "type": "web-ping",
    "config": { "address": "http://localhost:8080" },
    "send_alerts": ["stderr"],
    "backoff": { "type": "constant", "interval": 10 },
    "assertions": []
  }
]
```

`check-add` requires a running server using file configuration, and `--config-file` must identify the same file used by that server. Database, URL, and S3 configurations are not supported.

#### Configuration

Configure servers to monitor & alert settings via a configuration file:
* a local file (specified by `-f` or `--config-file`) - defaults to `config.json`
* a file remotely accessible via HTTP (specified by `-u` or `--config-url`)
* a file hosted in an AWS S3 bucket (specified by `-s` or `--config-s3`)

All three sources accept JSON or YAML. Files ending in `.yaml` or `.yml` (case-insensitive) use YAML; other and extensionless paths use JSON. For HTTP and S3, selection uses the URL path or object key, so query parameters do not affect it. YAML must contain one configuration document with string mapping keys and JSON-compatible values. Quote string values that resemble numbers or booleans, such as assertion targets and notifier settings. A YAML example is available in [config/config.yaml.sample](config/config.yaml.sample).

TODO: document Postgres configuration option

##### Example config.json
```
{
   "checks":[
      {
         "name":"Google",
         "type": "web-ping",
         "config": {
            "address":"http://google.com"
         },
         "send_alerts": ["stderr"],
         "backoff": {
            "type": "constant",
            "interval": 10
         },
         "assertions": [
             {
                 "comparison": "==",
                 "identifier": "status_code",
                 "source": "metadata",
                 "target": "200"
             }
         ]
      }
   ],
   "notifications": []
}
```

##### Example Larger config.json
```
{
    "checks": [
        {
            "name": "Demo HTTP Status Check",
            "type": "web-ping",
            "config": {
                "address": "http://httpstat.us/200",
                "headers": {
                    "X-Api-Key": "ABCD1234"
                }
            },
            "send_alerts": [
                "stderr"
            ],
            "backoff": {
                "interval": 10,
                "type": "constant"
            },
            "assertions": [
                {
                    "comparison": "==",
                    "identifier": "status_code",
                    "source": "metadata",
                    "target": "200"
                }
            ]
        },
        {
            "name": "Demo Response Check",
            "type": "web-ping",
            "config": {
                "address": "http://httpstat.us/400"
            },
            "send_alerts": [
                "stderr",
                "email",
                "chat",
                "sms"
            ],
            "backoff": {
                "interval": 10,
                "type": "linear"
            },
            "assertions": [
                {
                    "comparison": "less than",
                    "identifier": "latency",
                    "source": "metric",
                    "target": "1100"
                },
                {
                    "comparison": "==",
                    "identifier": "status_code",
                    "source": "metadata",
                    "target": "400"
                },
                {
                    "comparison": "==",
                    "source": "text",
                    "target": "400 Bad Request"
                }
            ],
            "verbose_logging": true
        },
        {
            "name": "Demo Exponential Backoff",
            "type": "web-ping",
            "config": {
                "address": "http://httpstat.us/200"
            },
            "send_alerts": [
                "stderr"
            ],
            "backoff": {
                "interval": 10,
                "multiplier": 2,
                "type": "exponential"
            },
            "assertions": [
                {
                    "comparison": "==",
                    "identifier": "status_code",
                    "source": "metadata",
                    "target": "500"
                }
            ]
        },
        {
            "name": "Docker Redis",
            "type": "tcp",
            "config": {
                "host": "192.168.99.100",
                "port": 1001
            },
            "send_alerts": [
                "stderr"
            ],
            "backoff": {
                "interval": 10,
                "type": "constant"
            }
        },
        {
            "name": "Docker stats",
            "type": "docker-stats",
            "config": {},
            "send_alerts": [
                "stderr"
            ],
            "backoff": {
                "interval": 30,
                "type": "linear"
            }
        },
        {
            "name": "production-docker-host",
            "type": "remote-docker",
            "config": {
                "host": "ec2-xx-xxx-xx-xxx.ap-southeast-1.compute.amazonaws.com",
                "user": "ubuntu"
            },
            "send_alerts": [
                "stderr"
            ],
            "backoff": {
                "interval": 5,
                "type": "linear"
            }
        },
        {
            "name": "scollector-metrics",
            "type": "scollector",
            "config": {
                "host": "hostname"
            },
            "send_alerts": [
                "stderr"
            ],
            "backoff": {
                "interval": 15,
                "type": "constant"
            }
        },
        {
            "name": "production-db",
            "type": "postgres",
            "config": {
                "connection_url": "postgres://user:pass@localhost:5432/dbname?sslmode=disable",
                "metric_queries": [
                    {
                        "metric": "client_count",
                        "query": "select count(*) from clients"
                    }
                ]
            },
            "send_alerts": [
                "stderr"
            ],
            "backoff": {
                "interval": 120,
                "type": "linear"
            }
        },
        {
            "name": "README size",
            "type": "command",
            "config": {
                "command": "cat README.md | wc -l",
                "output_type": "number"
            },
            "send_alerts": [
                "stderr"
            ],
            "backoff": {
                "interval": 10,
                "type": "constant"
            }
        },
        {
            "name": "List files",
            "type": "command",
            "config": {
                "command": "ls"
            },
            "send_alerts": [
                "stderr"
            ],
            "backoff": {
                "interval": 10,
                "type": "constant"
            }
        },
        {
            "name": "SHH into docker-alpine-sshd",
            "type": "remote-command",
            "config": {
                "command": "uptime",
                "ssh_auth_options": {
                  "user": "root",
                  "password": "root",
                  "host": "localhost",
                  "port": 2222
                }
            },
            "send_alerts": [
                "stderr"
            ],
            "assertions": [
                {
                    "comparison": "==",
                    "identifier": "exit_status",
                    "source": "metadata",
                    "target": "0"
                }
            ]
        },
        {
            "name": "Run Smoke Tests",
            "type": "test-report",
            "config": {
                "command": "./run-smoke-tests.sh"
            },
            "send_alerts": [
                "stderr"
            ],
            "assertions": [
                {
                    "comparison": "==",
                    "identifier": "status",
                    "source": "metadata",
                    "target": "PASSING"
                }
            ]
        }
    ],
    "notifications": [
        {
            "name": "email",
            "type": "gmail",
            "config": {
                "notification_addresses": "",
                "pass": "",
                "user": ""
            }
        },
        {
            "name": "chat",
            "type": "slack",
            "config": {
                "channel": "#general",
                "icon_emoji": ":rocket:",
                "username": "redalert",
                "webhook_url": ""
            }
        },
        {
            "name": "sms",
            "type": "twilio",
            "config": {
                "account_sid": "",
                "auth_token": "",
                "notification_numbers": "",
                "twilio_number": ""
            }
        }
    ],
    "preferences": {
        "notifications": {
          "fail_count_alert_threshold": 2,
          "repeat_fail_alerts": false
        }
    }
}

```

Build and run (capture stderr).
```
go build -o redalert .

./redalert 2> errors.log
```

#### Notification Preferences
* `fail_count_alert_threshold` controls sending an alert, only after N fails (defaults to 1)
* `repeat_fail_alerts` controls whether fail alerts are repeated, on consecutive failing checks (defaults to false)
```
"preferences": {
  "notifications": {
    "fail_count_alert_threshold": 2,
    "repeat_fail_alerts": false
  }
}
```

#### Backoffs
When a server check fails - the next check will be delayed according to the back-off algorithm. By default, there is no delay (i.e. `constant` back-off), with a default interval of 10 seconds between checks. When a failing server returns to normal, the check frequency returns to its original value.

##### Constant
Pinging interval will remain constant. i.e. will not provide any back-off after failure.

##### Linear
The pinging interval upon failure will be extended linearly. i.e. `failure count x pinging interval`.

##### Exponential
With each failure, the subsequent check will be delayed by the last delayed amount, times a multiplier, resulting in time between checks exponentially increasing. The `multiplier` is set to 2 by default.

#### Note for Gmail
If there are errors sending email via gmail - enable `Access for less secure apps` under Account permissions @ https://www.google.com/settings/u/2/security

### Deployment

#### CloudFormation Stacks

See [redalert-cloudformation](https://github.com/jonog/redalert-cloudformation)

##### EC2 & ELB
[![Launch Stack](https://cdn.rawgit.com/buildkite/cloudformation-launch-stack-button-svg/master/launch-stack.svg)](https://console.aws.amazon.com/cloudformation/home#/stacks/new?stackName=redalert&templateURL=https://s3-ap-southeast-2.amazonaws.com/redalert-cloudformation/redalert.yml)

##### EC2 & ELB & S3 config
[![Launch Stack](https://cdn.rawgit.com/buildkite/cloudformation-launch-stack-button-svg/master/launch-stack.svg)](https://console.aws.amazon.com/cloudformation/home#/stacks/new?stackName=redalert&templateURL=https://s3-ap-southeast-2.amazonaws.com/redalert-cloudformation/redalert.yml)


### Development

See [AGENTS.md](AGENTS.md) for the package map, toolchain, generated files, and completion criteria.

#### Setup
Use Go 1.27.1 (the module baseline is Go 1.27.0). Dependencies are managed with Go modules; no GOPATH layout, Glide installation, or checked-in vendor tree is required.

The module requirements use versions resolved from the Glide-locked revisions (including canonical tags where those exact revisions are tagged), with the remaining transitive requirements recorded in `go.mod` and `go.sum`. `glide.yaml` and `glide.lock` are historical inputs only and are not used by the supported build or test workflow.

```sh
go mod download
go mod tidy
go build ./...
go build -ldflags "-X main.version=0.2.4 -X main.commit=$(git rev-parse HEAD)" -o redalert .
./redalert version
```

The checked-in files in `web/assets` are embedded by the standard library at build time. Ordinary Go builds need no Node installation. To regenerate them after changing the UI, use Node.js 20.19.0 and npm 10.8.2 (`nvm use` in `ui/`), then run `make embed-static`; the UI build uses `npm ci` and its checked-in lockfile.

#### Local dashboard

Run `make dev` from the repository. It always rebuilds the current checkout, uses a temporary copy of `config/demo.json`, starts a loopback-only local fixture, and cleans up both processes on Ctrl-C, termination, startup failure, or application exit. It does not read or write your `config.json`. Open the dashboard URL printed by the command. Open the printed fixture URL with `/fail` to trigger a local stderr alert, then `/recover` to see recovery. No Docker, credentials, or external monitoring targets are needed. The browser libraries are still loaded from CDNs, so the dashboard may need network access to render fully.

Override the dashboard, RPC, and fixture ports with `REDALERT_DEV_PORT`, `REDALERT_DEV_RPC_PORT`, and `REDALERT_DEV_FIXTURE_PORT`, respectively:

```sh
REDALERT_DEV_PORT=8890 REDALERT_DEV_RPC_PORT=8891 REDALERT_DEV_FIXTURE_PORT=8892 make dev
```

`make check` is the common local and CI verification command. It checks formatting without modifying files, runs unit tests, builds all packages and the versioned executable, and runs the HTTP smoke check against that executable. Docker-backed integration checks remain separate under `make test-integration`.

#### Tests

Unit tests (default vet enabled):

```sh
go test ./...
```

Docker-backed SSH, Postgres, and Docker stats integration tests are opt-in. They need an x86_64 Docker daemon plus the digest-pinned `sickp/alpine-sshd` fixture and `postgres:9.5`; `make test-deps` pulls them. Set `DOCKER_API_VERSION=1.24` and `DOCKER_HOST` to the Docker API endpoint. The address used to reach published fixture ports is selected separately: set `DOCKER_TEST_HOST` when it differs from the endpoint host. For example, on macOS with Docker Desktop's Unix socket, use `DOCKER_HOST=unix:///var/run/docker.sock DOCKER_TEST_HOST=127.0.0.1`; for a remote daemon, set `DOCKER_TEST_HOST` to a host reachable from the test process (often `host.docker.internal` for Docker Desktop). Unix socket endpoints default to `127.0.0.1`. The PostgreSQL fixture runs with test-only trust authentication. Run the checks with:

```sh
make test-deps
REDALERT_INTEGRATION=1 go test -v ./...
make test-integration
```

The fixture-enabled command sets `REDALERT_INTEGRATION=1` and fails normally if Docker, the fixtures, or the tested behavior fails. `make test-integration` runs the Docker-backed cases in isolation with the expected fixture settings. Use `go test -v ./...` to see named opt-in cases and their skip reasons during the default test run. The remote-command integration cases currently construct a `command` check, so they do not verify SSH behavior; SSH helper correction remains separate work. Run `python3 scripts/smoke.py ./redalert` after `make build` to verify startup outside the checkout, embedded assets and content types, health, statistics, local-target failure and recovery alerts, and the enable/disable/trigger routes.

Race detection is intentionally excluded from this migration because known polling races are tracked separately.

### Credits
Rocket emoji via https://github.com/twitter/twemoji

### Next Features
See Github Issues [here](https://github.com/jonog/redalert/issues)
