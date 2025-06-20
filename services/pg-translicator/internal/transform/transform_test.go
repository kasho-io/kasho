package transform

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"kasho/proto"
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

func TestTransformPasswordBcrypt(t *testing.T) {
	tests := []struct {
		name      string
		cleartext string
		useSalt   bool
		cost      int
		original  string
		wantErr   bool
	}{
		{
			name:      "basic bcrypt with salt",
			cleartext: "password123",
			useSalt:   true,
			cost:      4, // lower cost for faster testing
			original:  "testuser",
			wantErr:   false,
		},
		{
			name:      "bcrypt without salt",
			cleartext: "password123",
			useSalt:   false,
			cost:      4,
			original:  "testuser",
			wantErr:   false,
		},
		{
			name:      "long password truncation",
			cleartext: strings.Repeat("a", 100), // > 72 chars
			useSalt:   true,
			cost:      4,
			original:  "testuser",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1, err := TransformPasswordBcrypt(tt.cleartext, tt.useSalt, tt.cost, tt.original)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransformPasswordBcrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				// Test deterministic behavior
				hash2, err := TransformPasswordBcrypt(tt.cleartext, tt.useSalt, tt.cost, tt.original)
				if err != nil {
					t.Errorf("TransformPasswordBcrypt() second call error = %v", err)
					return
				}
				
				if tt.useSalt {
					// With salt, should be deterministic (same input -> same output)
					if hash1 != hash2 {
						t.Errorf("TransformPasswordBcrypt() with salt should be deterministic, got %v != %v", hash1, hash2)
					}
				}
				
				// Verify it looks like a bcrypt hash
				if !strings.HasPrefix(hash1, "$2") {
					t.Errorf("TransformPasswordBcrypt() result should start with $2, got %v", hash1)
				}
			}
		})
	}
}

func TestTransformPasswordScrypt(t *testing.T) {
	hash1, err := TransformPasswordScrypt("password123", true, 16384, 8, 1, "testuser")
	if err != nil {
		t.Errorf("TransformPasswordScrypt() error = %v", err)
		return
	}
	
	// Test deterministic behavior
	hash2, err := TransformPasswordScrypt("password123", true, 16384, 8, 1, "testuser")
	if err != nil {
		t.Errorf("TransformPasswordScrypt() second call error = %v", err)
		return
	}
	
	if hash1 != hash2 {
		t.Errorf("TransformPasswordScrypt() should be deterministic, got %v != %v", hash1, hash2)
	}
	
	// Should contain salt$hash format
	if !strings.Contains(hash1, "$") {
		t.Errorf("TransformPasswordScrypt() should contain $ separator, got %v", hash1)
	}
}

func TestTransformPasswordPBKDF2(t *testing.T) {
	hash1, err := TransformPasswordPBKDF2("password123", true, 10000, "SHA256", "testuser")
	if err != nil {
		t.Errorf("TransformPasswordPBKDF2() error = %v", err)
		return
	}
	
	// Test deterministic behavior
	hash2, err := TransformPasswordPBKDF2("password123", true, 10000, "SHA256", "testuser")
	if err != nil {
		t.Errorf("TransformPasswordPBKDF2() second call error = %v", err)
		return
	}
	
	if hash1 != hash2 {
		t.Errorf("TransformPasswordPBKDF2() should be deterministic, got %v != %v", hash1, hash2)
	}
	
	// Should contain salt$hash format
	if !strings.Contains(hash1, "$") {
		t.Errorf("TransformPasswordPBKDF2() should contain $ separator, got %v", hash1)
	}
}

