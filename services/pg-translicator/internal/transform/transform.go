package transform

import (
	"fmt"
	"hash/fnv"
	"regexp"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

// ScalarValue represents any value that can be stored in a database column
type ScalarValue interface {
	~string | ~int | ~int64 | ~float64 | ~bool | time.Time
}

func hash[T ScalarValue](value T) uint64 {
	h := fnv.New64a()
	h.Write([]byte(fmt.Sprintf("%v", value)))
	return h.Sum64()
}

func seed[T ScalarValue](value T) {
	gofakeit.Seed(hash(value))
}

// Personal Information (Gofakeit-based)
func TransformFakeName(original string) string {
	seed(original)
	return gofakeit.Name()
}

func TransformFakeFirstName(original string) string {
	seed(original)
	return gofakeit.FirstName()
}

func TransformFakeLastName(original string) string {
	seed(original)
	return gofakeit.LastName()
}

func TransformFakeEmail(original string) string {
	seed(original)
	return gofakeit.Email()
}

func TransformFakeSSN(original string) string {
	seed(original)
	ssn := gofakeit.SSN()
	if len(ssn) == 9 {
		return fmt.Sprintf("%s-%s-%s", ssn[0:3], ssn[3:5], ssn[5:9])
	}
	return ssn
}

func TransformFakeDateOfBirth(original string) string {
	seed(original)
	date := gofakeit.Date()
	return date.Format("2006-01-02")
}

func TransformFakePhone(original string) string {
	seed(original)
	return gofakeit.Phone()
}

func TransformFakeGender(original string) string {
	seed(original)
	return gofakeit.Gender()
}

func TransformFakeTitle(original string) string {
	seed(original)
	return gofakeit.NamePrefix()
}

func TransformFakeJobTitle(original string) string {
	seed(original)
	return gofakeit.JobTitle()
}

func TransformFakeIndustry(original string) string {
	seed(original)
	return gofakeit.Company() + " Industry"
}

func TransformFakeDomainName(original string) string {
	seed(original)
	return gofakeit.DomainName()
}

func TransformFakeUsername(original string) string {
	seed(original)
	return gofakeit.Username()
}

func TransformFakePassword(original string) string {
	seed(original)
	return gofakeit.Password(true, true, true, true, true, 12)
}

// Address Information (Gofakeit-based)
func TransformFakeStreetAddress(original string) string {
	seed(original)
	return gofakeit.Address().Address
}

func TransformFakeStreet(original string) string {
	seed(original)
	return gofakeit.Address().Street
}

func TransformFakeCity(original string) string {
	seed(original)
	return gofakeit.Address().City
}

func TransformFakeState(original string) string {
	seed(original)
	return gofakeit.Address().State
}

func TransformFakeStateAbbr(original string) string {
	seed(original)
	return gofakeit.StateAbr()
}

func TransformFakeZip(original string) string {
	seed(original)
	return gofakeit.Address().Zip
}

func TransformFakeCountry(original string) string {
	seed(original)
	return gofakeit.Address().Country
}

func TransformFakeLatitude(original float64) float64 {
	seed(original)
	return gofakeit.Latitude()
}

func TransformFakeLongitude(original float64) float64 {
	seed(original)
	return gofakeit.Longitude()
}

// Product Information (Gofakeit-based)
func TransformFakeCompany(original string) string {
	seed(original)
	return gofakeit.Company()
}

func TransformFakeProduct(original string) string {
	seed(original)
	return gofakeit.Product().Name
}

func TransformFakeProductName(original string) string {
	seed(original)
	return gofakeit.ProductName()
}

func TransformFakeProductDescription(original string) string {
	seed(original)
	return gofakeit.ProductDescription()
}

// Text Content (Gofakeit-based)
func TransformFakeParagraph(original string) string {
	seed(original)
	return gofakeit.Paragraph(1, 3, 5, "\n")
}

func TransformFakeWord(original string) string {
	seed(original)
	return gofakeit.Word()
}

// Date and Time (Gofakeit-based)
func TransformFakeMonth(original string) string {
	seed(original)
	return gofakeit.MonthString()
}

func TransformFakeMonthNum(original int) int {
	seed(original)
	return int(gofakeit.Date().Month())
}

func TransformFakeWeekDay(original string) string {
	seed(original)
	return gofakeit.WeekDay()
}

func TransformFakeYear(original int) int {
	seed(original)
	return gofakeit.Date().Year()
}

// Financial Information (Gofakeit-based)
func TransformFakeCreditCardType(original string) string {
	seed(original)
	return gofakeit.CreditCardType()
}

func TransformFakeCreditCardNum(original string) string {
	seed(original)
	return gofakeit.CreditCardNumber(nil)
}

func TransformFakeCurrency(original string) string {
	seed(original)
	return gofakeit.Currency().Short
}

// Boolean
func TransformBool(original bool) bool {
	seed := hash(original)
	return seed%2 == 1
}

// Regex transform support
var (
	regexCache   = make(map[string]*regexp.Regexp)
	regexCacheMu sync.RWMutex
)

// getCompiledRegex returns a compiled regex from cache or compiles and caches it
func getCompiledRegex(pattern string) (*regexp.Regexp, error) {
	regexCacheMu.RLock()
	if compiled, exists := regexCache[pattern]; exists {
		regexCacheMu.RUnlock()
		return compiled, nil
	}
	regexCacheMu.RUnlock()

	// Compile regex
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Cache it
	regexCacheMu.Lock()
	regexCache[pattern] = compiled
	regexCacheMu.Unlock()

	return compiled, nil
}

// TransformRegex applies a regex pattern and replacement to a string
func TransformRegex(pattern, replacement string) func(string) (string, error) {
	return func(original string) (string, error) {
		re, err := getCompiledRegex(pattern)
		if err != nil {
			return "", err
		}
		return re.ReplaceAllString(original, replacement), nil
	}
}
