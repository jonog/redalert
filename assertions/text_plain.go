package assertions

import (
	"errors"
	"fmt"
	"log"
)

func init() {
	Register("text", NewText)
}

type Text struct {
	Config
	log *log.Logger
}

var NewText = func(config Config, logger *log.Logger) (Asserter, error) {
	return Asserter(&Text{config, logger}), nil
}

var UnknownTextComparisonErr = errors.New("text asserter: unknown comparison")

func (m *Text) Assert(options Options) (Outcome, error) {
	current := string(options.CheckResponse.Response)
	return evaluateText(m.Comparison, current, m.Target)
}

func (m *Text) ValidateConfig() error {
	return nil
}

func evaluateText(operator, current, target string) (Outcome, error) {
	switch operator {
	case "==", "=", "equals":
		if current == target {
			return Outcome{Assertion: true}, nil
		}
		return Outcome{Assertion: false, Message: fmt.Sprintf("(%s) is not equal to %s", current, target)}, nil
	default:
		return Outcome{}, UnknownTextComparisonErr
	}
}
