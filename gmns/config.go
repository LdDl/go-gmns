package gmns

import "fmt"

// Config describes a working dataset. Physical values use meters, km/h and seconds.
// Coordinate geometry uses WGS84; interchange unit conversion belongs at the boundary.
type Config struct {
	datasetName   string
	idType        IDType
	currency      string
	versionNumber *float64
}

func NewConfig(kind IDType, options ...func(*Config)) *Config {
	config := &Config{idType: kind}
	for _, option := range options {
		option(config)
	}
	return config
}

func (config *Config) DatasetName() string {
	return config.datasetName
}

func (config *Config) IDType() IDType {
	return config.idType
}

func (config *Config) Currency() string {
	return config.currency
}

func (config *Config) ShortLength() string {
	return "meter"
}

func (config *Config) LongLength() string {
	return "meter"
}

func (config *Config) Speed() string {
	return "km/h"
}

func (config *Config) CRS() string {
	return "EPSG:4326"
}

func (config *Config) VersionNumber() (float64, bool) {
	if config.versionNumber == nil {
		return 0, false
	}
	return *config.versionNumber, true
}

func (config *Config) Clone() *Config {
	if config == nil {
		return nil
	}
	result := *config
	if config.versionNumber != nil {
		value := *config.versionNumber
		result.versionNumber = &value
	}
	return &result
}

func (config *Config) Validate() error {
	if config == nil {
		return fmt.Errorf("dataset config is missing")
	}
	err := config.idType.Validate()
	if err != nil {
		return err
	}
	value, present := config.VersionNumber()
	if present && (!Finite(value) || value <= 0) {
		return fmt.Errorf("version_number must be positive and finite")
	}
	return nil
}

func WithDatasetName(name string) func(*Config) {
	return func(config *Config) {
		config.datasetName = name
	}
}

func WithCurrency(currency string) func(*Config) {
	return func(config *Config) {
		config.currency = currency
	}
}

func WithVersionNumber(version float64) func(*Config) {
	return func(config *Config) {
		value := version
		config.versionNumber = &value
	}
}

func WithoutVersionNumber() func(*Config) {
	return func(config *Config) {
		config.versionNumber = nil
	}
}
