package transform

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"
)

func testTransform[T comparable](t *testing.T, name string, transform func(T) T, original T) {
	t.Run(name, func(t *testing.T) {
		// Test 1: Same input yields same output (determinism)
		result1 := transform(original)
		result2 := transform(original)
		if result1 != result2 {
			t.Errorf("transform() = %v, %v, want same value for same input", result1, result2)
		}

		// Test 2: Different input yields different output (uniqueness)
		var different T
		switch v := any(original).(type) {
		case string:
			different = any(v + "different").(T)
		case int:
			different = any(v + 1).(T)
		case float64:
			different = any(v + 1.0).(T)
		case bool:
			different = any(!v).(T)
		case time.Time:
			different = any(v.Add(time.Hour)).(T)
		}
		result3 := transform(different)
		if result1 == result3 {
			t.Errorf("transform() = %v, want different value for different input", result3)
		}

		// Test 3: Input and output are different (transformation)
		if result1 == original {
			t.Errorf("transform() = %v, want different value from input", result1)
		}
	})
}

func testLimitedTransform[T comparable](t *testing.T, name string, transform func(T) T, original T, validValues []T) {
	t.Run(name, func(t *testing.T) {
		// Test 1: Same input yields same output (determinism)
		result1 := transform(original)
		result2 := transform(original)
		if result1 != result2 {
			t.Errorf("transform() = %v, %v, want same value for same input", result1, result2)
		}

		// Test 2: Result is one of the valid values
		found := false
		for _, v := range validValues {
			if result1 == v {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("transform() = %v, want one of %v", result1, validValues)
		}
	})
}

func TestTransformName(t *testing.T) {
	testTransform(t, "Name", TransformName, "test123")
}

func TestTransformFirstName(t *testing.T) {
	testTransform(t, "FirstName", TransformFirstName, "test123")
}

func TestTransformLastName(t *testing.T) {
	testTransform(t, "LastName", TransformLastName, "test123")
}

func TestTransformEmail(t *testing.T) {
	testTransform(t, "Email", TransformEmail, "test123")
}

func TestTransformSSN(t *testing.T) {
	testTransform(t, "SSN", TransformSSN, "test123")
}

func TestTransformDateOfBirth(t *testing.T) {
	testTransform(t, "DateOfBirth", TransformDateOfBirth, "2024-03-20")
}

func TestTransformPhone(t *testing.T) {
	testTransform(t, "Phone", TransformPhone, "test123")
}

func TestTransformGender(t *testing.T) {
	testLimitedTransform(t, "Gender", TransformGender, "test123", []string{"male", "female", "other"})
}

func TestTransformTitle(t *testing.T) {
	testTransform(t, "Title", TransformTitle, "test123")
}

func TestTransformJobTitle(t *testing.T) {
	testTransform(t, "JobTitle", TransformJobTitle, "test123")
}

func TestTransformIndustry(t *testing.T) {
	testTransform(t, "Industry", TransformIndustry, "test123")
}

func TestTransformDomainName(t *testing.T) {
	testTransform(t, "DomainName", TransformDomainName, "test123")
}

func TestTransformUsername(t *testing.T) {
	testTransform(t, "Username", TransformUsername, "test123")
}

func TestTransformPassword(t *testing.T) {
	testTransform(t, "Password", TransformPassword, "test123")
}

func TestTransformStreetAddress(t *testing.T) {
	testTransform(t, "StreetAddress", TransformStreetAddress, "test123")
}

func TestTransformStreet(t *testing.T) {
	testTransform(t, "Street", TransformStreet, "test123")
}

func TestTransformCity(t *testing.T) {
	testTransform(t, "City", TransformCity, "test123")
}

func TestTransformState(t *testing.T) {
	testTransform(t, "State", TransformState, "test123")
}

func TestTransformStateAbbr(t *testing.T) {
	testTransform(t, "StateAbbr", TransformStateAbbr, "test123")
}

func TestTransformZip(t *testing.T) {
	testTransform(t, "Zip", TransformZip, "test123")
}

func TestTransformCountry(t *testing.T) {
	testTransform(t, "Country", TransformCountry, "test123")
}

func TestTransformLatitude(t *testing.T) {
	testTransform(t, "Latitude", TransformLatitude, 0.0)
}

func TestTransformLongitude(t *testing.T) {
	testTransform(t, "Longitude", TransformLongitude, 0.0)
}

func TestTransformCompany(t *testing.T) {
	testTransform(t, "Company", TransformCompany, "test123")
}

func TestTransformProduct(t *testing.T) {
	testTransform(t, "Product", TransformProduct, "test123")
}

func TestTransformProductName(t *testing.T) {
	testTransform(t, "ProductName", TransformProductName, "test123")
}

func TestTransformProductDescription(t *testing.T) {
	testTransform(t, "ProductDescription", TransformProductDescription, "test123")
}

func TestTransformParagraph(t *testing.T) {
	testTransform(t, "Paragraph", TransformParagraph, "test123")
}

func TestTransformWord(t *testing.T) {
	testTransform(t, "Word", TransformWord, "test123")
}

func TestTransformMonth(t *testing.T) {
	testTransform(t, "Month", TransformMonth, "test123")
}

func TestTransformMonthNum(t *testing.T) {
	testTransform(t, "MonthNum", TransformMonthNum, 1)
}

func TestTransformWeekDay(t *testing.T) {
	testTransform(t, "WeekDay", TransformWeekDay, "test123")
}

func TestTransformYear(t *testing.T) {
	testTransform(t, "Year", TransformYear, 2024)
}

func TestTransformCreditCardType(t *testing.T) {
	testTransform(t, "CreditCardType", TransformCreditCardType, "test123")
}

func TestTransformCreditCardNum(t *testing.T) {
	testTransform(t, "CreditCardNum", TransformCreditCardNum, "test123")
}

func TestTransformCurrency(t *testing.T) {
	testTransform(t, "Currency", TransformCurrency, "test123")
}

func TestTransformBool(t *testing.T) {
	testLimitedTransform(t, "Bool", TransformBool, true, []bool{true, false})
}

// TestTransformOutputFormats verifies that transformed values have the expected format
func TestTransformOutputFormats(t *testing.T) {
	tests := []struct {
		name      string
		transform func(string) string
		input     string
		validate  func(string) error
	}{
		{
			name:      "Email format",
			transform: TransformEmail,
			input:     "test@example.com",
			validate: func(s string) error {
				if !strings.Contains(s, "@") {
					return fmt.Errorf("email missing @: %s", s)
				}
				if !strings.Contains(s, ".") {
					return fmt.Errorf("email missing domain: %s", s)
				}
				return nil
			},
		},
		{
			name:      "SSN format",
			transform: TransformSSN,
			input:     "123-45-6789",
			validate: func(s string) error {
				// Check format XXX-XX-XXXX
				matched, _ := regexp.MatchString(`^\d{3}-\d{2}-\d{4}$`, s)
				if !matched {
					return fmt.Errorf("SSN not in XXX-XX-XXXX format: %s", s)
				}
				return nil
			},
		},
		{
			name:      "Phone format",
			transform: TransformPhone,
			input:     "555-123-4567",
			validate: func(s string) error {
				// Check various phone formats
				matched, _ := regexp.MatchString(`^[\d\s\-\(\)\.]+$`, s)
				if !matched {
					return fmt.Errorf("phone contains invalid characters: %s", s)
				}
				return nil
			},
		},
		{
			name:      "Zip code format",
			transform: TransformZip,
			input:     "12345",
			validate: func(s string) error {
				matched, _ := regexp.MatchString(`^\d{5}(-\d{4})?$`, s)
				if !matched {
					return fmt.Errorf("zip not in XXXXX or XXXXX-XXXX format: %s", s)
				}
				return nil
			},
		},
		{
			name:      "State abbreviation format",
			transform: TransformStateAbbr,
			input:     "CA",
			validate: func(s string) error {
				if len(s) != 2 {
					return fmt.Errorf("state abbreviation not 2 characters: %s", s)
				}
				if strings.ToUpper(s) != s {
					return fmt.Errorf("state abbreviation not uppercase: %s", s)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.transform(tt.input)
			if err := tt.validate(result); err != nil {
				t.Errorf("Validation failed: %v", err)
			}
		})
	}
}

// TestTransformRanges verifies that numeric transforms produce values in expected ranges
func TestTransformRanges(t *testing.T) {
	// Test latitude range
	t.Run("Latitude range", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			lat := TransformLatitude(float64(i))
			if lat < -90 || lat > 90 {
				t.Errorf("Latitude out of range: %f", lat)
			}
		}
	})

	// Test longitude range
	t.Run("Longitude range", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			lon := TransformLongitude(float64(i))
			if lon < -180 || lon > 180 {
				t.Errorf("Longitude out of range: %f", lon)
			}
		}
	})

	// Test month number range
	t.Run("Month number range", func(t *testing.T) {
		for i := 1; i <= 12; i++ {
			month := TransformMonthNum(i)
			if month < 1 || month > 12 {
				t.Errorf("Month number out of range: %d", month)
			}
		}
	})
}

// TestTransformEdgeCases tests edge cases and special inputs
func TestTransformEdgeCases(t *testing.T) {
	t.Run("Empty string handling", func(t *testing.T) {
		// Should handle empty strings without panicking
		result1 := TransformEmail("")
		result2 := TransformEmail("")
		if result1 != result2 {
			t.Errorf("Empty string produced non-deterministic results")
		}
	})

	t.Run("Special characters", func(t *testing.T) {
		inputs := []string{"O'Brien", "José", "François"}
		for _, input := range inputs {
			result1 := TransformName(input)
			result2 := TransformName(input)
			if result1 != result2 {
				t.Errorf("Special character input %q produced non-deterministic results", input)
			}
		}
	})

	t.Run("Very long input", func(t *testing.T) {
		longInput := strings.Repeat("a", 1000)
		result1 := TransformParagraph(longInput)
		result2 := TransformParagraph(longInput)
		if result1 != result2 {
			t.Errorf("Long input produced non-deterministic results")
		}
	})
}
