package transform

import (
	"fmt"
	"os"
	"strings"
	"time"

	"kasho/pkg/version"
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

	// Template-based transforms
	Template TransformType = "Template"

	// Password transforms with different algorithms
	PasswordBcrypt   TransformType = "PasswordBcrypt"
	PasswordScrypt   TransformType = "PasswordScrypt"
	PasswordPBKDF2   TransformType = "PasswordPBKDF2"
	PasswordArgon2id TransformType = "PasswordArgon2id"
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
	Type   TransformType  `yaml:"type"`
	Config map[string]any `yaml:",inline"`
}

// UnmarshalYAML handles both string and object formats
func (ct *ColumnTransform) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try to unmarshal as a string first (simple format)
	var transformType string
	if err := unmarshal(&transformType); err == nil {
		ct.Type = TransformType(transformType)
		ct.Config = make(map[string]any)
		return nil
	}

	// If that fails, try as a map (object format)
	var raw map[string]any
	if err := unmarshal(&raw); err != nil {
		return err
	}

	// Extract type field
	if typeVal, ok := raw["type"]; ok {
		if typeStr, ok := typeVal.(string); ok {
			ct.Type = TransformType(typeStr)
			delete(raw, "type") // Remove type from config
		} else {
			return fmt.Errorf("type field must be a string")
		}
	} else {
		return fmt.Errorf("type field is required")
	}

	// The rest is config
	if len(raw) > 0 {
		ct.Config = raw
	} else {
		ct.Config = make(map[string]any)
	}
	return nil
}

// TableConfig represents the configuration for a single table
type TableConfig map[string]ColumnTransform

// Config represents the entire configuration
type Config struct {
	MajorVersion int                    `yaml:"major_version"`
	Tables       map[string]TableConfig `yaml:"tables"`
}


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
	// Check if major version matches Kasho major version
	kashoMajorVersion := version.MajorVersion()
	
	if config.MajorVersion != kashoMajorVersion {
		return fmt.Errorf("config major version mismatch: got %d, expected %d (Kasho version %s)", 
			config.MajorVersion, kashoMajorVersion, version.Version)
	}

	return nil
}

