## Redalert

[![Circle CI](https://circleci.com/gh/jonog/redalert.svg?style=svg)](https://circleci.com/gh/jonog/redalert)

For monitoring your infrastructure and sending notifications if stuff is not ok.
(e.g. pinging your websites/APIs via HTTP GET at specified intervals, and alerting you if there is downtime).

![](https://cloud.githubusercontent.com/assets/1314353/7707829/7e18fe10-fe84-11e4-9762-322544d1142b.png)

### Features

#### Checks
* *Website monitoring* & latency measurement (check type: `web-ping`)
* *Server metrics* from local machine (check type: `scollector`)
* *Docker container metrics* from remote host (check type: `remote-docker`)
* *Postgres counts/stats* via SQL queries (check type: `postgres`)
* *TCP connectivity monitoring* & latency measurement (check type: `tcp`)
* *Execute local commands* & capture output (check type: `command`)

#### Dashboard and Alerts
* Alert notifications available on several channels:
  * sending email (`gmail`)
  * sending SMS (`twilio`)
  * posting a message to Slack (`slack`)
  * unix stream (`stderr`)
* Provides ping status & latency info to `stdout`.
* Adjustable back-off after a check fails (constant, linear, exponential - see notes below).
* Includes a web UI as indicated by the screenshot above. (visit localhost:8888/, configure port via env RA_PORT)
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
* Assert on response (using MIME types)
  * source: `text/plain`

#### API
* Event stats available via `/v1/stats`

#### Screenshots
![](https://cloud.githubusercontent.com/assets/1314353/5157264/edb21476-733a-11e4-8452-4b96b443f7ee.jpg)

### Getting started
Run via Docker:
```
docker run -d -P -v /path/to/config.json:/config.json jonog/redalert
```

#### General Configuration
Configure using environment variables
```
RA_PORT=3000 (defaults to 8888)
RA_DISABLE_BRAND=true (defaults to false)
```

#### Monitoring Configuration
Configure servers to monitor & alert settings via `config.json`.

##### Simple config.json
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
                    "source": "text/plain",
                    "target": "400 Bad Request"
                }
            ]
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
    ]
}

```

Build and run (capture stderr).
```
go build

./redalert 2> errors.log
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

### Development

#### Setup
Getting started:
```
go get github.com/tools/godep
```

Embedding static web files:
```
go get github.com/GeertJohan/go.rice
go get github.com/GeertJohan/go.rice/rice
cd web && rice embed-go && cd ..
```

#### Dockerizing Redalert
```
docker run --rm \
  -v "$(pwd):/src" \
  -v /var/run/docker.sock:/var/run/docker.sock \
  centurylink/golang-builder
```

### Credits
Rocket emoji via https://github.com/twitter/twemoji

### TODO / Roadmap
 - [ ] Assert on JSON response
 - [ ] Build out stats API & document endpoints (i.e. `/v1/stats`)
 - [ ] Alerts based on calculated values
 - [ ] Add more checks (expvars, remote command, consul)
 - [ ] Add more notifiers (webhooks, msgqueue)
 - [ ] Push events into a time-series DB (e.g. influx, elasticsearch)
 - [ ] Safely handle concurrent read/writes in key data structures accessed in different goroutines.
 - [ ] Tags/searchable to handle more metrics/ categorising dashboards
 - [ ] More charts
