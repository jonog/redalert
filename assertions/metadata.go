package assertions

import (
	"errors"
	"fmt"
	"log"
)

func init() {
	Register("metadata", NewMetadata)
}

type Metadata struct {
	Config
	log *log.Logger
}

var NewMetadata = func(config Config, logger *log.Logger) (Asserter, error) {
	return Asserter(&Metadata{config, logger}), nil
}

var MissingMetadataErr = errors.New("metadata asserter: missing metadata")
var UnknownMetadataComparisonErr = errors.New("metadata asserter: unknown comparison")

func (m *Metadata) Assert(options Options) (Outcome, error) {
	currentVal, exists := options.CheckResponse.Metadata[m.Identifier]
	if !exists {
		return Outcome{}, MissingMetadataErr
	}
	return evaluateMetadata(m.Identifier, m.Comparison, currentVal, m.Target)
}

func (m *Metadata) ValidateConfig() error {
	return nil
}

func evaluateMetadata(identifier, operator, current, target string) (Outcome, error) {
	switch operator {
	case "==", "=", "equals":
		if current == target {
			return Outcome{Assertion: true}, nil
		}
		return Outcome{Assertion: false, Message: fmt.Sprintf("%s (%s) is not equal to %s", identifier, current, target)}, nil
	default:
		return Outcome{}, UnknownMetadataComparisonErr
	}
}
