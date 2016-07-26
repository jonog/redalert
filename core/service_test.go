package core

import (
	"log"
	"testing"

	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/data"
	"github.com/jonog/redalert/notifiers"
	"github.com/jonog/redalert/storage"
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
	}, storage.NewMemoryList(100))
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
	if service.checks[check.Data.ID].Data.Name != "myservice" && service.checks[check.Data.ID].Data.Type != "fake" {
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
	}, storage.NewMemoryList(100))
	if err != nil || check == nil {
		t.Fail()
	}
	err = service.RegisterCheck(check, []string{"my_notifier"}, 0)
	if err != nil {
		t.Fail()
	}
	if len(service.checks[check.Data.ID].Notifiers) != 1 {
		t.Fail()
	}
	if service.checks[check.Data.ID].Notifiers[0].Name() != "my_notifier" {
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

func (c *fakeChecker) Check() (data.CheckResponse, error) {
	return data.CheckResponse{Metrics: data.Metrics(make(map[string]*float64))}, nil
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
