package core

import (
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/jonog/redalert/assertions"
	"github.com/jonog/redalert/data"
	"github.com/jonog/redalert/events"
	"github.com/jonog/redalert/utils"
)

func (c *Check) Start() {
	c.Enabled = true

	c.wait.Add(1)

	serviceStop := make(chan bool)
	c.run(serviceStop)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		for _ = range sigChan {
			serviceStop <- true
		}
	}()

	c.wait.Wait()

}

func (c *Check) Stop() {
	c.Enabled = false
	c.stopChan <- true
}

func (c *Check) cleanup() {
	c.wait.Done()
}

func (c *Check) run(serviceStop chan bool) {

	go func() {

		var err error
		var event *events.Event
		var checkResponse data.CheckResponse
		var fail bool
		var failMessages []string

		delay := c.Backoff.Init()

		for {

			checkResponse, err = c.Checker.Check()
			event = events.NewEvent(checkResponse)

			fail, failMessages = c.isFailing(err, event)
			if fail {

				// Trigger RedAlert as check has failed
				event.MarkRedAlert(failMessages)
				c.Log.Println(utils.Red, "redalert", failMessages, utils.Reset)

				// increase fail count and delay between checks
				failCount, storeErr := c.Store.IncrFailCount("redalert")
				if storeErr != nil {
					c.Log.Println(utils.Red, "ERROR: storing failure stats", utils.Reset)
				}
				if failCount > 0 {
					delay = c.Backoff.Next(failCount)
				}

			} else {

				lastEvent, storeErr := c.Store.Last()
				if storeErr != nil {
					c.Log.Println(utils.Red, "ERROR: retrieving event from store", utils.Reset)
				}

				// Trigger GreenAlert if check is successful and was previously failing
				isRedalertRecovery := lastEvent != nil && lastEvent.IsRedAlert()
				if isRedalertRecovery {
					event.MarkGreenAlert()
					c.Log.Println(utils.Green, "greenalert", utils.Reset)

					// reset fail count & delay between checks
					delay = c.Backoff.Init()
					storeErr := c.Store.ResetFailCount("redalert")
					if storeErr != nil {
						c.Log.Println(utils.Red, "ERROR: storing failure stats", utils.Reset)
					}

				}

			}

			c.Store.Store(event)
			c.processNotifications(event)

			select {
			case <-time.After(delay):
			case <-c.stopChan:
				c.cleanup()
				return
			case <-serviceStop:
				c.cleanup()
				return
			}
		}
	}()

}

func (c *Check) isFailing(err error, event *events.Event) (bool, []string) {
	messages := []string{}
	if err != nil {
		messages = append(messages, "check failure: "+err.Error())
		return true, messages
	}
	for _, assertion := range c.Assertions {
		// TODO: consider returning a message here too
		outcome, err := assertion.Assert(assertions.Options{CheckResponse: event.Data})
		if err != nil {
			messages = append(messages, "assertion failure: "+err.Error())
		} else if !outcome.Assertion {
			messages = append(messages, outcome.Message)
		}
	}
	return len(messages) > 0, messages
}

func (c *Check) processNotifications(event *events.Event) {

	msgPrefix := c.Name + " :: (" + c.Type + " - " + c.Checker.MessageContext() + ") "

	// Process Redalert/Greenalert (Failure / Recovery)

	if len(event.Tags) == 0 {
		return
	}

	go func() {

		if !event.IsRedAlert() && !event.IsGreenAlert() {
			return
		}

		var err error
		for _, notifier := range c.Notifiers {

			c.Log.Println(utils.White, "Sending "+event.DisplayTags()+" via "+notifier.Name(), utils.Reset)

			var msg string
			if event.IsRedAlert() {
				msg = msgPrefix + "fail: " + strings.Join(event.Messages, ",")
			} else if event.IsGreenAlert() {
				msg = msgPrefix + "recovery"
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
