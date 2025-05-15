package main

import (
	"fmt"
	"os"
	"time"

	"github.com/icrowley/fake"
	"gopkg.in/yaml.v3"
)

// ScalarValue represents any value that can be stored in a database column
type ScalarValue interface {
	~string | ~int | ~int64 | ~float64 | ~bool | time.Time
}

// TransformFunction represents a function that generates fake data
type TransformFunction[T ScalarValue] func() T

// TransformType is an enum-like type for fake data generators
type TransformType string

const (
	// Personal Information
	FullName    TransformType = "FullName"
	FirstName   TransformType = "FirstName"
	LastName    TransformType = "LastName"
	Email       TransformType = "Email"
	SSN         TransformType = "SSN"
	DateOfBirth TransformType = "DateOfBirth"
	Phone       TransformType = "Phone"
	Gender      TransformType = "Gender"
	Title       TransformType = "Title"
	JobTitle    TransformType = "JobTitle"
	Company     TransformType = "Company"
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
	Product     TransformType = "Product"
	ProductName TransformType = "ProductName"
	Brand       TransformType = "Brand"
	Model       TransformType = "Model"

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
	CurrencyCode   TransformType = "CurrencyCode"

	// Boolean
	Bool TransformType = "Bool"
)

var transformFunctions = map[TransformType]any{
	// Personal Information
	FullName:  fake.FullName,
	FirstName: fake.FirstName,
	LastName:  fake.LastName,
	Email:     fake.EmailAddress,
	SSN:       func() string { return fmt.Sprintf("%s-%s-%s", fake.DigitsN(3), fake.DigitsN(2), fake.DigitsN(4)) },
	DateOfBirth: func() time.Time {
		return time.Date(fake.Year(1970, 2010), time.Month(fake.MonthNum()), fake.Day(), 0, 0, 0, 0, time.UTC)
	},
	Phone:      fake.Phone,
	Gender:     fake.Gender,
	Title:      fake.Title,
	JobTitle:   fake.JobTitle,
	Company:    fake.Company,
	Industry:   fake.Industry,
	DomainName: fake.DomainName,
	Username:   fake.UserName,
	Password:   fake.SimplePassword,

	// Address Information
	StreetAddress: fake.StreetAddress,
	Street:        fake.Street,
	City:          fake.City,
	State:         fake.State,
	StateAbbr:     fake.StateAbbrev,
	Zip:           fake.Zip,
	Country:       fake.Country,
	Latitude:      fake.Latitude,
	Longitude:     fake.Longitude,

	// Product Information
	Product:     fake.Product,
	ProductName: fake.ProductName,
	Brand:       fake.Brand,
	Model:       fake.Model,

	// Text Content
	Paragraph:  fake.Paragraph,
	Sentence:   fake.Sentence,
	Word:       fake.Word,
	Words:      fake.Words,
	Characters: fake.Characters,
	Character:  fake.Character,
	Digits:     fake.Digits,

	// Date and Time
	Month:    fake.Month,
	MonthNum: fake.MonthNum,
	WeekDay:  fake.WeekDay,
	Year:     func() int { return fake.Year(0, 3000) },

	// Financial Information
	CreditCardType: fake.CreditCardType,
	CreditCardNum:  func() string { return fake.CreditCardNum(fake.CreditCardType()) },
	Currency:       fake.Currency,
	CurrencyCode:   fake.CurrencyCode,

	// Boolean
	Bool: func() bool { return fake.MonthNum() < 7 },
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

// GetFakeValue generates a fake value for a given table and column
func (c *Config) GetFakeValue(table string, column string) (any, error) {
	tableConfig, exists := c.Tables[table]
	if !exists {
		return nil, fmt.Errorf("table %s not found in config", table)
	}

	fakeType, exists := tableConfig[column]
	if !exists {
		return nil, fmt.Errorf("column %s not found in table %s", column, table)
	}

	fn, err := fakeType.GetTransformFunction()
	if err != nil {
		return nil, err
	}

	switch f := fn.(type) {
	case func() string:
		return f(), nil
	case func() int:
		return f(), nil
	case func() float64:
		return f(), nil
	case func() bool:
		return f(), nil
	case func() time.Time:
		return f(), nil
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
