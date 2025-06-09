package transform

import (
	"fmt"
	"hash/fnv"
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

// Personal Information
func TransformName(original string) string {
	seed(original)
	return gofakeit.Name()
}

func TransformFirstName(original string) string {
	seed(original)
	return gofakeit.FirstName()
}

func TransformLastName(original string) string {
	seed(original)
	return gofakeit.LastName()
}

func TransformEmail(original string) string {
	seed(original)
	return gofakeit.Email()
}

func TransformSSN(original string) string {
	seed(original)
	ssn := gofakeit.SSN()
	if len(ssn) == 9 {
		return fmt.Sprintf("%s-%s-%s", ssn[0:3], ssn[3:5], ssn[5:9])
	}
	return ssn
}

func TransformDateOfBirth(original string) string {
	seed(original)
	date := gofakeit.Date()
	return date.Format("2006-01-02")
}

func TransformPhone(original string) string {
	seed(original)
	return gofakeit.Phone()
}

func TransformGender(original string) string {
	seed(original)
	return gofakeit.Gender()
}

func TransformTitle(original string) string {
	seed(original)
	return gofakeit.NamePrefix()
}

func TransformJobTitle(original string) string {
	seed(original)
	return gofakeit.JobTitle()
}

func TransformIndustry(original string) string {
	seed(original)
	return gofakeit.Company() + " Industry"
}

func TransformDomainName(original string) string {
	seed(original)
	return gofakeit.DomainName()
}

func TransformUsername(original string) string {
	seed(original)
	return gofakeit.Username()
}

func TransformPassword(original string) string {
	seed(original)
	return gofakeit.Password(true, true, true, true, true, 12)
}

// Address Information
func TransformStreetAddress(original string) string {
	seed(original)
	return gofakeit.Address().Address
}

func TransformStreet(original string) string {
	seed(original)
	return gofakeit.Address().Street
}

func TransformCity(original string) string {
	seed(original)
	return gofakeit.Address().City
}

func TransformState(original string) string {
	seed(original)
	return gofakeit.Address().State
}

func TransformStateAbbr(original string) string {
	seed(original)
	return gofakeit.StateAbr()
}

func TransformZip(original string) string {
	seed(original)
	return gofakeit.Address().Zip
}

func TransformCountry(original string) string {
	seed(original)
	return gofakeit.Address().Country
}

func TransformLatitude(original float64) float64 {
	seed(original)
	return gofakeit.Latitude()
}

func TransformLongitude(original float64) float64 {
	seed(original)
	return gofakeit.Longitude()
}

// Product Information
func TransformCompany(original string) string {
	seed(original)
	return gofakeit.Company()
}

func TransformProduct(original string) string {
	seed(original)
	return gofakeit.Product().Name
}

func TransformProductName(original string) string {
	seed(original)
	return gofakeit.ProductName()
}

func TransformProductDescription(original string) string {
	seed(original)
	return gofakeit.ProductDescription()
}

// Text Content
func TransformParagraph(original string) string {
	seed(original)
	return gofakeit.Paragraph(1, 3, 5, "\n")
}

func TransformWord(original string) string {
	seed(original)
	return gofakeit.Word()
}

// Date and Time
func TransformMonth(original string) string {
	seed(original)
	return gofakeit.MonthString()
}

func TransformMonthNum(original int) int {
	seed(original)
	return int(gofakeit.Date().Month())
}

func TransformWeekDay(original string) string {
	seed(original)
	return gofakeit.WeekDay()
}

func TransformYear(original int) int {
	seed(original)
	return gofakeit.Date().Year()
}

// Financial Information
func TransformCreditCardType(original string) string {
	seed(original)
	return gofakeit.CreditCardType()
}

func TransformCreditCardNum(original string) string {
	seed(original)
	return gofakeit.CreditCardNumber(nil)
}

func TransformCurrency(original string) string {
	seed(original)
	return gofakeit.Currency().Short
}

// Boolean
func TransformBool(original bool) bool {
	seed := hash(original)
	return seed%2 == 1
}
