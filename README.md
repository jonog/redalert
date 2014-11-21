### Redalert
For monitoring a series of servers at specified intervals & triggering actions if there is downtime (e.g. send email, post webhook).

#### Getting started:
Configure servers to monitor via `servers.json`:
```
{
   "servers":[
      {
         "name":"Server 1",
         "address":"http://server1.com/healthcheck",
         "interval":3,
         "actions":["console", "email"]
      },
      {
         "name":"Server 2",
         "address":"http://server2.com/healthcheck",
         "interval":3,
         "actions":["console", "slack"]
      },
      {
         "name":"Server 1",
         "address":"http://server3.com/healthcheck",
         "interval":3,
         "actions":["console"]
      }
   ]
}
```

Build and run with env variables set for configuring actions.
```
go build

RA_GMAIL_USER=<insert> \
RA_GMAIL_PASS=<insert> \
RA_GMAIL_NOTIFICATION_ADDRESS=<insert> \
RA_SLACK_URL=<insert> \
RA_TWILIO_ACCOUNT_SID=<insert> \
RA_TWILIO_AUTH_TOKEN=<insert> \
RA_TWILIO_PHONE_NUMBER=<insert> \
RA_TWILIO_TWILIO_NUMBER=<insert> \
./redalert
```

### Note for Gmail:
If there are errors sending email via gmail - enable `Access for less secure apps` under Account permissions @ https://www.google.com/settings/u/2/security

### TODO
* Setup server info & alerting configuration via config file(s)
* Add more alerting configurations. I.e. add additional types which satisfy the Action interface.
```
type ExecuteCommand struct{}
```