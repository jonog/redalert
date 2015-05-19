package core

func (c *Check) triggerAlerts(event *Event) {

	go func() {

		var err error
		for _, notifier := range c.Notifiers {

			if event.isRedAlert() {
				c.Log.Println(red, "Sending red alert via", notifier.Name(), reset)
			} else if event.isGreenAlert() {
				c.Log.Println(green, "Sending green alert via", notifier.Name(), reset)
			}

			err = notifier.Notify(event)
			if err != nil {
				c.Log.Println(red, "CRITICAL: Failure triggering alert ["+notifier.Name()+"]: ", err.Error())
			}
		}

	}()
}
