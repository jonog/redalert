package assertions

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/jsonq"
)

func init() {
	Register("json", NewJSON)
}

type JSON struct {
	Config
	log *log.Logger
}

var NewJSON = func(config Config, logger *log.Logger) (Asserter, error) {
	return Asserter(&JSON{config, logger}), nil
}

var (
	InvalidJSONIdentifierError = errors.New("json: invalid identifier")
	UnknownJSONComparisonErr   = errors.New("json: unknown comparison")
)

func (m *JSON) Assert(options Options) (Outcome, error) {
	data := map[string]interface{}{}
	dec := json.NewDecoder(bytes.NewReader(options.CheckResponse.Response))
	dec.Decode(&data)
	jq := jsonq.NewQuery(data)
	identifiedField, err := jq.String(strings.Split(m.Identifier, ".")...)
	if err != nil {
		return Outcome{}, InvalidJSONIdentifierError
	}
	return evaluateJSON(m.Comparison, identifiedField, m.Target)
}

func (m *JSON) ValidateConfig() error {
	return nil
}

func evaluateJSON(operator, current, target string) (Outcome, error) {
	switch operator {
	case "==", "=", "equals":
		if current == target {
			return Outcome{Assertion: true}, nil
		}
		return Outcome{Assertion: false, Message: fmt.Sprintf("(%s) is not equal to %s", current, target)}, nil
	default:
		return Outcome{}, UnknownJSONComparisonErr
	}
}
