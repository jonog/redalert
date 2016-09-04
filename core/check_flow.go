package core

import (
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/jonog/redalert/assertions"
	"github.com/jonog/redalert/events"
	"github.com/jonog/redalert/notifiers"
	"github.com/jonog/redalert/servicepb"
	"github.com/jonog/redalert/utils"
)

func (c *Check) Start() {
	c.Data.Enabled = true
	c.Data.Status = servicepb.Check_UNKNOWN
	c.Stats.StateTransitionedAt.Mark()

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
	c.Stats.StateTransitionedAt.Mark()
	c.stopChan <- true
}

func (c *Check) cleanup() {
	c.wait.Done()
}

func (c *Check) handleFailing(ev *events.Event, failMessages []string) int {
	if c.Data.Status != servicepb.Check_FAILING {
		c.Stats.StateTransitionedAt.Mark()
	}
	c.Data.Status = servicepb.Check_FAILING
	failCount := c.Stats.FailureSequence.Inc()
	c.Stats.FailureTotal.Inc()
	c.Stats.SuccessfulSequence.Reset()
	c.Stats.LastFailedAt.Mark()

	ev.MarkRedAlert(failMessages)
	c.Log.Println(utils.Red, "redalert", failMessages, utils.Reset)

	return failCount
}

func (c *Check) handleSuccessful() {
	if c.Data.Status != servicepb.Check_SUCCESSFUL {
		c.Stats.StateTransitionedAt.Mark()
	}
	c.Data.Status = servicepb.Check_SUCCESSFUL
	c.Stats.SuccessfulSequence.Inc()
	c.Stats.SuccessfulTotal.Inc()
	c.Stats.FailureSequence.Reset()
	c.Stats.CurrentAlertCount.Reset()
	c.Stats.LastSuccessfulAt.Mark()
}

func (c *Check) handleRecovery(ev *events.Event) {
	ev.MarkGreenAlert()
	c.Log.Println(utils.Green, "greenalert", utils.Reset)
}

func (c *Check) run(serviceStop chan bool) {

	// add some jitter
	time.Sleep(time.Duration(randInt(1, 2500)) * time.Millisecond)
	time.Sleep(time.Duration(randInt(1, 2500)) * time.Millisecond)

	go func() {

		delay := c.Backoff.Init()

		for {

			checkResponse, err := c.Checker.Check()
			event := events.NewEvent(c.Data.ID, c.Data.Name, c.Data.Type, checkResponse)
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

			c.Stats.LastCheckedAt.Mark()
			c.Store.Store(event)

			if c.hasNotifications(event) {
				c.processNotifications(event)
			}

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

func (c *Check) hasNotifications(event *events.Event) bool {
	if len(event.Tags) == 0 {
		return false
	}
	if !event.IsRedAlert() && !event.IsGreenAlert() {
		return false
	}
	// don't send alerts if the number of consequitive failures is below the threshold
	if event.IsRedAlert() && c.Stats.FailureSequence.Get() < c.FailCountAlertThreshold {
		return false
	}
	// don't send alerts if already have been sent
	if event.IsRedAlert() && c.Stats.CurrentAlertCount.Get() > 0 && !c.RepeatFailAlerts {
		return false
	}
	return true
}

func (c *Check) processNotifications(event *events.Event) {
	c.Stats.CurrentAlertCount.Inc()
	go func() {
		for _, notifier := range c.Notifiers {
			go func(n notifiers.Notifier) {
				c.Log.Println(utils.White, "Sending "+event.DisplayTags()+" via "+n.Name(), utils.Reset)
				err := n.Notify(notifiers.Message{
					DefaultMessage: c.message(event),
					Event:          event,
				})
				if err != nil {
					c.Log.Println(utils.Red, "CRITICAL: Failure triggering alert ["+n.Name()+"]: ", err.Error())
				}
			}(notifier)
		}
	}()
}

func (c *Check) message(event *events.Event) string {
	msgPrefix := c.Data.Name + " :: (" + c.Data.Type + " - " + c.Checker.MessageContext() + ") "
	var msg string
	if event.IsRedAlert() {
		msg = msgPrefix + "fail: " + strings.Join(event.Messages, ",")
	} else if event.IsGreenAlert() {
		msg = msgPrefix + "recovery - check is now successful"
	}
	return msg
}

func randInt(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}
