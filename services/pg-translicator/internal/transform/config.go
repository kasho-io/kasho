package transform

import (
	"fmt"
	"os"
	"time"

	"pg-change-stream/api"

	"gopkg.in/yaml.v3"
)

// TransformFunction represents a function that generates fake data
type TransformFunction[T ScalarValue] func(original string) T

// TransformType is an enum-like type for fake data generators
type TransformType string

const (
	// Personal Information
	Name        TransformType = "Name"
	FirstName   TransformType = "FirstName"
	LastName    TransformType = "LastName"
	Email       TransformType = "Email"
	SSN         TransformType = "SSN"
	DateOfBirth TransformType = "DateOfBirth"
	Phone       TransformType = "Phone"
	Gender      TransformType = "Gender"
	Title       TransformType = "Title"
	JobTitle    TransformType = "JobTitle"
	Industry    TransformType = "Industry"
	DomainName  TransformType = "DomainName"
	Username    TransformType = "Username"
	Password    TransformType = "Password"

	// Address Information
	StreetAddress TransformType = "StreetAddress"
	Street        TransformType = "Street"
	City          TransformType = "City"
	State         TransformType = "State"
	StateAbbr     TransformType = "StateAbbr"
	Zip           TransformType = "Zip"
	Country       TransformType = "Country"
	Latitude      TransformType = "Latitude"
	Longitude     TransformType = "Longitude"

	// Product Information
	Company            TransformType = "Company"
	Product            TransformType = "Product"
	ProductName        TransformType = "ProductName"
	ProductDescription TransformType = "ProductDescription"

	// Text Content
	Paragraph  TransformType = "Paragraph"
	Sentence   TransformType = "Sentence"
	Word       TransformType = "Word"
	Words      TransformType = "Words"
	Characters TransformType = "Characters"
	Character  TransformType = "Character"
	Digits     TransformType = "Digits"

	// Date and Time
	Month    TransformType = "Month"
	MonthNum TransformType = "MonthNum"
	WeekDay  TransformType = "WeekDay"
	Year     TransformType = "Year"

	// Financial Information
	CreditCardType TransformType = "CreditCardType"
	CreditCardNum  TransformType = "CreditCardNum"
	Currency       TransformType = "Currency"

	// Boolean
	Bool TransformType = "Bool"
)

var transformFunctions = map[TransformType]any{
	// Personal Information
	Name:        TransformName,
	FirstName:   TransformFirstName,
	LastName:    TransformLastName,
	Email:       TransformEmail,
	SSN:         TransformSSN,
	DateOfBirth: TransformDateOfBirth,
	Phone:       TransformPhone,
	Gender:      TransformGender,
	Title:       TransformTitle,
	JobTitle:    TransformJobTitle,
	Industry:    TransformIndustry,
	DomainName:  TransformDomainName,
	Username:    TransformUsername,
	Password:    TransformPassword,

	// Address Information
	StreetAddress: TransformStreetAddress,
	Street:        TransformStreet,
	City:          TransformCity,
	State:         TransformState,
	StateAbbr:     TransformStateAbbr,
	Zip:           TransformZip,
	Country:       TransformCountry,
	Latitude:      TransformLatitude,
	Longitude:     TransformLongitude,

	// Product Information
	Company:            TransformCompany,
	Product:            TransformProduct,
	ProductName:        TransformProductName,
	ProductDescription: TransformProductDescription,

	// Text Content
	Paragraph: TransformParagraph,
	Word:      TransformWord,

	// Date and Time
	Month:    TransformMonth,
	MonthNum: TransformMonthNum,
	WeekDay:  TransformWeekDay,
	Year:     TransformYear,

	// Financial Information
	CreditCardType: TransformCreditCardType,
	CreditCardNum:  TransformCreditCardNum,
	Currency:       TransformCurrency,

	// Boolean
	Bool: TransformBool,
}

// TableConfig represents the configuration for a single table
type TableConfig map[string]TransformType

// Config represents the entire configuration
type Config struct {
	Tables map[string]TableConfig `yaml:"tables"`
}

// LoadConfig loads the configuration from a YAML file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &config, nil
}

// GetFakeValue generates a fake value for a given table, column, and original value
func GetFakeValue[T ScalarValue](c *Config, table string, column string, original T) (any, error) {
	tableConfig, exists := c.Tables[table]
	if !exists {
		return nil, nil // not an error, just no transform for this table
	}

	fakeType, exists := tableConfig[column]
	if !exists {
		return nil, nil // not an error, just no transform for this column
	}

	fn, err := fakeType.GetTransformFunction()
	if err != nil {
		return nil, err
	}

	switch f := fn.(type) {
	case func(string) string:
		if str, ok := any(original).(string); ok {
			return f(str), nil
		}
		return nil, fmt.Errorf("expected string input, got %T", original)
	case func(int) int:
		if i, ok := any(original).(int); ok {
			return f(i), nil
		}
		return nil, fmt.Errorf("expected int input, got %T", original)
	case func(float64) float64:
		if flt, ok := any(original).(float64); ok {
			return f(flt), nil
		}
		return nil, fmt.Errorf("expected float64 input, got %T", original)
	case func(bool) bool:
		if b, ok := any(original).(bool); ok {
			return f(b), nil
		}
		return nil, fmt.Errorf("expected bool input, got %T", original)
	case func(time.Time) time.Time:
		if t, ok := any(original).(time.Time); ok {
			return f(t), nil
		}
		return nil, fmt.Errorf("expected time.Time input, got %T", original)
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
func TransformChange(c *Config, change *api.Change) (*api.Change, error) {
	// Create a new Change object to avoid modifying the original
	newChange := &api.Change{
		Lsn:  change.Lsn,
		Type: change.Type,
	}

	switch data := change.Data.(type) {
	case *api.Change_Dml:
		// Create a new DMLData object
		newDML := &api.DMLData{
			Table:        data.Dml.Table,
			ColumnNames:  make([]string, len(data.Dml.ColumnNames)),
			ColumnValues: make([]string, len(data.Dml.ColumnValues)),
			Kind:         data.Dml.Kind,
		}
		copy(newDML.ColumnNames, data.Dml.ColumnNames)
		copy(newDML.ColumnValues, data.Dml.ColumnValues)

		// Transform column values if configured
		for i, col := range newDML.ColumnNames {
			transformed, err := GetFakeValue(c, newDML.Table, col, newDML.ColumnValues[i])
			if err == nil && transformed != nil {
				// Only update if transformation was successful and returned a value
				newDML.ColumnValues[i] = fmt.Sprintf("%v", transformed)
			} else if err != nil {
				return nil, fmt.Errorf("error transforming %s.%s: %w", newDML.Table, col, err)
			}
		}

		// Copy old keys if present
		if data.Dml.OldKeys != nil {
			newDML.OldKeys = &api.OldKeys{
				KeyNames:  make([]string, len(data.Dml.OldKeys.KeyNames)),
				KeyValues: make([]string, len(data.Dml.OldKeys.KeyValues)),
			}
			copy(newDML.OldKeys.KeyNames, data.Dml.OldKeys.KeyNames)
			copy(newDML.OldKeys.KeyValues, data.Dml.OldKeys.KeyValues)
		}

		newChange.Data = &api.Change_Dml{Dml: newDML}

	case *api.Change_Ddl:
		// For DDL changes, just copy the DDL data
		newChange.Data = &api.Change_Ddl{
			Ddl: &api.DDLData{
				Ddl: data.Ddl.Ddl,
			},
		}

	default:
		return nil, fmt.Errorf("unsupported change type: %T", change.Data)
	}

	return newChange, nil
}