func TestTransformPasswordArgon2id(t *testing.T) {
	hash1, err := TransformPasswordArgon2id("password123", true, 3, 1024, 4, "testuser")
	if err != nil {
		t.Errorf("TransformPasswordArgon2id() error = %v", err)
		return
	}
	
	// Test deterministic behavior
	hash2, err := TransformPasswordArgon2id("password123", true, 3, 1024, 4, "testuser")
	if err != nil {
		t.Errorf("TransformPasswordArgon2id() second call error = %v", err)
		return
	}
	
	if hash1 != hash2 {
		t.Errorf("TransformPasswordArgon2id() should be deterministic, got %v != %v", hash1, hash2)
	}
	
	// Should contain salt$hash format
	if !strings.Contains(hash1, "$") {
		t.Errorf("TransformPasswordArgon2id() should contain $ separator, got %v", hash1)
	}
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

// TestTransformRegex tests the regex transform functionality
func TestTransformRegex(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		replacement string
		input       string
		want        string
		wantErr     bool
	}{
		{
			name:        "Phone number masking",
			pattern:     `\d{3}-\d{3}-\d{4}`,
			replacement: "XXX-XXX-XXXX",
			input:       "123-456-7890",
			want:        "XXX-XXX-XXXX",
		},
		{
			name:        "Email domain replacement",
			pattern:     `@[\w.-]+\.[\w.-]+`,
			replacement: "@example.com",
			input:       "user@company.org",
			want:        "user@example.com",
		},
		{
			name:        "Partial replacement with capture groups",
			pattern:     `(\d{4})-(\d{4})-(\d{4})-(\d{4})`,
			replacement: "XXXX-XXXX-XXXX-$4",
			input:       "1234-5678-9012-3456",
			want:        "XXXX-XXXX-XXXX-3456",
		},
		{
			name:        "No match - return original",
			pattern:     `\d+`,
			replacement: "NUMBER",
			input:       "no numbers here",
			want:        "no numbers here",
		},
		{
			name:        "Multiple matches",
			pattern:     `\d+`,
			replacement: "X",
			input:       "abc123def456",
			want:        "abcXdefX",
		},
		{
			name:        "Invalid regex pattern",
			pattern:     `[`,
			replacement: "replacement",
			input:       "test",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformFunc := TransformRegex(tt.pattern, tt.replacement)
			got, err := transformFunc(tt.input)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("TransformRegex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && got != tt.want {
				t.Errorf("TransformRegex() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTransformTemplateFunction tests the template transform functionality
func TestTransformTemplateFunction(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		row         map[string]*proto.ColumnValue
		want        string
		wantErr     bool
	}{
		{
			name:     "Simple field access",
			template: "{{.name}}",
			row: map[string]*proto.ColumnValue{
				"name": {Value: &proto.ColumnValue_StringValue{StringValue: "John Doe"}},
			},
			want: "John Doe",
		},
		{
			name:     "Multiple fields",
			template: "{{.first_name}} {{.last_name}}",
			row: map[string]*proto.ColumnValue{
				"first_name": {Value: &proto.ColumnValue_StringValue{StringValue: "Jane"}},
				"last_name":  {Value: &proto.ColumnValue_StringValue{StringValue: "Smith"}},
			},
			want: "Jane Smith",
		},
		{
			name:     "Cross-column email generation",
			template: "{{.first_name | lower}}.{{.last_name | lower}}@company.com",
			row: map[string]*proto.ColumnValue{
				"first_name": {Value: &proto.ColumnValue_StringValue{StringValue: "John"}},
				"last_name":  {Value: &proto.ColumnValue_StringValue{StringValue: "Doe"}},
			},
			want: "john.doe@company.com",
		},
		{
			name:     "Helper function - lower",
			template: "{{.name | lower}}",
			row: map[string]*proto.ColumnValue{
				"name": {Value: &proto.ColumnValue_StringValue{StringValue: "JOHN DOE"}},
			},
			want: "john doe",
		},
		{
			name:     "Helper function - upper",
			template: "{{.name | upper}}",
			row: map[string]*proto.ColumnValue{
				"name": {Value: &proto.ColumnValue_StringValue{StringValue: "john doe"}},
			},
			want: "JOHN DOE",
		},
		{
			name:     "Helper function - slugify",
			template: "{{.title | slugify}}",
			row: map[string]*proto.ColumnValue{
				"title": {Value: &proto.ColumnValue_StringValue{StringValue: "Hello World! This is a Test."}},
			},
			want: "hello-world-this-is-a-test",
		},
		{
			name:     "Helper function - before",
			template: "{{.email | before \"@\"}}",
			row: map[string]*proto.ColumnValue{
				"email": {Value: &proto.ColumnValue_StringValue{StringValue: "user@example.com"}},
			},
			want: "user",
		},
		{
			name:     "Helper function - after",
			template: "{{.email | after \"@\"}}",
			row: map[string]*proto.ColumnValue{
				"email": {Value: &proto.ColumnValue_StringValue{StringValue: "user@example.com"}},
			},
			want: "example.com",
		},
		{
			name:     "Chained helpers",
			template: "{{.name | lower | slugify}}",
			row: map[string]*proto.ColumnValue{
				"name": {Value: &proto.ColumnValue_StringValue{StringValue: "John Doe Jr."}},
			},
			want: "john-doe-jr",
		},
		{
			name:     "Integer field",
			template: "User ID: {{.id}}",
			row: map[string]*proto.ColumnValue{
				"id": {Value: &proto.ColumnValue_IntValue{IntValue: 123}},
			},
			want: "User ID: 123",
		},
		{
			name:     "Float field",
			template: "Score: {{.score}}",
			row: map[string]*proto.ColumnValue{
				"score": {Value: &proto.ColumnValue_FloatValue{FloatValue: 95.5}},
			},
			want: "Score: 95.5",
		},
		{
			name:     "Boolean field",
			template: "Active: {{.active}}",
			row: map[string]*proto.ColumnValue{
				"active": {Value: &proto.ColumnValue_BoolValue{BoolValue: true}},
			},
			want: "Active: true",
		},
		{
			name:     "Timestamp field",
			template: "Created: {{.created_at}}",
			row: map[string]*proto.ColumnValue{
				"created_at": {Value: &proto.ColumnValue_TimestampValue{TimestampValue: "2024-01-01T12:00:00Z"}},
			},
			want: "Created: 2024-01-01T12:00:00Z",
		},
		{
			name:     "Complex business logic example",
			template: "{{if .active}}ACTIVE{{else}}INACTIVE{{end}}: {{.first_name}} {{.last_name}} ({{.email | after \"@\"}})",
			row: map[string]*proto.ColumnValue{
				"active":     {Value: &proto.ColumnValue_BoolValue{BoolValue: true}},
				"first_name": {Value: &proto.ColumnValue_StringValue{StringValue: "John"}},
				"last_name":  {Value: &proto.ColumnValue_StringValue{StringValue: "Doe"}},
				"email":      {Value: &proto.ColumnValue_StringValue{StringValue: "john@company.com"}},
			},
			want: "ACTIVE: John Doe (company.com)",
		},
		{
			name:        "Invalid template syntax",
			template:    "{{.name",
			row:         map[string]*proto.ColumnValue{},
			wantErr:     true,
		},
		{
			name:     "Missing field",
			template: "{{.missing_field}}",
			row: map[string]*proto.ColumnValue{
				"name": {Value: &proto.ColumnValue_StringValue{StringValue: "John"}},
			},
			want: "<no value>",
		},
		{
			name:     "Nil field",
			template: "{{.description}}",
			row: map[string]*proto.ColumnValue{
				"description": nil,
			},
			want: "<no value>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TransformTemplate(tt.template, tt.row)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("TransformTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && got != tt.want {
				t.Errorf("TransformTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTransformTemplateFunctionDeterminism tests that template transforms are deterministic
func TestTransformTemplateFunctionDeterminism(t *testing.T) {
	template := "{{.first_name | lower}}.{{.last_name | lower}}@{{.department | slugify}}.com"
	row := map[string]*proto.ColumnValue{
		"first_name": {Value: &proto.ColumnValue_StringValue{StringValue: "John"}},
		"last_name":  {Value: &proto.ColumnValue_StringValue{StringValue: "Doe"}},
		"department": {Value: &proto.ColumnValue_StringValue{StringValue: "Engineering & Development"}},
	}

	// Run the same template multiple times and ensure results are identical
	var results []string
	for i := 0; i < 5; i++ {
		result, err := TransformTemplate(template, row)
		if err != nil {
			t.Fatalf("TransformTemplate() error = %v", err)
		}
		results = append(results, result)
	}

	// All results should be identical
	for i := 1; i < len(results); i++ {
		if results[0] != results[i] {
			t.Errorf("TransformTemplate() produced different results: %v vs %v", results[0], results[i])
		}
	}

	// Verify expected result
	expected := "john.doe@engineering-development.com"
	if results[0] != expected {
		t.Errorf("TransformTemplate() = %v, want %v", results[0], expected)
	}
}
