package assertions

import (
	"errors"
	"fmt"
	"log"
	"strconv"
)

func init() {
	Register("metric", NewMetric)
}

type Metric struct {
	Config
	log *log.Logger
}

var NewMetric = func(config Config, logger *log.Logger) (Asserter, error) {
	return Asserter(&Metric{config, logger}), nil
}

var MissingMetricErr = errors.New("metric asserter: missing metric")
var InvalidTargetTypeErr = errors.New("metric asserter: target is invalid numeric type")
var UnknownMetricComparisonErr = errors.New("metric asserter: unknown comparison")

func (m *Metric) Assert(options Options) (Outcome, error) {
	currentVal, exists := options.CheckResponse.Metrics[m.Identifier]
	if !exists || currentVal == nil {
		return Outcome{}, MissingMetricErr
	}
	// ignore error as validated upon loading config
	targetVal, _ := strconv.ParseFloat(m.Target, 64)
	return evaluateMetric(m.Identifier, m.Comparison, *currentVal, targetVal)
}

func (m *Metric) ValidateConfig() error {
	_, err := strconv.ParseFloat(m.Target, 64)
	if err != nil {
		return InvalidTargetTypeErr
	}
	return nil
}

func evaluateMetric(identifier, operator string, num1, num2 float64) (Outcome, error) {
	switch operator {
	case ">", "greater than":
		if num1 > num2 {
			return Outcome{Assertion: true}, nil
		}
		return Outcome{Assertion: false, Message: fmt.Sprintf("%s (%f) is not greater than %f", identifier, num1, num2)}, nil
	case ">=", "greater than or equal":
		if num1 >= num2 {
			return Outcome{Assertion: true}, nil
		}
		return Outcome{Assertion: false, Message: fmt.Sprintf("%s (%f) is not greater than or equal to %f", identifier, num1, num2)}, nil
	case "<", "less than":
		if num1 < num2 {
			return Outcome{Assertion: true}, nil
		}
		return Outcome{Assertion: false, Message: fmt.Sprintf("%s (%f) is not less than %f", identifier, num1, num2)}, nil
	case "<=", "less than or equal":
		if num1 <= num2 {
			return Outcome{Assertion: true}, nil
		}
		return Outcome{Assertion: false, Message: fmt.Sprintf("%s (%f) is not less than or equal to %f", identifier, num1, num2)}, nil
	case "==", "=", "equals":
		if num1 == num2 {
			return Outcome{Assertion: true}, nil
		}
		return Outcome{Assertion: false, Message: fmt.Sprintf("%s (%f) is not equal to %f", identifier, num1, num2)}, nil
	default:
		return Outcome{}, UnknownMetricComparisonErr
	}
}
