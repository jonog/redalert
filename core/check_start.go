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

			// Trigger RedAlert if check fails
			if err != nil {

				event = NewRedAlert(c, checkData)
				c.Store.Store(event)

				c.Log.Println(red, "ERROR:", err, reset)
				c.triggerAlerts(event)

				// increase fail count and delay between checks
				c.incrFailCount()
				if c.failCount > 0 {
					delay = time.Second * time.Duration(c.failCount*c.Interval)
				}

			}

			// Trigger GreenAlert if check is successful and was previously failing
			if err == nil {

				lastEvent, storeErr := c.Store.Last()
				if storeErr != nil {
					c.Log.Println(red, "ERROR: retrieving event from store", reset)
				}

				isRedalertRecovery := lastEvent != nil && lastEvent.isRedAlert()

				event = NewGreenAlert(c, checkData)
				c.Store.Store(event)

				if isRedalertRecovery {
					c.Log.Println(green, "RECOVERY: ", reset)
					c.triggerAlerts(event)
				}

				// reset fail count & delay between checks
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
