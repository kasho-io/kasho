package transform

import (
	"fmt"
	"os"
	"time"

	"kasho/proto"

	"gopkg.in/yaml.v3"
)

// TransformFunction represents a function that generates fake data
type TransformFunction[T ScalarValue] func(original string) T

// TransformType is an enum-like type for fake data generators
type TransformType string

const (
	// Personal Information (Gofakeit-based)
	FakeName        TransformType = "FakeName"
	FakeFirstName   TransformType = "FakeFirstName"
	FakeLastName    TransformType = "FakeLastName"
	FakeEmail       TransformType = "FakeEmail"
	FakeSSN         TransformType = "FakeSSN"
	FakeDateOfBirth TransformType = "FakeDateOfBirth"
	FakePhone       TransformType = "FakePhone"
	FakeGender      TransformType = "FakeGender"
	FakeTitle       TransformType = "FakeTitle"
	FakeJobTitle    TransformType = "FakeJobTitle"
	FakeIndustry    TransformType = "FakeIndustry"
	FakeDomainName  TransformType = "FakeDomainName"
	FakeUsername    TransformType = "FakeUsername"
	FakePassword    TransformType = "FakePassword"

	// Address Information (Gofakeit-based)
	FakeStreetAddress TransformType = "FakeStreetAddress"
	FakeStreet        TransformType = "FakeStreet"
	FakeCity          TransformType = "FakeCity"
	FakeState         TransformType = "FakeState"
	FakeStateAbbr     TransformType = "FakeStateAbbr"
	FakeZip           TransformType = "FakeZip"
	FakeCountry       TransformType = "FakeCountry"
	FakeLatitude      TransformType = "FakeLatitude"
	FakeLongitude     TransformType = "FakeLongitude"

	// Product Information (Gofakeit-based)
	FakeCompany            TransformType = "FakeCompany"
	FakeProduct            TransformType = "FakeProduct"
	FakeProductName        TransformType = "FakeProductName"
	FakeProductDescription TransformType = "FakeProductDescription"

	// Text Content (Gofakeit-based)
	FakeParagraph  TransformType = "FakeParagraph"
	FakeSentence   TransformType = "FakeSentence"
	FakeWord       TransformType = "FakeWord"
	FakeWords      TransformType = "FakeWords"
	FakeCharacters TransformType = "FakeCharacters"
	FakeCharacter  TransformType = "FakeCharacter"
	FakeDigits     TransformType = "FakeDigits"

	// Date and Time (Gofakeit-based)
	FakeMonth    TransformType = "FakeMonth"
	FakeMonthNum TransformType = "FakeMonthNum"
	FakeWeekDay  TransformType = "FakeWeekDay"
	FakeYear     TransformType = "FakeYear"

	// Financial Information (Gofakeit-based)
	FakeCreditCardType TransformType = "FakeCreditCardType"
	FakeCreditCardNum  TransformType = "FakeCreditCardNum"
	FakeCurrency       TransformType = "FakeCurrency"

	// Custom transforms (non-gofakeit)
	Bool TransformType = "Bool"

	// Pattern-based transforms
	Regex TransformType = "Regex"
)

var transformFunctions = map[TransformType]any{
	// Personal Information (Gofakeit-based)
	FakeName:        TransformFakeName,
	FakeFirstName:   TransformFakeFirstName,
	FakeLastName:    TransformFakeLastName,
	FakeEmail:       TransformFakeEmail,
	FakeSSN:         TransformFakeSSN,
	FakeDateOfBirth: TransformFakeDateOfBirth,
	FakePhone:       TransformFakePhone,
	FakeGender:      TransformFakeGender,
	FakeTitle:       TransformFakeTitle,
	FakeJobTitle:    TransformFakeJobTitle,
	FakeIndustry:    TransformFakeIndustry,
	FakeDomainName:  TransformFakeDomainName,
	FakeUsername:    TransformFakeUsername,
	FakePassword:    TransformFakePassword,

	// Address Information (Gofakeit-based)
	FakeStreetAddress: TransformFakeStreetAddress,
	FakeStreet:        TransformFakeStreet,
	FakeCity:          TransformFakeCity,
	FakeState:         TransformFakeState,
	FakeStateAbbr:     TransformFakeStateAbbr,
	FakeZip:           TransformFakeZip,
	FakeCountry:       TransformFakeCountry,
	FakeLatitude:      TransformFakeLatitude,
	FakeLongitude:     TransformFakeLongitude,

	// Product Information (Gofakeit-based)
	FakeCompany:            TransformFakeCompany,
	FakeProduct:            TransformFakeProduct,
	FakeProductName:        TransformFakeProductName,
	FakeProductDescription: TransformFakeProductDescription,

	// Text Content (Gofakeit-based)
	FakeParagraph: TransformFakeParagraph,
	FakeWord:      TransformFakeWord,

	// Date and Time (Gofakeit-based)
	FakeMonth:    TransformFakeMonth,
	FakeMonthNum: TransformFakeMonthNum,
	FakeWeekDay:  TransformFakeWeekDay,
	FakeYear:     TransformFakeYear,

	// Financial Information (Gofakeit-based)
	FakeCreditCardType: TransformFakeCreditCardType,
	FakeCreditCardNum:  TransformFakeCreditCardNum,
	FakeCurrency:       TransformFakeCurrency,

	// Custom transforms (non-gofakeit)
	Bool: TransformBool,
}

