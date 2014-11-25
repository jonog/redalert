### Redalert
For monitoring a series of servers at specified intervals & triggering alerts if there is downtime. Currently supports:
* sending email (via gmail)
* sending SMS (via Twilio)
* posting a message to Slack
* messaging on `stderr`

Current alert options: ["stderr", "gmail", "slack", "twilio"]

![](https://cloud.githubusercontent.com/assets/1314353/5157264/edb21476-733a-11e4-8452-4b96b443f7ee.jpg)

#### Getting started:
Configure servers to monitor & alert settings via `config.json`:
```
{  
   "servers":[  
      {  
         "name":"Server 1",
         "address":"http://server1.com/healthcheck",
         "interval":10,
         "alerts":["stderr"]
      },
      {  
         "name":"Server 2",
         "address":"http://server2.com/healthcheck",
         "interval":10,
         "alerts":["stderr", "gmail", "slack", "twilio"]
      },
      {  
         "name":"Server 3",
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
      "webhook_url": ""
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

### Note for Gmail:
If there are errors sending email via gmail - enable `Access for less secure apps` under Account permissions @ https://www.google.com/settings/u/2/security

### TODO
* Store latency information
* Change server config on the fly
* Exponential backoff after X attempts
