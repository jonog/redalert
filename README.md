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
* Adjustable back-off after failed pings (constant, linear, exponential - see notes below).
* Includes a web UI as indicated by the screenshot above. (visit localhost:8888/, configure port via env RA_PORT)
* Triggers a failure alert (`redalert`) when a check is failing, and a recovery alert (`greenalert`) when the check has recovered (e.g. a successful ping, following a failing ping).
* Triggers an alert when specified metric is above/below threshold.

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
         }
      }
   ],
   "notifications": []
}
```

##### Example Larger config.json
```
{
   "checks":[
      {
         "name":"Server 1",
         "type": "web-ping",
         "config": {
            "address":"http://server1.com/healthcheck",
            "headers": {
              "X-Api-Key": "ABCD1234"
            }
         },
         "send_alerts": ["stderr"],
         "backoff": {
            "type": "constant",
            "interval": 10
         },
         "triggers": [
            {
               "metric": "latency",
               "criteria": ">100"
            }
         ]
      },
      {
         "name":"Server 2",
         "type": "web-ping",
         "config": {
            "address":"http://server2.com/healthcheck"
         },
         "send_alerts": ["stderr", "email", "chat", "sms"],
         "backoff": {
            "type": "linear",
            "interval": 10
         }
      },
      {
         "name":"Server 3",
         "type": "web-ping",
         "config": {
            "address":"http://server3.com/healthcheck"
         },
         "send_alerts": ["stderr"],
         "backoff": {
            "type": "exponential",
            "interval": 10,
            "multiplier": 2
         }
      },
      {
         "name":"Docker Redis",
         "type": "tcp",
         "config": {
            "host":"192.168.99.100",
            "port": 1001
         },
         "send_alerts": ["stderr"],
         "backoff": {
            "type": "constant",
            "interval": 10
         }
      },
      {
         "name": "production-docker-host",
         "type": "remote-docker",
         "config": {
            "host": "ec2-xx-xxx-xx-xxx.ap-southeast-1.compute.amazonaws.com",
            "user": "ubuntu"
         },
         "send_alerts": ["stderr"],
         "backoff": {
            "type": "linear",
            "interval": 5
         }
      },
      {
         "name": "scollector-metrics",
         "type": "scollector",
         "config": {
            "host": "hostname"
         },
         "send_alerts": ["stderr"],
         "backoff": {
            "type": "constant",
            "interval": 15
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
         "send_alerts": ["stderr"],
         "backoff": {
            "type": "linear",
            "interval": 120
         }
      },
      {
         "name":"README size",
         "type": "command",
         "config": {
            "command":"cat README.md | wc -l",
            "output_type": "number"
         },
         "send_alerts": ["stderr"],
         "backoff": {
            "type": "constant",
            "interval": 10
         }
      },
      {
         "name":"List files",
         "type": "command",
         "config": {
            "command":"ls"
         },
         "send_alerts": ["stderr"],
         "backoff": {
            "type": "constant",
            "interval": 10
         }
      }
   ],
   "notifications": [
      {
         "name": "email",
         "type": "gmail",
         "config": {
            "user": "",
            "pass": "",
            "notification_addresses": ""
         }
      },
      {
         "name": "chat",
         "type": "slack",
         "config": {
            "webhook_url": "",
            "channel": "#general",
            "username": "redalert",
            "icon_emoji": ":rocket:"
         }
      },
      {
         "name": "sms",
         "type": "twilio",
         "config": {
            "account_sid": "",
            "auth_token": "",
            "twilio_number": "",
            "notification_numbers": ""
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
 - [ ] Build out stats API & document endpoints (i.e. `/v1/stats`)
 - [ ] Alerts based on calculated values
 - [ ] Add more checks (expvars, remote command, consul)
 - [ ] Add more notifiers (webhooks, msgqueue)
 - [ ] Push events into a time-series DB (e.g. influx, elasticsearch)
 - [ ] Distinguish between an error performing a check & a failing check
 - [ ] Safely handle concurrent read/writes in key data structures accessed in different goroutines.
 - [ ] Tags/searchable to handle more metrics/ categorising dashboards
 - [ ] More charts