func init() {
}

// ColumnTransform represents a transform configuration for a column
// It can be either a simple string (transform type) or a complex object
type ColumnTransform struct {
	Type        TransformType `yaml:"type"`
	Pattern     string        `yaml:"pattern,omitempty"`     // For Regex transforms
	Replacement string        `yaml:"replacement,omitempty"` // For Regex transforms
}

// UnmarshalYAML handles both string and object formats
func (ct *ColumnTransform) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try to unmarshal as a string first (simple format)
	var transformType string
	if err := unmarshal(&transformType); err == nil {
		ct.Type = TransformType(transformType)
		return nil
	}

	// If that fails, try as a struct (object format)
	type rawColumnTransform ColumnTransform
	var raw rawColumnTransform
	if err := unmarshal(&raw); err != nil {
		return err
	}

	*ct = ColumnTransform(raw)
	return nil
}

// TableConfig represents the configuration for a single table
type TableConfig map[string]ColumnTransform

// Config represents the entire configuration
type Config struct {
	Version string                   `yaml:"version"`
	Tables  map[string]TableConfig `yaml:"tables"`
}

// Supported configuration versions
const (
	ConfigVersionV1 = "v1"
	CurrentVersion  = ConfigVersionV1
)

// LoadConfig loads the configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Handle version validation and migration
	if err := validateAndMigrateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// validateAndMigrateConfig validates the config version and handles migrations
func validateAndMigrateConfig(config *Config) error {
	// Handle legacy configs without version field (assume v1)
	if config.Version == "" {
		fmt.Printf("Warning: No version specified in config, assuming %s\n", ConfigVersionV1)
		config.Version = ConfigVersionV1
	}

	switch config.Version {
	case ConfigVersionV1:
		// Current version, no migration needed
		return nil
	default:
		return fmt.Errorf("unsupported config version: %s (supported: %s)", 
			config.Version, ConfigVersionV1)
	}
}

