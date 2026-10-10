package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type configFormat int

const (
	jsonFormat configFormat = iota
	yamlFormat
)

func formatForPath(path string) configFormat {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		return yamlFormat
	default:
		return jsonFormat
	}
}

func decodeConfig(data []byte, format configFormat, target interface{}) error {
	if format == jsonFormat {
		return json.Unmarshal(data, target)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var value interface{}
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("YAML configuration must contain exactly one document")
		}
		return err
	}
	b, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("YAML contains values incompatible with JSON configuration: %w", err)
	}
	return json.Unmarshal(b, target)
}

func encodeConfig(value interface{}, format configFormat) ([]byte, error) {
	b, err := json.Marshal(value)
	if err != nil || format == jsonFormat {
		return b, err
	}
	var jsonValue interface{}
	if err := json.Unmarshal(b, &jsonValue); err != nil {
		return nil, err
	}
	return yaml.Marshal(jsonValue)
}
