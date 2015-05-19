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
* Has a linear back-off after failed pings (see notes below).
* Includes a web UI as indicated by the screenshot above. (visit localhost:8888/, configure port via env RA_PORT)
* Triggers a failure alert (`redalert`) when a check is failing, and a recovery alert (`greenalert`) when the check has recovered (e.g. a successful ping, following a failing ping).

#### Coming soon:
* Server metrics
* Metric threshold alerting (i.e. metric beyond threshold / outside range)

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
         "interval":10,
         "send_alerts": ["stderr"]
      },
      {  
         "name":"Server 2",
         "type": "web-ping",
         "address":"http://server2.com/healthcheck",
         "interval":10,
         "send_alerts": ["stderr", "email", "chat", "sms"]
      },
      {  
         "name":"Server 3",
         "type": "web-ping",
         "address":"http://server3.com/healthcheck",
         "interval":10,
         "send_alerts": ["stderr"]
      },
      {
         "name": "scollector-metrics",
         "type": "scollector",
         "host": "hostname",
         "interval": 15,
         "send_alerts": ["stderr"]
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


#### Linear back-off after failure
The pinging interval will be adjusted to X * pinging interval where X is the number of times the pinger has failed. E.g. after 1 failure, the pinging interval will not be changed, but a server (with pinging interval of 10s) which has failed ping once will only be pinged after 20s, then 30s, 40s etc.
When a failing server is successfully pinged, the pinging frequency returns to the originally configured value.

#### Note for Gmail:
If there are errors sending email via gmail - enable `Access for less secure apps` under Account permissions @ https://www.google.com/settings/u/2/security

#### Credits:
Rocket emoji via https://github.com/twitter/twemoji

### TODO
* Set alerts based on metric threshold values / calculated values
* Integrate more checks (db query, expvars, remote command via ssh, consul)
* Integrate more notifiers (webhooks, msgqueue)
* Push events to a time-series database
* Distinguish between an error performing a check & a failing check. i.e. Check should return two errors.
* Safely handle concurrent read/writes in key data structures accessed in different goroutines.
* Change server config on the fly / interface out storage of config.
* Alternative backoff configurations (e.g. no backoff / exponential backoff after X attempts)
* Improved handling/recovering from any errors detected in redalert

### TODO UI
* Tags/searchable to handle more metrics/ categorising dashboards
* Live updates via websockets/poll
* More charts
