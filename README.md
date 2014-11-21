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
         "actions":["console"]
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
RA_SLACK_URL=<webhook_url> ./redalert
```

### TODO
* Setup server info & alerting configuration via config file(s)
* Add more alerting configurations. I.e. add additional types which satisfy the Action interface.
```
type Email struct{}
type SMS struct{}
type ExecuteCommand struct{}
```