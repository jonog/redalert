### Redalert
For monitoring a series of servers at specified intervals & triggering actions if there is downtime (e.g. send email, post webhook).

### TODO
* Setup server info & alerting configuration via config file(s)
* Add more alerting configurations. I.e. add additional types which satisfy the Action interface.
```
type Email struct{}
type SMS struct{}
type SlackWebhook struct{}
type ExecuteCommand struct{}
```