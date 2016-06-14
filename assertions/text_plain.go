package assertions

import (
	"errors"
	"fmt"
	"log"
)

func init() {
	Register("text/plain", NewTextPlain)
}

type TextPlain struct {
	Config
	log *log.Logger
}

var NewTextPlain = func(config Config, logger *log.Logger) (Asserter, error) {
	return Asserter(&TextPlain{config, logger}), nil
}

var UnknownTextPlainComparisonErr = errors.New("text/plain asserter: unknown comparison")

func (m *TextPlain) Assert(options Options) (Outcome, error) {
	current := string(options.CheckResponse.Response)
	return evaluateTextPlain(m.Identifier, m.Comparison, current, m.Target)
}

func (m *TextPlain) ValidateConfig() error {
	return nil
}

func evaluateTextPlain(identifier, operator, current, target string) (Outcome, error) {
	switch operator {
	case "==", "=", "equals":
		if current == target {
			return Outcome{Assertion: true}, nil
		}
		return Outcome{Assertion: false, Message: fmt.Sprintf("%s (%s) is not equal to %s", identifier, current, target)}, nil
	default:
		return Outcome{}, UnknownTextPlainComparisonErr
	}
}
