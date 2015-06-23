package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/jonog/redalert/backoffs"
	"github.com/jonog/redalert/checks"
	"github.com/jonog/redalert/notifiers"

	_ "github.com/lib/pq"
	"gopkg.in/gorp.v1"
)

type ConfigDB struct {
	DB *gorp.DbMap
}

type CheckRecord struct {
	Id         int64
	Name       string           `db:"name"`
	Type       string           `db:"type"`
	SendAlerts []string         `db:"send_alerts"`
	Backoff    backoffs.Config  `db:"backoff"`
	Config     json.RawMessage  `db:"config"`
	Triggers   []checks.Trigger `db:"triggers"`
}

func (c *ConfigDB) CreateCheckRecord(check checks.Config) (*CheckRecord, error) {
	record := &CheckRecord{
		Name:       check.Name,
		Type:       check.Type,
		SendAlerts: check.SendAlerts,
		Backoff:    check.Backoff,
		Config:     check.Config,
		Triggers:   check.Triggers,
	}
	err := c.DB.Insert(record)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (c *ConfigDB) getAllCheckRecords() (checks []CheckRecord, err error) {
	_, err = c.DB.Select(&checks, "select id, name, type, send_alerts, backoff, config, triggers from checks")
	return checks, err
}

type NotificationRecord struct {
	Id     int64             `db:"id"`
	Name   string            `db:"name"`
	Type   string            `db:"type"`
	Config map[string]string `db:"config"`
}

func (c *ConfigDB) CreateNotificationRecord(notifier notifiers.Config) (*NotificationRecord, error) {
	record := &NotificationRecord{
		Name:   notifier.Name,
		Type:   notifier.Type,
		Config: notifier.Config,
	}
	err := c.DB.Insert(record)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (c *ConfigDB) getAllNotificationRecords() (notifications []NotificationRecord, err error) {
	_, err = c.DB.Select(&notifications, "select id, name, type, config from notifications")
	return notifications, err
}

func NewConfigDB(connectionURL string) (*ConfigDB, error) {
	db, err := sql.Open("postgres", connectionURL)
	if err != nil {
		return nil, err
	}
	gorpDB := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	gorpDB.AddTableWithName(CheckRecord{}, "checks").SetKeys(true, "Id")
	gorpDB.AddTableWithName(NotificationRecord{}, "notifications").SetKeys(true, "Id")
	gorpDB.TypeConverter = TypeConverter{}
	return &ConfigDB{DB: gorpDB}, nil
}

func (c *ConfigDB) Notifications() ([]notifiers.Config, error) {
	notificationConfigs := make([]notifiers.Config, 0)
	records, err := c.getAllNotificationRecords()
	if err != nil {
		return notificationConfigs, err
	}
	for _, record := range records {
		notificationConfigs = append(notificationConfigs, notifiers.Config{
			Name:   record.Name,
			Type:   record.Type,
			Config: record.Config,
		})
	}
	return notificationConfigs, nil
}

func (c *ConfigDB) Checks() ([]checks.Config, error) {
	checkConfigs := make([]checks.Config, 0)
	records, err := c.getAllCheckRecords()
	if err != nil {
		return checkConfigs, err
	}
	for _, record := range records {
		checkConfigs = append(checkConfigs, checks.Config{
			Name:       record.Name,
			Type:       record.Type,
			SendAlerts: record.SendAlerts,
			Backoff:    record.Backoff,
			Config:     record.Config,
			Triggers:   record.Triggers,
		})
	}
	return checkConfigs, nil
}

type TypeConverter struct{}

func (t TypeConverter) ToDb(val interface{}) (interface{}, error) {
	switch t := val.(type) {
	case map[string]string, []string, backoffs.Config, []checks.Trigger:
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
	case *map[string]string, *[]string, *backoffs.Config, *[]checks.Trigger:
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
				return errors.New(fmt.Sprint("FromDb: Unable to convert target to *CustomStringType: ", reflect.TypeOf(target)))
			}
			*st = json.RawMessage(b)
			return nil
		}
		return gorp.CustomScanner{new(string), target, binder}, true
	}
	return gorp.CustomScanner{}, false
}
