### Redalert
For monitoring a series of servers at specified intervals & triggering alerts if there is downtime. Currently supports:
* sending email (via gmail)
* sending SMS (via Twilio)
* posting a message to Slack
* messaging on `stderr`

#### Features:
Alert options: ["stderr", "gmail", "slack", "twilio"]
Provides ping status & latency info to `stdout`.
Has a linear back-off after failed pings (see notes below).
Provides a web status UI (visit localhost:8888/, configure port via env RA_PORT)
Provides an alert when a failing server has recovered (i.e. a successful ping, following a failing ping).

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
         "alerts":["stderr"]
      },
      {  
         "name":"Server 2",
         "type": "web-ping",
         "address":"http://server2.com/healthcheck",
         "interval":10,
         "alerts":["stderr", "gmail", "slack", "twilio"]
      },
      {  
         "name":"Server 3",
         "type": "web-ping",
         "address":"http://server3.com/healthcheck",
         "interval":10,
         "alerts":["stderr"]
      }
   ],
   "gmail": {
      "user": "",
      "pass": "",
      "notification_addresses": []
   },
   "slack": {
      "webhook_url": "",
      "channel": "#general",
      "username": "redalert",
      "icon_emoji": ":rocket:"
   },
   "twilio": {
      "account_sid": "",
      "auth_token": "",
      "twilio_number": "",
      "notification_numbers": []
   }

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
* Distinguish between an error performing a check & a failing check. i.e. Check should return two errors.
* Safely handle concurrent read/writes in key data structures accessed in different goroutines.
* Change server config on the fly
* Alternative backoff configurations (e.g. no backoff / exponential backoff after X attempts)
* Improved handling/recovering from any errors detected in redalert