// GetTransformedValue generates a transformed value for a given table, column, and original value
// For template and password transforms, it also accepts the full DMLData to provide row context
func GetTransformedValue(c *Config, table string, column string, original *proto.ColumnValue, dmlData *proto.DMLData) (*proto.ColumnValue, error) {
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
		// Extract pattern and replacement from config
		pattern, ok := colTransform.Config["pattern"].(string)
		if !ok {
			return nil, fmt.Errorf("regex transform requires 'pattern' field")
		}
		replacement, ok := colTransform.Config["replacement"].(string)
		if !ok {
			return nil, fmt.Errorf("regex transform requires 'replacement' field")
		}
		
		// Regex only works on string values
		if v, ok := original.Value.(*proto.ColumnValue_StringValue); ok {
			transformFunc := TransformRegex(pattern, replacement)
			transformed, err := transformFunc(v.StringValue)
			if err != nil {
				return nil, fmt.Errorf("regex transform failed: %w", err)
			}
			return &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: transformed}}, nil
		}
		return nil, fmt.Errorf("regex transform requires string value, got %T", original.Value)
	}

	// Handle Template transform specially
	if colTransform.Type == Template {
		// Extract template from config
		templateStr, ok := colTransform.Config["template"].(string)
		if !ok {
			return nil, fmt.Errorf("template transform requires 'template' field")
		}
		
		if dmlData == nil {
			return nil, fmt.Errorf("template transform requires DML data for row context")
		}
		
		// Build row context from DMLData
		rowContext := make(map[string]*proto.ColumnValue)
		for i, colName := range dmlData.ColumnNames {
			if i < len(dmlData.ColumnValues) {
				rowContext[colName] = dmlData.ColumnValues[i]
			}
		}
		
		transformed, err := TransformTemplate(templateStr, rowContext)
		if err != nil {
			return nil, fmt.Errorf("template transform failed: %w", err)
		}
		return &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: transformed}}, nil
	}

	// Handle Password transforms specially
	isPasswordTransform := colTransform.Type == PasswordBcrypt || 
		colTransform.Type == PasswordScrypt || 
		colTransform.Type == PasswordPBKDF2 || 
		colTransform.Type == PasswordArgon2id
	
	if isPasswordTransform {
		// Extract cleartext from config
		cleartext, ok := colTransform.Config["cleartext"].(string)
		if !ok {
			return nil, fmt.Errorf("password transform requires 'cleartext' field")
		}
		
		// Extract use_salt with default true
		useSalt := true
		if useSaltVal, ok := colTransform.Config["use_salt"]; ok {
			if b, ok := useSaltVal.(bool); ok {
				useSalt = b
			}
		}
		
		// Get original value as string for seeding
		originalStr := ""
		if v, ok := original.Value.(*proto.ColumnValue_StringValue); ok {
			originalStr = v.StringValue
		}
		
		// Process cleartext as template if needed
		if dmlData != nil && strings.Contains(cleartext, "{{") {
			// Build row context from DMLData
			rowContext := make(map[string]*proto.ColumnValue)
			for i, colName := range dmlData.ColumnNames {
				if i < len(dmlData.ColumnValues) {
					rowContext[colName] = dmlData.ColumnValues[i]
				}
			}
			
			processedCleartext, err := processPasswordCleartext(cleartext, rowContext)
			if err != nil {
				return nil, fmt.Errorf("failed to process cleartext template: %w", err)
			}
			cleartext = processedCleartext
		}
		
		var hashedPassword string
		var err error
		
		switch colTransform.Type {
		case PasswordBcrypt:
			cost := 10 // default
			if costVal, ok := colTransform.Config["cost"]; ok {
				if c, ok := costVal.(float64); ok { // YAML numbers come as float64
					cost = int(c)
				}
			}
			// Note: bcrypt doesn't use useSalt or originalStr - it always generates random salt
			hashedPassword, err = TransformPasswordBcrypt(cleartext, cost)
			
		case PasswordScrypt:
			n := 131072 // default 2^17
			r := 8
			p := 1
			if nVal, ok := colTransform.Config["n"]; ok {
				if val, ok := nVal.(float64); ok {
					n = int(val)
				}
			}
			if rVal, ok := colTransform.Config["r"]; ok {
				if val, ok := rVal.(float64); ok {
					r = int(val)
				}
			}
			if pVal, ok := colTransform.Config["p"]; ok {
				if val, ok := pVal.(float64); ok {
					p = int(val)
				}
			}
			hashedPassword, err = TransformPasswordScrypt(cleartext, useSalt, n, r, p, originalStr)
			
		case PasswordPBKDF2:
			iterations := 600000 // default
			hashFunc := "SHA256"
			if iterVal, ok := colTransform.Config["iterations"]; ok {
				if i, ok := iterVal.(float64); ok {
					iterations = int(i)
				}
			}
			if hashVal, ok := colTransform.Config["hash"]; ok {
				if h, ok := hashVal.(string); ok {
					hashFunc = h
				}
			}
			hashedPassword, err = TransformPasswordPBKDF2(cleartext, useSalt, iterations, hashFunc, originalStr)
			
		case PasswordArgon2id:
			time := uint32(3) // default
			memory := uint32(65536) // 64MB default
			threads := uint8(4) // default
			if timeVal, ok := colTransform.Config["time"]; ok {
				if t, ok := timeVal.(float64); ok {
					time = uint32(t)
				}
			}
			if memVal, ok := colTransform.Config["memory"]; ok {
				if m, ok := memVal.(float64); ok {
					memory = uint32(m)
				}
			}
			if threadVal, ok := colTransform.Config["threads"]; ok {
				if t, ok := threadVal.(float64); ok {
					threads = uint8(t)
				}
			}
			hashedPassword, err = TransformPasswordArgon2id(cleartext, useSalt, time, memory, threads, originalStr)
		}
		
		if err != nil {
			return nil, fmt.Errorf("password transform failed: %w", err)
		}
		
		return &proto.ColumnValue{Value: &proto.ColumnValue_StringValue{StringValue: hashedPassword}}, nil
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
// Uses a two-pass strategy: first processes non-Template transforms, then Template transforms
// with access to the already-transformed row data
func TransformChange(c *Config, change *proto.Change) (*proto.Change, error) {
	// Create a new Change object to avoid modifying the original
	newChange := &proto.Change{
		Position: change.Position,
		Type:     change.Type,
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

		// PASS 1: Transform all non-Template columns first
		for i, col := range newDML.ColumnNames {
			// Check if this column has a transform configured
			tableConfig, tableExists := c.Tables[newDML.Table]
			if !tableExists {
				// No transforms for this table, copy original value
				newDML.ColumnValues[i] = data.Dml.ColumnValues[i]
				continue
			}
			
			colTransform, colExists := tableConfig[col]
			if !colExists {
				// No transform for this column, copy original value
				newDML.ColumnValues[i] = data.Dml.ColumnValues[i]
				continue
			}
			
			// Skip Template and Password transforms in this pass
			if colTransform.Type == Template || 
				colTransform.Type == PasswordBcrypt ||
				colTransform.Type == PasswordScrypt ||
				colTransform.Type == PasswordPBKDF2 ||
				colTransform.Type == PasswordArgon2id {
				// For now, copy the original value (will be replaced in pass 2)
				newDML.ColumnValues[i] = data.Dml.ColumnValues[i]
				continue
			}
			
			// Process non-Template transforms
			transformed, err := GetTransformedValue(c, newDML.Table, col, data.Dml.ColumnValues[i], data.Dml)
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

		// PASS 2: Process Template and Password transforms with access to transformed row data
		for i, col := range newDML.ColumnNames {
			// Check if this column has a Template or Password transform configured
			tableConfig, tableExists := c.Tables[newDML.Table]
			if !tableExists {
				continue
			}
			
			colTransform, colExists := tableConfig[col]
			if !colExists {
				continue
			}
			
			// Check if it's a Template or Password transform
			isPass2Transform := colTransform.Type == Template ||
				colTransform.Type == PasswordBcrypt ||
				colTransform.Type == PasswordScrypt ||
				colTransform.Type == PasswordPBKDF2 ||
				colTransform.Type == PasswordArgon2id
			
			if !isPass2Transform {
				continue
			}
			
			// Create updated DMLData with transformed values for template context
			updatedDMLData := &proto.DMLData{
				Table:        newDML.Table,
				ColumnNames:  newDML.ColumnNames,
				ColumnValues: newDML.ColumnValues, // Use the transformed values from pass 1
				Kind:         newDML.Kind,
			}
			
			// Process Template transform with updated context
			transformed, err := GetTransformedValue(c, newDML.Table, col, data.Dml.ColumnValues[i], updatedDMLData)
			if err != nil {
				return nil, fmt.Errorf("error transforming template %s.%s: %w", newDML.Table, col, err)
			}
			if transformed != nil {
				newDML.ColumnValues[i] = transformed
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
