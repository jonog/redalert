package core

import "strings"

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
			err = alert.Notify(event)
			if err != nil {
				c.Log.Println(red, "CRITICAL: Failure triggering alert ["+alert.Name()+"]: ", err.Error())
			}
		}

	}()
}

func (c *Check) RecentMetrics(metric string) string {
	var output []string
	for e := c.EventHistory.Front(); e != nil; e = e.Next() {
		event := e.Value.(*Event)
		if event != nil {
			metricStr := event.DisplayMetric(metric)
			output = append([]string{metricStr}, output...)
		}
	}
	return strings.Join(output, ",")
}
