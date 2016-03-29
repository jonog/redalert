package core

import (
	"log"
	"testing"

	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/notifiers"
)

func TestRegisterNotifiers(t *testing.T) {
	service := NewService()
	_, exists := service.notifiers["fake"]
	if exists {
		t.Fail()
	}
	_, exists = service.notifiers["fake2"]
	if exists {
		t.Fail()
	}
	err := service.RegisterNotifier(&fakeNotifier{"fake"})
	if err != nil {
		t.Fail()
	}
	err = service.RegisterNotifier(&fakeNotifier{"fake2"})
	if err != nil {
		t.Fail()
	}
	_, exists = service.notifiers["fake"]
	if !exists {
		t.Fail()
	}
	_, exists = service.notifiers["fake2"]
	if !exists {
		t.Fail()
	}
}

func TestRegisterNotifierDuplicate(t *testing.T) {
	service := NewService()
	err := service.RegisterNotifier(&fakeNotifier{"fake"})
	if err != nil {
		t.Fail()
	}
	err = service.RegisterNotifier(&fakeNotifier{"fake"})
	if err == nil {
		t.Fail()
	}
}

func TestRegisterCheck(t *testing.T) {
	checks.Register("fake", NewFakeChecker)
	check, err := NewCheck(checks.Config{
		Name: "myservice",
		Type: "fake",
	})
	if err != nil || check == nil {
		t.Fail()
	}
	service := NewService()
	if len(service.checks) != 0 {
		t.Fail()
	}
	err = service.RegisterCheck(check, []string{}, 0)
	if err != nil {
		t.Fail()
	}
	if len(service.checks) != 1 {
		t.Fail()
	}
	if service.checks[check.ID].Name != "myservice" && service.checks[check.ID].Type != "fake" {
		t.Fail()
	}
}

func TestRegisterCheckNotifications(t *testing.T) {
	checks.Register("fake", NewFakeChecker)
	service := NewService()
	err := service.RegisterNotifier(&fakeNotifier{"my_notifier"})
	if err != nil {
		t.Fail()
	}
	check, err := NewCheck(checks.Config{
		Name:       "myservice",
		Type:       "fake",
		SendAlerts: []string{"my_notifier"},
	})
	if err != nil || check == nil {
		t.Fail()
	}
	err = service.RegisterCheck(check, []string{"my_notifier"}, 0)
	if err != nil {
		t.Fail()
	}
	if len(service.checks[check.ID].Notifiers) != 1 {
		t.Fail()
	}
	if service.checks[check.ID].Notifiers[0].Name() != "my_notifier" {
		t.Fail()
	}
}

///////////////
// Helpers
///////////////

// Fake Notifier

type fakeNotifier struct {
	name string
}

func (n *fakeNotifier) Notify(msg notifiers.Message) error {
	return nil
}

func (n *fakeNotifier) Name() string {
	return n.name
}

// Fake Checker

type fakeChecker struct {
}

var NewFakeChecker = func(config checks.Config, logger *log.Logger) (checks.Checker, error) {
	return checks.Checker(&fakeChecker{}), nil
}

func (c *fakeChecker) Check() (checks.Metrics, error) {
	return checks.Metrics(make(map[string]*float64)), nil
}

func (c *fakeChecker) MetricInfo(metric string) checks.MetricInfo {
	return checks.MetricInfo{Unit: ""}
}

func (c *fakeChecker) RedAlertMessage() string {
	return "redalert"
}

func (c *fakeChecker) GreenAlertMessage() string {
	return "greenalert"
}

func (c *fakeChecker) MessageContext() string {
	return "message_context"
}
