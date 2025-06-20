package transform

import (
	"crypto/sha256"
	"fmt"
	"hash/fnv"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
	"kasho/proto"
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

// Template function helpers
var templateFuncMap = template.FuncMap{
	"lower": strings.ToLower,
	"upper": strings.ToUpper,
	"slugify": func(s string) string {
		// Convert to lowercase and replace non-alphanumeric with hyphens
		s = strings.ToLower(s)
		s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
		return strings.Trim(s, "-")
	},
	"before": func(sep, s string) string {
		if idx := strings.Index(s, sep); idx >= 0 {
			return s[:idx]
		}
		return s
	},
	"after": func(sep, s string) string {
		if idx := strings.Index(s, sep); idx >= 0 {
			return s[idx+len(sep):]
		}
		return ""
	},
}

// convertRowToTemplateData converts protobuf row data to a map suitable for templates
func convertRowToTemplateData(row map[string]*proto.ColumnValue) map[string]interface{} {
	data := make(map[string]interface{})
	for key, value := range row {
		if value == nil {
			data[key] = nil
			continue
		}
		
		switch v := value.Value.(type) {
		case *proto.ColumnValue_StringValue:
			data[key] = v.StringValue
		case *proto.ColumnValue_IntValue:
			data[key] = v.IntValue
		case *proto.ColumnValue_FloatValue:
			data[key] = v.FloatValue
		case *proto.ColumnValue_BoolValue:
			data[key] = v.BoolValue
		case *proto.ColumnValue_TimestampValue:
			data[key] = v.TimestampValue
		default:
			data[key] = nil
		}
	}
	return data
}

// TransformTemplate applies a Go template to generate values using full row context
func TransformTemplate(templateStr string, row map[string]*proto.ColumnValue) (string, error) {
	tmpl, err := template.New("transform").Funcs(templateFuncMap).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}
	
	data := convertRowToTemplateData(row)
	
	var result strings.Builder
	err = tmpl.Execute(&result, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	
	return result.String(), nil
}

// Password transform helper functions

// generateDeterministicSalt creates a deterministic salt based on the original value
func generateDeterministicSalt(original string, length int) []byte {
	h := sha256.New()
	h.Write([]byte(original))
	fullHash := h.Sum(nil)
	
	// If we need more bytes than SHA256 provides, cycle through the hash
	salt := make([]byte, length)
	for i := 0; i < length; i++ {
		salt[i] = fullHash[i%len(fullHash)]
	}
	return salt
}

// processPasswordCleartext handles template processing for cleartext field
func processPasswordCleartext(cleartext string, row map[string]*proto.ColumnValue) (string, error) {
	// If it contains template syntax, process it
	if strings.Contains(cleartext, "{{") {
		return TransformTemplate(cleartext, row)
	}
	// Otherwise return as-is
	return cleartext, nil
}

// TransformPasswordBcrypt applies bcrypt hashing to the cleartext
func TransformPasswordBcrypt(cleartext string, useSalt bool, cost int, original string) (string, error) {
	// Generate a deterministic "salt" by seeding the random generator
	// bcrypt generates its own salt internally, but we can make it deterministic
	// by using a consistent seed based on the original value
	if useSalt {
		seed(original) // This affects gofakeit's random generator
	}
	
	// bcrypt has a maximum password length of 72 bytes
	if len(cleartext) > 72 {
		cleartext = cleartext[:72]
	}
	
	// Generate hash
	hash, err := bcrypt.GenerateFromPassword([]byte(cleartext), cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash failed: %w", err)
	}
	
	return string(hash), nil
}

// TransformPasswordScrypt applies scrypt hashing to the cleartext
func TransformPasswordScrypt(cleartext string, useSalt bool, n, r, p int, original string) (string, error) {
	var salt []byte
	if useSalt {
		salt = generateDeterministicSalt(original, 16) // 16 bytes salt
	} else {
		salt = make([]byte, 16) // Empty salt
	}
	
	// Generate hash
	hash, err := scrypt.Key([]byte(cleartext), salt, n, r, p, 32) // 32 bytes output
	if err != nil {
		return "", fmt.Errorf("scrypt hash failed: %w", err)
	}
	
	// Format: salt$hash (both hex encoded)
	return fmt.Sprintf("%x$%x", salt, hash), nil
}

// TransformPasswordPBKDF2 applies PBKDF2 hashing to the cleartext
func TransformPasswordPBKDF2(cleartext string, useSalt bool, iterations int, hashFunc string, original string) (string, error) {
	var salt []byte
	if useSalt {
		salt = generateDeterministicSalt(original, 16) // 16 bytes salt
	} else {
		salt = make([]byte, 16) // Empty salt
	}
	
	// Only SHA256 supported for now (can extend later)
	if hashFunc != "SHA256" && hashFunc != "" {
		return "", fmt.Errorf("unsupported hash function: %s (only SHA256 supported)", hashFunc)
	}
	
	// Generate hash
	hash := pbkdf2.Key([]byte(cleartext), salt, iterations, 32, sha256.New)
	
	// Format: salt$hash (both hex encoded)
	return fmt.Sprintf("%x$%x", salt, hash), nil
}

// TransformPasswordArgon2id applies Argon2id hashing to the cleartext
func TransformPasswordArgon2id(cleartext string, useSalt bool, time, memory uint32, threads uint8, original string) (string, error) {
	var salt []byte
	if useSalt {
		salt = generateDeterministicSalt(original, 16) // 16 bytes salt
	} else {
		salt = make([]byte, 16) // Empty salt
	}
	
	// Generate hash
	hash := argon2.IDKey([]byte(cleartext), salt, time, memory, threads, 32) // 32 bytes output
	
	// Format: salt$hash (both hex encoded)
	return fmt.Sprintf("%x$%x", salt, hash), nil
}
