package core

var MaxEventsStored = 100

func (c *Check) storeEvent(event *Event) {
	c.LastEvent = event
	c.EventHistory.PushFront(event)
	if c.EventHistory.Len() > MaxEventsStored {
		c.EventHistory.Remove(c.EventHistory.Back())
	}
}

func (c *Check) triggerAlerts(event *Event) {

	go func() {

		var err error
		for _, alert := range c.Alerts {
			err = alert.Trigger(event)
			if err != nil {
				c.Log.Println(red, "CRITICAL: Failure triggering alert ["+alert.Name()+"]: ", err.Error())
			}
		}

	}()
}
