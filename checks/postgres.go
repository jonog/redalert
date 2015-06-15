package checks

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"

	_ "github.com/lib/pq"
	"gopkg.in/gorp.v1"
)

func init() {
	Register("postgres", NewPostgres)
}

type Postgres struct {
	ConnectionURL string        `json:"connection_url"`
	MetricQueries []MetricQuery `json:"metric_queries"`
}

type MetricQuery struct {
	Metric string `json:"metric"`
	Query  string `json:"query"`
}

var NewPostgres = func(config Config, logger *log.Logger) (Checker, error) {
	postgres := new(Postgres)
	err := json.Unmarshal([]byte(config.Config), postgres)
	if err != nil {
		return nil, err
	}
	if postgres.ConnectionURL == "" {
		return nil, errors.New("postgres: connection url cannot be blank")
	}
	if len(postgres.MetricQueries) == 0 {
		return nil, errors.New("postgres: no metrics to query")
	}
	return Checker(postgres), nil
}

func (p *Postgres) Check() (Metrics, error) {

	output := Metrics(make(map[string]*float64))

	db, err := initDB(p.ConnectionURL)
	if err != nil {
		return output, err
	}

	for _, mq := range p.MetricQueries {
		count, err := query(db, mq.Query)
		if err != nil {
			return output, err
		}
		metricVal := float64(count)
		output[mq.Metric] = &metricVal
	}

	return output, nil
}

func initDB(url string) (*gorp.DbMap, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	gorpDB := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	return gorpDB, nil
}

func query(db *gorp.DbMap, query string) (int64, error) {
	return db.SelectInt(query)
}

func (p *Postgres) MetricInfo(metric string) MetricInfo {
	return MetricInfo{Unit: ""}
}

func (p *Postgres) MessageContext() string {
	return ""
}