// GetFakeValue generates a fake value for a given table, column, and original value
func GetFakeValue(c *Config, table string, column string, original *proto.ColumnValue) (*proto.ColumnValue, error) {
	tableConfig, exists := c.Tables[table]
	if !exists {
		return nil, nil // not an error, just no transform for this table
	}

	colTransform, exists := tableConfig[column]
	if !exists {
		return nil, nil // not an error, just no transform for this column
	}

	// Handle Regex transform specially
	if colTransform.Type == Regex {
		// Regex only works on string values
		if v, ok := original.Value.(*proto.ColumnValue_StringValue); ok {
			transformFunc := TransformRegex(colTransform.Pattern, colTransform.Replacement)
			transformed, err := transformFunc(v.StringValue)
			if err != nil {
				return nil, fmt.Errorf("regex transform failed: %w", err)
			}
			return &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: transformed}}, nil
		}
		return nil, fmt.Errorf("regex transform requires string value, got %T", original.Value)
	}

	// For other transforms, use the existing logic
	fn, err := colTransform.Type.GetTransformFunction()
	if err != nil {
		return nil, err
	}

	// Extract the raw value based on its type
	var rawValue any
	switch v := original.Value.(type) {
	case *proto.ColumnValue_StringValue:
		rawValue = v.StringValue
	case *proto.ColumnValue_IntValue:
		rawValue = v.IntValue
	case *proto.ColumnValue_FloatValue:
		rawValue = v.FloatValue
	case *proto.ColumnValue_BoolValue:
		rawValue = v.BoolValue
	case *proto.ColumnValue_TimestampValue:
		if t, err := time.Parse(time.RFC3339, v.TimestampValue); err == nil {
			rawValue = t
		} else {
			rawValue = v.TimestampValue
		}
	default:
		return nil, fmt.Errorf("unsupported value type: %T", original.Value)
	}

	// Apply the transform function
	switch f := fn.(type) {
	case func(string) string:
		if str, ok := rawValue.(string); ok {
			transformed := f(str)
			return &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: transformed}}, nil
		}
		return nil, fmt.Errorf("expected string input, got %T", rawValue)
	case func(int) int:
		if i, ok := rawValue.(int64); ok {
			transformed := f(int(i))
			return &proto.ColumnValue{Value: &proto.ColumnValue_IntValue{IntValue: int64(transformed)}}, nil
		}
		return nil, fmt.Errorf("expected int64 input, got %T", rawValue)
	case func(float64) float64:
		if flt, ok := rawValue.(float64); ok {
			transformed := f(flt)
			return &proto.ColumnValue{Value: &proto.ColumnValue_FloatValue{FloatValue: transformed}}, nil
		}
		return nil, fmt.Errorf("expected float64 input, got %T", rawValue)
	case func(bool) bool:
		if b, ok := rawValue.(bool); ok {
			transformed := f(b)
			return &proto.ColumnValue{Value: &proto.ColumnValue_BoolValue{BoolValue: transformed}}, nil
		}
		return nil, fmt.Errorf("expected bool input, got %T", rawValue)
	case func(time.Time) time.Time:
		if t, ok := rawValue.(time.Time); ok {
			transformed := f(t)
			return &proto.ColumnValue{Value: &proto.ColumnValue_TimestampValue{TimestampValue: transformed.Format(time.RFC3339)}}, nil
		}
		return nil, fmt.Errorf("expected time.Time input, got %T", rawValue)
	default:
		return nil, fmt.Errorf("unsupported function type: %T", fn)
	}
}

// GetTransformFunction returns the corresponding fake function for a TransformType
func (ft TransformType) GetTransformFunction() (any, error) {
	if fn, exists := transformFunctions[ft]; exists {
		return fn, nil
	}
	return nil, fmt.Errorf("unknown transform type: %s", ft)
}

// TransformChange takes a Change object and returns a new Change object with transformed values
func TransformChange(c *Config, change *proto.Change) (*proto.Change, error) {
	// Create a new Change object to avoid modifying the original
	newChange := &proto.Change{
		Lsn:  change.Lsn,
		Type: change.Type,
	}

	switch data := change.Data.(type) {
	case *proto.Change_Dml:
		// Create a new DMLData object
		newDML := &proto.DMLData{
			Table:        data.Dml.Table,
			ColumnNames:  make([]string, len(data.Dml.ColumnNames)),
			ColumnValues: make([]*proto.ColumnValue, len(data.Dml.ColumnValues)),
			Kind:         data.Dml.Kind,
		}
		copy(newDML.ColumnNames, data.Dml.ColumnNames)

		// Transform column values if configured
		for i, col := range newDML.ColumnNames {
			transformed, err := GetFakeValue(c, newDML.Table, col, data.Dml.ColumnValues[i])
			if err != nil {
				return nil, fmt.Errorf("error transforming %s.%s: %w", newDML.Table, col, err)
			}
			if transformed != nil {
				newDML.ColumnValues[i] = transformed
			} else {
				// If no transformation, copy the original value
				newDML.ColumnValues[i] = data.Dml.ColumnValues[i]
			}
		}

		// Copy old keys if present
		if data.Dml.OldKeys != nil {
			newDML.OldKeys = &proto.OldKeys{
				KeyNames:  make([]string, len(data.Dml.OldKeys.KeyNames)),
				KeyValues: make([]*proto.ColumnValue, len(data.Dml.OldKeys.KeyValues)),
			}
			copy(newDML.OldKeys.KeyNames, data.Dml.OldKeys.KeyNames)
			copy(newDML.OldKeys.KeyValues, data.Dml.OldKeys.KeyValues)
		}

		newChange.Data = &proto.Change_Dml{Dml: newDML}

	case *proto.Change_Ddl:
		// For DDL changes, just copy the DDL data
		newChange.Data = &proto.Change_Ddl{
			Ddl: &proto.DDLData{
				Ddl: data.Ddl.Ddl,
			},
		}
	}

	return newChange, nil
}
