package assertions

import (
	"errors"
	"log"

	"github.com/jonog/redalert/data"
)

type Config struct {

	// Source can be one of metric | metadata | application/json | text/plain
	Source string `json:"source"`

	// Relevant name. e.g. metric name
	Identifier string `json:"identifier"`

	// Comparison type e.g. == | >=
	Comparison string `json:"comparison"`

	// Target value to compare against
	Target string `json:"target"`
}

type Asserter interface {
	Assert(Options) (Outcome, error)
	ValidateConfig() error
}

type Options struct {
	CheckResponse data.CheckResponse
}

type Outcome struct {
	Assertion bool
	Message   string
}

var registry = make(map[string]func(Config, *log.Logger) (Asserter, error))

func Register(name string, constructorFn func(Config, *log.Logger) (Asserter, error)) {
	registry[name] = constructorFn
}

func New(config Config, logger *log.Logger) (Asserter, error) {
	checkerFn, ok := registry[config.Source]
	if !ok {
		return nil, errors.New("asserters: asserter unavailable: " + config.Source)
	}
	return checkerFn(config, logger)
}
