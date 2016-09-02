package config

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/jonog/redalert/assertions"
	"github.com/jonog/redalert/backoffs"
	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/notifiers"

	_ "github.com/lib/pq"
	"gopkg.in/gorp.v1"
)

type DBStore struct {
	DB *gorp.DbMap
}

type CheckRecord struct {
	ID         string              `db:"id"`
	Name       string              `db:"name"`
	Type       string              `db:"type"`
	SendAlerts []string            `db:"send_alerts"`
	Backoff    backoffs.Config     `db:"backoff"`
	Config     json.RawMessage     `db:"config"`
	Assertions []assertions.Config `db:"assertions"`
}

var checkRecordColumns string = strings.Join([]string{"id", "name", "type", "send_alerts", "backoff", "config", "assertions"}, ",")

func (c *DBStore) findCheckRecord(id string) (*CheckRecord, error) {
	r := new(CheckRecord)
	err := c.DB.SelectOne(r, "select "+checkRecordColumns+" from checks where id=$1", id)
	return r, err
}

func (c *DBStore) createOrUpdateCheck(check checks.Config) error {
	// TODO: improve using upsert
	record := &CheckRecord{
		ID:         check.ID,
		Name:       check.Name,
		Type:       check.Type,
		SendAlerts: check.SendAlerts,
		Backoff:    check.Backoff,
		Config:     check.Config,
		Assertions: check.Assertions,
	}
	err := c.DB.Insert(record)
	if err == nil {
		return nil
	}
	matched, _ := regexp.MatchString("duplicate.*checks_pkey", err.Error())
	if !matched {
		return err
	}
	if err != nil {
		cr, err := c.findCheckRecord(check.ID)
		if err != nil && err == sql.ErrNoRows {
			return errors.New("config: cannot find check. name: " + record.Name)
		}
		if err != nil {
			return err
		}
		cr.Name = record.Name
		cr.Type = record.Type
		cr.SendAlerts = record.SendAlerts
		cr.Backoff = record.Backoff
		cr.Config = record.Config
		cr.Assertions = record.Assertions
		_, err = c.DB.Update(cr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *DBStore) getAllCheckRecords() (checks []CheckRecord, err error) {
	_, err = c.DB.Select(&checks, "select "+checkRecordColumns+" from checks")
	return checks, err
}

type NotificationRecord struct {
	ID     string            `db:"id"`
	Name   string            `db:"name"`
	Type   string            `db:"type"`
	Config map[string]string `db:"config"`
}

var notificationRecordColumns string = strings.Join([]string{"id", "name", "type", "config"}, ",")

func (c *DBStore) findNotificationRecord(id string) (*NotificationRecord, error) {
	r := new(NotificationRecord)
	err := c.DB.SelectOne(r, "select "+notificationRecordColumns+" from notifications where id=$1", id)
	return r, err
}

func (c *DBStore) createOrUpdateNotification(notifier notifiers.Config) error {
	// TODO: improve using upsert
	record := &NotificationRecord{
		ID:     notifier.ID,
		Name:   notifier.Name,
		Type:   notifier.Type,
		Config: notifier.Config,
	}
	err := c.DB.Insert(record)
	if err == nil {
		return nil
	}
	matched, _ := regexp.MatchString("duplicate.*notifications_pkey", err.Error())
	if !matched {
		return err
	}
	nr, err := c.findNotificationRecord(notifier.ID)
	if err != nil && err == sql.ErrNoRows {
		return errors.New("config: cannot find notification. name: " + record.Name)
	}
	if err != nil {
		return err
	}
	nr.Name = record.Name
	nr.Type = record.Type
	nr.Config = record.Config
	_, err = c.DB.Update(nr)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBStore) getAllNotificationRecords() (notifications []NotificationRecord, err error) {
	_, err = c.DB.Select(&notifications, "select "+notificationRecordColumns+" from notifications")
	return notifications, err
}

func NewDBStore(connectionURL string) (*DBStore, error) {
	db, err := sql.Open("postgres", connectionURL)
	if err != nil {
		return nil, err
	}
	gorpDB := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	gorpDB.AddTableWithName(CheckRecord{}, "checks").SetKeys(false, "ID")
	gorpDB.AddTableWithName(NotificationRecord{}, "notifications").SetKeys(false, "ID")
	gorpDB.AddTableWithName(PreferencesRecord{}, "preferences").SetKeys(false, "ID")
	gorpDB.TypeConverter = TypeConverter{}
	return &DBStore{DB: gorpDB}, nil
}

func (c *DBStore) Notifications() ([]notifiers.Config, error) {
	notificationConfigs := make([]notifiers.Config, 0)
	records, err := c.getAllNotificationRecords()
	if err != nil {
		return notificationConfigs, err
	}
	for _, record := range records {
		notificationConfigs = append(notificationConfigs, notifiers.Config{
			ID:     record.ID,
			Name:   record.Name,
			Type:   record.Type,
			Config: record.Config,
		})
	}
	return notificationConfigs, nil
}

func (c *DBStore) Checks() ([]checks.Config, error) {
	checkConfigs := make([]checks.Config, 0)
	records, err := c.getAllCheckRecords()
	if err != nil {
		return checkConfigs, err
	}
	for _, record := range records {
		checkConfigs = append(checkConfigs, checks.Config{
			ID:         record.ID,
			Name:       record.Name,
			Type:       record.Type,
			SendAlerts: record.SendAlerts,
			Backoff:    record.Backoff,
			Config:     record.Config,
			Assertions: record.Assertions,
		})
	}
	return checkConfigs, nil
}

type PreferencesRecord struct {
	ID          int             `db:"id"`
	Preferences json.RawMessage `db:"preferences"`
}

func (c *DBStore) getPreferences() (*PreferencesRecord, error) {
	r := new(PreferencesRecord)
	err := c.DB.SelectOne(r, "select id, preferences from preferences limit 1")
	return r, err
}

func (c *DBStore) updatePreferences(preferences Preferences) error {
	pr, err := c.getPreferences()
	if err != nil {
		return err
	}
	b, err := json.Marshal(preferences)
	fmt.Println(string(b))
	if err != nil {
		return err
	}
	pr.Preferences = json.RawMessage(b)
	_, err = c.DB.Update(pr)
	return err
}

func (c *DBStore) Preferences() (Preferences, error) {
	var p Preferences
	r, err := c.getPreferences()
	if err != nil {
		return Preferences{}, err
	}
	err = json.Unmarshal(r.Preferences, &p)
	return p, err
}

type TypeConverter struct{}

func (t TypeConverter) ToDb(val interface{}) (interface{}, error) {
	switch t := val.(type) {
	case map[string]string, []string, backoffs.Config, []assertions.Config:
		b, err := json.Marshal(t)
		if err != nil {
			return "", err
		}
		return string(b), nil
	case json.RawMessage:
		return string(t), nil
	}
	return val, nil
}

func (t TypeConverter) FromDb(target interface{}) (gorp.CustomScanner, bool) {
	switch target.(type) {
	case *map[string]string, *[]string, *backoffs.Config, *[]assertions.Config:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*string)
			if !ok {
				return errors.New("Unable to convert to *string")
			}
			b := []byte(*s)
			return json.Unmarshal(b, target)
		}
		return gorp.CustomScanner{new(string), target, binder}, true
	case *json.RawMessage:
		binder := func(holder, target interface{}) error {
			s, ok := holder.(*string)
			if !ok {
				return errors.New("Unable to convert to *string")
			}
			b := []byte(*s)
			st, ok := target.(*json.RawMessage)
			if !ok {
				return errors.New(fmt.Sprint("FromDb: Unable to convert target to *json.RawMessage: ", reflect.TypeOf(target)))
			}
			*st = json.RawMessage(b)
			return nil
		}
		return gorp.CustomScanner{new(string), target, binder}, true
	}
	return gorp.CustomScanner{}, false
}
