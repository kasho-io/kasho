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

func TestTransformFakeName(t *testing.T) {
	testTransform(t, "Name", TransformFakeName, "test123")
}

func TestTransformFakeFirstName(t *testing.T) {
	testTransform(t, "FirstName", TransformFakeFirstName, "test123")
}

func TestTransformFakeLastName(t *testing.T) {
	testTransform(t, "LastName", TransformFakeLastName, "test123")
}

func TestTransformFakeEmail(t *testing.T) {
	testTransform(t, "Email", TransformFakeEmail, "test123")
}

func TestTransformFakeSSN(t *testing.T) {
	testTransform(t, "SSN", TransformFakeSSN, "test123")
}

func TestTransformFakeDateOfBirth(t *testing.T) {
	testTransform(t, "DateOfBirth", TransformFakeDateOfBirth, "2024-03-20")
}

func TestTransformFakePhone(t *testing.T) {
	testTransform(t, "Phone", TransformFakePhone, "test123")
}

func TestTransformFakeGender(t *testing.T) {
	testLimitedTransform(t, "Gender", TransformFakeGender, "test123", []string{"male", "female", "other"})
}

func TestTransformFakeTitle(t *testing.T) {
	testTransform(t, "Title", TransformFakeTitle, "test123")
}

func TestTransformFakeJobTitle(t *testing.T) {
	testTransform(t, "JobTitle", TransformFakeJobTitle, "test123")
}

func TestTransformFakeIndustry(t *testing.T) {
	testTransform(t, "Industry", TransformFakeIndustry, "test123")
}

func TestTransformFakeDomainName(t *testing.T) {
	testTransform(t, "DomainName", TransformFakeDomainName, "test123")
}

func TestTransformFakeUsername(t *testing.T) {
	testTransform(t, "Username", TransformFakeUsername, "test123")
}

func TestTransformFakePassword(t *testing.T) {
	testTransform(t, "Password", TransformFakePassword, "test123")
}

func TestTransformFakeStreetAddress(t *testing.T) {
	testTransform(t, "StreetAddress", TransformFakeStreetAddress, "test123")
}

func TestTransformFakeStreet(t *testing.T) {
	testTransform(t, "Street", TransformFakeStreet, "test123")
}

func TestTransformFakeCity(t *testing.T) {
	testTransform(t, "City", TransformFakeCity, "test123")
}

func TestTransformFakeState(t *testing.T) {
	testTransform(t, "State", TransformFakeState, "test123")
}

func TestTransformFakeStateAbbr(t *testing.T) {
	testTransform(t, "StateAbbr", TransformFakeStateAbbr, "test123")
}

func TestTransformFakeZip(t *testing.T) {
	testTransform(t, "Zip", TransformFakeZip, "test123")
}

func TestTransformFakeCountry(t *testing.T) {
	testTransform(t, "Country", TransformFakeCountry, "test123")
}

func TestTransformFakeLatitude(t *testing.T) {
	testTransform(t, "Latitude", TransformFakeLatitude, 0.0)
}

func TestTransformFakeLongitude(t *testing.T) {
	testTransform(t, "Longitude", TransformFakeLongitude, 0.0)
}

func TestTransformFakeCompany(t *testing.T) {
	testTransform(t, "Company", TransformFakeCompany, "test123")
}

func TestTransformFakeProduct(t *testing.T) {
	testTransform(t, "Product", TransformFakeProduct, "test123")
}

func TestTransformFakeProductName(t *testing.T) {
	testTransform(t, "ProductName", TransformFakeProductName, "test123")
}

func TestTransformFakeProductDescription(t *testing.T) {
	testTransform(t, "ProductDescription", TransformFakeProductDescription, "test123")
}

func TestTransformFakeParagraph(t *testing.T) {
	testTransform(t, "Paragraph", TransformFakeParagraph, "test123")
}

func TestTransformFakeWord(t *testing.T) {
	testTransform(t, "Word", TransformFakeWord, "test123")
}

func TestTransformFakeMonth(t *testing.T) {
	testTransform(t, "Month", TransformFakeMonth, "test123")
}

func TestTransformFakeMonthNum(t *testing.T) {
	testTransform(t, "MonthNum", TransformFakeMonthNum, 1)
}

func TestTransformFakeWeekDay(t *testing.T) {
	testTransform(t, "WeekDay", TransformFakeWeekDay, "test123")
}

func TestTransformFakeYear(t *testing.T) {
	testTransform(t, "Year", TransformFakeYear, 2024)
}

func TestTransformFakeCreditCardType(t *testing.T) {
	testTransform(t, "CreditCardType", TransformFakeCreditCardType, "test123")
}

func TestTransformFakeCreditCardNum(t *testing.T) {
	testTransform(t, "CreditCardNum", TransformFakeCreditCardNum, "test123")
}

func TestTransformFakeCurrency(t *testing.T) {
	testTransform(t, "Currency", TransformFakeCurrency, "test123")
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
			transform: TransformFakeEmail,
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
			transform: TransformFakeSSN,
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
			transform: TransformFakePhone,
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
			transform: TransformFakeZip,
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
			transform: TransformFakeStateAbbr,
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
			lat := TransformFakeLatitude(float64(i))
			if lat < -90 || lat > 90 {
				t.Errorf("Latitude out of range: %f", lat)
			}
		}
	})

	// Test longitude range
	t.Run("Longitude range", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			lon := TransformFakeLongitude(float64(i))
			if lon < -180 || lon > 180 {
				t.Errorf("Longitude out of range: %f", lon)
			}
		}
	})

	// Test month number range
	t.Run("Month number range", func(t *testing.T) {
		for i := 1; i <= 12; i++ {
			month := TransformFakeMonthNum(i)
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
		result1 := TransformFakeEmail("")
		result2 := TransformFakeEmail("")
		if result1 != result2 {
			t.Errorf("Empty string produced non-deterministic results")
		}
	})

	t.Run("Special characters", func(t *testing.T) {
		inputs := []string{"O'Brien", "José", "François"}
		for _, input := range inputs {
			result1 := TransformFakeName(input)
			result2 := TransformFakeName(input)
			if result1 != result2 {
				t.Errorf("Special character input %q produced non-deterministic results", input)
			}
		}
	})

	t.Run("Very long input", func(t *testing.T) {
		longInput := strings.Repeat("a", 1000)
		result1 := TransformFakeParagraph(longInput)
		result2 := TransformFakeParagraph(longInput)
		if result1 != result2 {
			t.Errorf("Long input produced non-deterministic results")
		}
	})
}
