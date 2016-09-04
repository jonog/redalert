package core

import (
	"errors"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/jonog/redalert/assertions"
	"github.com/jonog/redalert/backoffs"
	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/config"
	"github.com/jonog/redalert/notifiers"
	"github.com/jonog/redalert/servicepb"
	"github.com/jonog/redalert/stats"
	"github.com/jonog/redalert/storage"
)

type Check struct {
	Data servicepb.Check

	Backoff    backoffs.Backoff
	Notifiers  []notifiers.Notifier
	Log        *log.Logger
	Store      storage.EventStorage
	Checker    checks.Checker
	Assertions []assertions.Asserter

	Stats *stats.CheckStats

	// send an alert only after N fails
	FailCountAlertThreshold int

	// continue to send fail alerts
	RepeatFailAlerts bool

	ConfigRank int

	stopChan chan bool
	wait     sync.WaitGroup
}

func NewCheck(cfg checks.Config, eventStorage storage.EventStorage, preferences config.Preferences) (*Check, error) {

	logger := log.New(os.Stdout, cfg.Name+" ", log.Ldate|log.Ltime)

	checker, err := checks.New(cfg, logger)
	if err != nil {
		return nil, err
	}

	asserters := make([]assertions.Asserter, 0)
	for _, assertionConfig := range cfg.Assertions {
		var err error
		asserter, err := assertions.New(assertionConfig, logger)
		if err != nil {
			return nil, err
		}
		err = asserter.ValidateConfig()
		if err != nil {
			return nil, err
		}
		asserters = append(asserters, asserter)
	}

	initialStats := stats.NewCheckStats()
	initialStats.StateTransitionedAt.Mark()

	return &Check{
		Data: servicepb.Check{
			ID:      cfg.ID,
			Name:    cfg.Name,
			Type:    cfg.Type,
			Enabled: cfg.Enabled == nil || *cfg.Enabled,
			Status:  initState(cfg),
		},

		Backoff:    backoffs.New(cfg.Backoff),
		Notifiers:  make([]notifiers.Notifier, 0),
		Log:        logger,
		Stats:      initialStats,
		Store:      eventStorage,
		Checker:    checker,
		Assertions: asserters,

		FailCountAlertThreshold: intPtrDefault(
			preferences.Notifications.FailCountAlertThreshold,
			config.DefaultFailCountAlertThreshold),
		RepeatFailAlerts: boolPtrDefault(
			preferences.Notifications.RepeatFailAlerts,
			config.DefaultRepeatFailAlerts),

		stopChan: make(chan bool),
	}, nil
}

func intPtrDefault(ptr *int, def int) int {
	if ptr == nil {
		return def
	}
	return *ptr
}

func boolPtrDefault(ptr *bool, def bool) bool {
	if ptr == nil {
		return def
	}
	return *ptr
}

type CheckState int

const (
	Disabled CheckState = iota
	Unknown
	Successful
	Failing
)

func (c *Check) DisplayState() string {
	return c.Data.Status.String()
}

func initState(config checks.Config) servicepb.Check_Status {
	if config.Enabled == nil || *config.Enabled {
		return servicepb.Check_UNKNOWN
	}
	return servicepb.Check_DISABLED
}

func (c *Check) AddNotifiers(service *Service, names []string) error {
	for _, name := range names {
		notifier, err := getNotifier(service, name)
		if err != nil {
			return err
		}
		c.Notifiers = append(c.Notifiers, notifier)
	}
	return nil
}

func getNotifier(service *Service, name string) (notifiers.Notifier, error) {
	notifier, ok := service.notifiers[name]
	if !ok {
		return nil, errors.New("redalert: notifier requested has not be registered. name: " + name)
	}
	return notifier, nil
}

func (c *Check) RecentMetrics(metric string) string {
	events, err := c.Store.GetRecent()
	if err != nil {
		c.Log.Println("ERROR: retrieving recent events")
		return ""
	}
	var output []string
	for _, event := range events {
		output = append([]string{event.DisplayMetric(metric)}, output...)
	}
	return strings.Join(output, ",")
}
