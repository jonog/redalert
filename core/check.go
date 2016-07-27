package core

import (
	"errors"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"

	"github.com/jonog/redalert/assertions"
	"github.com/jonog/redalert/backoffs"
	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/notifiers"
	"github.com/jonog/redalert/servicepb"
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

	Counter storage.Counter
	Tracker storage.Tracker

	ConfigRank int

	stopChan chan bool
	wait     sync.WaitGroup
}

func NewCheck(config checks.Config, eventStorage storage.EventStorage) (*Check, error) {
	logger := log.New(os.Stdout, config.Name+" ", log.Ldate|log.Ltime)

	checker, err := checks.New(config, logger)
	if err != nil {
		return nil, err
	}

	asserters := make([]assertions.Asserter, 0)
	for _, assertionConfig := range config.Assertions {
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

	return &Check{
		Data: servicepb.Check{
			ID:      generateID(8),
			Name:    config.Name,
			Type:    config.Type,
			Enabled: config.Enabled == nil || *config.Enabled,
			Status:  initState(config),
		},

		Backoff:    backoffs.New(config.Backoff),
		Notifiers:  make([]notifiers.Notifier, 0),
		Log:        logger,
		Counter:    storage.NewBasicCounter(),
		Tracker:    storage.NewBasicTracker(),
		Store:      eventStorage,
		Checker:    checker,
		Assertions: asserters,

		stopChan: make(chan bool),
	}, nil
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

var idLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func generateID(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = idLetters[rand.Intn(len(idLetters))]
	}
	return string(b)
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
