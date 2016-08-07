package core

import (
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/jonog/redalert/assertions"
	"github.com/jonog/redalert/events"
	"github.com/jonog/redalert/servicepb"
	"github.com/jonog/redalert/utils"
)

func (c *Check) Start() {
	c.Data.Enabled = true
	c.Data.Status = servicepb.Check_UNKNOWN

	c.wait.Add(1)

	serviceStop := make(chan bool)
	c.run(serviceStop)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		for range sigChan {
			serviceStop <- true
		}
	}()

	c.wait.Wait()

}

func (c *Check) Stop() {
	c.Data.Enabled = false
	c.Data.Status = servicepb.Check_DISABLED
	c.stopChan <- true
}

func (c *Check) cleanup() {
	c.wait.Done()
}

func (c *Check) handleFailing(ev *events.Event, failMessages []string) int {
	c.Data.Status = servicepb.Check_FAILING
	failCount := c.Counter.Inc("failing_seq", 1)
	c.Counter.Inc("failing_all", 1)
	c.Counter.Reset("successful_seq")
	c.Tracker.Track("last_failed_at")

	ev.MarkRedAlert(failMessages)
	c.Log.Println(utils.Red, "redalert", failMessages, utils.Reset)

	return failCount
}

func (c *Check) handleSuccessful() {
	c.Data.Status = servicepb.Check_SUCCESSFUL
	c.Counter.Inc("successful_seq", 1)
	c.Counter.Inc("successful_all", 1)
	c.Counter.Reset("failing_seq")
	c.Tracker.Track("last_successful_at")
}

func (c *Check) handleRecovery(ev *events.Event) {
	ev.MarkGreenAlert()
	c.Log.Println(utils.Green, "greenalert", utils.Reset)
	c.Counter.Reset("failing")
}

func (c *Check) run(serviceStop chan bool) {

	go func() {

		delay := c.Backoff.Init()

		for {

			checkResponse, err := c.Checker.Check()
			event := events.NewEvent(checkResponse)
			prevState := c.Data.Status

			fail, failMessages := c.isFailing(err, event)
			if fail {
				failCount := c.handleFailing(event, failMessages)
				if failCount > 0 {
					delay = c.Backoff.Next(failCount)
				}
			} else {
				c.handleSuccessful()
				if prevState == servicepb.Check_FAILING {
					c.handleRecovery(event)
					delay = c.Backoff.Init()
				}
			}

			c.Tracker.Track("last_checked_at")
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

	msgPrefix := c.Data.Name + " :: (" + c.Data.Type + " - " + c.Checker.MessageContext() + ") "

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
