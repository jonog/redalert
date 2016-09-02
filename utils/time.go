package utils

import (
	"strings"
	"time"
)

type RFCTime struct {
	time.Time
}

func (d RFCTime) format() string {
	return d.Time.Format(time.RFC3339)
}

func (d RFCTime) String() string {
	return d.format()
}

func (d RFCTime) MarshalText() ([]byte, error) {
	return []byte(d.format()), nil
}

func (d RFCTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.format() + `"`), nil
}

func (d *RFCTime) UnmarshalJSON(b []byte) error {
	goTime, err := time.Parse(time.RFC3339, strings.Trim(string(b), "\""))
	if err != nil {
		return err
	}
	*d = RFCTime{goTime}
	return nil
}
