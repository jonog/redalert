### Redalert

For monitoring your infrastructure and sending notifications if stuff is not ok.
(e.g. pinging your websites/APIs via HTTP GET at specified intervals, and alerting you if there is downtime).

![](https://cloud.githubusercontent.com/assets/1314353/7707829/7e18fe10-fe84-11e4-9762-322544d1142b.png)

#### Features:
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

#### Coming soon:
* Server metrics

#### Screenshots:
![](https://cloud.githubusercontent.com/assets/1314353/5157264/edb21476-733a-11e4-8452-4b96b443f7ee.jpg)

#### Getting started:
Configure servers to monitor & alert settings via `config.json`:
```
{  
   "checks":[  
      {  
         "name":"Server 1",
         "type": "web-ping",
         "address":"http://server1.com/healthcheck",
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
         "address":"http://server2.com/healthcheck",
         "send_alerts": ["stderr", "email", "chat", "sms"],
         "backoff": {
            "type": "linear",
            "interval": 10
         }
      },
      {  
         "name":"Server 3",
         "type": "web-ping",
         "address":"http://server3.com/healthcheck",
         "send_alerts": ["stderr"],
         "backoff": {
            "type": "exponential",
            "interval": 10,
            "multiplier": 2
         }
      },
      {
         "name": "scollector-metrics",
         "type": "scollector",
         "host": "hostname",
         "send_alerts": ["stderr"],
         "backoff": {
            "type": "constant",
            "interval": 15
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

#### Note for Gmail:
If there are errors sending email via gmail - enable `Access for less secure apps` under Account permissions @ https://www.google.com/settings/u/2/security

#### Credits:
Rocket emoji via https://github.com/twitter/twemoji

### TODO
* Set alerts based on metric threshold values for N intervals / based on calculation
* Integrate more checks (db query, expvars, remote command via ssh, consul)
* Integrate more notifiers (webhooks, msgqueue)
* Push events to a time-series database
* Distinguish between an error performing a check & a failing check. i.e. Check should return two errors.
* Safely handle concurrent read/writes in key data structures accessed in different goroutines.
* Change server config on the fly / interface out configuration details.
* Improved handling/recovering from any errors detected in redalert

### TODO UI
* Tags/searchable to handle more metrics/ categorising dashboards
* Live updates via websockets/poll
* More charts
