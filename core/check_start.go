package core

import (
	"os"
	"os/signal"
	"sync"
	"time"
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
		var event *Event
		var checkData map[string]float64

		originalDelay := time.Second * time.Duration(c.Interval)
		delay := time.Second * time.Duration(c.Interval)

		for {

			checkData, err = c.Checker.Check()

			if err != nil {

				c.Log.Println(red, "ERROR: ", err, reset)

				// before sending an alert, pause 5 seconds & retry
				// prevent alerts from occaisional errors ('no such host' / 'i/o timeout') on cloud providers
				// todo: adjust sleep to fit with interval
				time.Sleep(5 * time.Second)
				checkData, reCheckErr := c.Checker.Check()
				if reCheckErr != nil {

					// re-check fails (confirms error)

					event = NewRedAlert(c, checkData)
					c.storeEvent(event)
					c.triggerAlerts(event)

					c.incrFailCount()
					if c.failCount > 0 {
						delay = time.Second * time.Duration(c.failCount*c.Interval)
					}

				} else {

					// re-check succeeds (likely false positive)

					delay = originalDelay
					c.resetFailCount()
				}

			} else {

				isRedalertRecovery := c.LastEvent != nil && c.LastEvent.isRedAlert()
				event = NewGreenAlert(c, checkData)
				c.storeEvent(event)
				if isRedalertRecovery {
					c.Log.Println(green, "RECOVERY: ", reset, c.Name)
					c.triggerAlerts(event)
				}

				delay = originalDelay
				c.resetFailCount()

			}

			select {
			case <-time.After(delay):
			case <-stopChan:
				return
			}
		}
	}()

}
