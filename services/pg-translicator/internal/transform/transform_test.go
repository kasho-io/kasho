package transform

import (
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
