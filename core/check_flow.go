package core

import (
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/jonog/redalert/events"
	"github.com/jonog/redalert/utils"
)

func (c *Check) Start() {

	c.service.wg.Add(1)

	var wg sync.WaitGroup
	wg.Add(1)

	stopScheduler := make(chan bool)
	c.run(stopScheduler)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		for _ = range sigChan {
			stopScheduler <- true
			wg.Done()
		}
	}()

	wg.Wait()
	c.service.wg.Done()

}

func (c *Check) run(stopChan chan bool) {

	go func() {

		var err error
		var event *events.Event
		var checkData map[string]float64

		originalDelay := c.Backoff.Init()
		delay := c.Backoff.Init()

		for {

			checkData, err = c.Checker.Check()
			event = events.NewEvent(checkData)

			if err != nil {

				// Trigger RedAlert as check has failed
				event.SetType("redalert")
				c.Log.Println(utils.Red, "redalert", err, utils.Reset)

				// increase fail count and delay between checks
				c.incrFailCount()
				if c.failCount > 0 {
					delay = c.Backoff.Next(c.failCount)
				}

			}

			if err == nil {

				lastEvent, storeErr := c.Store.Last()
				if storeErr != nil {
					c.Log.Println(utils.Red, "ERROR: retrieving event from store", utils.Reset)
				}

				// Trigger GreenAlert if check is successful and was previously failing
				isRedalertRecovery := lastEvent != nil && lastEvent.IsRedAlert()
				if isRedalertRecovery {
					event.SetType("greenalert")
					c.Log.Println(utils.Green, "greenalert", utils.Reset)
				}

				// reset fail count & delay between checks
				delay = originalDelay
				c.resetFailCount()

			}

			c.Store.Store(event)
			c.processNotifications(event)

			select {
			case <-time.After(delay):
			case <-stopChan:
				return
			}
		}
	}()

}

func (c *Check) processNotifications(event *events.Event) {

	if event.Type == "" {
		return
	}

	// TODO:
	// threshold notifications

	go func() {

		var err error
		for _, notifier := range c.Notifiers {

			c.Log.Println(utils.White, "Sending "+event.Type+" via "+notifier.Name(), utils.Reset)

			var msg string
			switch event.Type {
			case "redalert":
				msg = c.Checker.RedAlertMessage()
			case "greenalert":
				msg = c.Checker.GreenAlertMessage()
			}

			err = notifier.Notify(AlertMessage{msg})
			if err != nil {
				c.Log.Println(utils.Red, "CRITICAL: Failure triggering alert ["+notifier.Name()+"]: ", err.Error())
			}
		}

	}()
}

type AlertMessage struct {
	Short string
}

func (m AlertMessage) ShortMessage() string {
	return m.Short
}
