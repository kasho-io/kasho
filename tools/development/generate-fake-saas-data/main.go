package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	"kasho/pkg/dialect"
)

const (
	// Data volumes
	numOrganizations   = 10
	maxUsersPerOrg     = 100
	minUsersPerOrg     = 1
	maxCardsPerOrg     = 3
	minCardsPerOrg     = 1
	maxProjectsPerOrg  = 20
	minProjectsPerOrg  = 1
	maxTasksPerProject = 50
	minTasksPerProject = 1

	// Time ranges
	maxMonthsBack = 36
)

type Organization struct {
	ID             string
	Name           string
	BillingAddress string
	OwnerID        string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type User struct {
	ID             string
	OrganizationID string
	Name           string
	Email          string
	Password       string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Subscription struct {
	ID                  string
	OrganizationID      string
	PlanID              string
	MonthlyPerUserPrice float64
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type CreditCard struct {
	ID             string
	OrganizationID string
	Number         string
	ExpMonth       int
	ExpYear        int
	CVV            string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Invoice struct {
	ID             string
	OrganizationID string
	Date           time.Time
	Cost           float64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Project struct {
	ID             string
	OrganizationID string
	OwnerID        string
	Name           string
	Description    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Task struct {
	ID             string
	OrganizationID string
	ProjectID      string
	AssigneeID     string
	Name           string
	Description    string
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func main() {
	// Parse command-line flags
	dialectName := flag.String("dialect", "postgresql", "SQL dialect: postgresql or mysql")
	flag.Parse()

	// Get dialect
	d, err := dialect.FromName(*dialectName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Supported dialects: postgresql, mysql\n")
		os.Exit(1)
	}

	// Seed the random number generator
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	gofakeit.Seed(time.Now().UnixNano())

	// Write DDL
	writeDDL(os.Stdout, d)

	// Generate and write data
	organizations := generateOrganizations(r)
	users := generateUsers(organizations, r)
	subscriptions := generateSubscriptions(organizations, r)
	creditCards := generateCreditCards(organizations, r)
	invoices := generateInvoices(organizations, users, subscriptions, r)
	projects := generateProjects(organizations, users, r)
	tasks := generateTasks(organizations, projects, users, r)

	// Write DML
	writeDML(os.Stdout, d, organizations, users, subscriptions, creditCards, invoices, projects, tasks)
}

func writeDDL(f *os.File, d dialect.Dialect) {
	// Use dialect-specific type names
	uuid := d.TypeUUID()
	text := d.TypeText()
	ts := d.TypeTimestamp()
	decimal := d.TypeDecimal(10, 2)
	integer := d.TypeInteger()

	// MySQL-specific prefix
	if d.Name() == "mysql" {
		f.WriteString("SET FOREIGN_KEY_CHECKS = 0;\n\n")
	}

	// Build DDL with dialect-specific types
	ddl := fmt.Sprintf(`CREATE TABLE organizations (
    id %s PRIMARY KEY,
    name %s NOT NULL,
    billing_address %s NOT NULL,
    owner_id %s NOT NULL,
    created_at %s NOT NULL,
    updated_at %s NOT NULL
)`, uuid, text, text, uuid, ts, ts)
	if d.Name() == "mysql" {
		ddl += " ENGINE=InnoDB"
	}
	ddl += ";\n\n"
	f.WriteString(ddl)

	ddl = fmt.Sprintf(`CREATE TABLE users (
    id %s PRIMARY KEY,
    organization_id %s NOT NULL,
    name %s NOT NULL,
    email %s NOT NULL,
    password %s NOT NULL,
    created_at %s NOT NULL,
    updated_at %s NOT NULL
)`, uuid, uuid, text, text, text, ts, ts)
	if d.Name() == "mysql" {
		ddl += " ENGINE=InnoDB"
	}
	ddl += ";\n\n"
	f.WriteString(ddl)

	ddl = fmt.Sprintf(`CREATE TABLE subscriptions (
    id %s PRIMARY KEY,
    organization_id %s NOT NULL,
    plan_id %s NOT NULL,
    monthly_per_user_price %s NOT NULL,
    created_at %s NOT NULL,
    updated_at %s NOT NULL
)`, uuid, uuid, text, decimal, ts, ts)
	if d.Name() == "mysql" {
		ddl += " ENGINE=InnoDB"
	}
	ddl += ";\n\n"
	f.WriteString(ddl)

	ddl = fmt.Sprintf(`CREATE TABLE credit_cards (
    id %s PRIMARY KEY,
    organization_id %s NOT NULL,
    number %s NOT NULL,
    exp_month %s NOT NULL,
    exp_year %s NOT NULL,
    cvv %s NOT NULL,
    created_at %s NOT NULL,
    updated_at %s NOT NULL
)`, uuid, uuid, text, integer, integer, text, ts, ts)
	if d.Name() == "mysql" {
		ddl += " ENGINE=InnoDB"
	}
	ddl += ";\n\n"
	f.WriteString(ddl)

	ddl = fmt.Sprintf(`CREATE TABLE invoices (
    id %s PRIMARY KEY,
    organization_id %s NOT NULL,
    date DATE NOT NULL,
    cost %s NOT NULL,
    created_at %s NOT NULL,
    updated_at %s NOT NULL
)`, uuid, uuid, decimal, ts, ts)
	if d.Name() == "mysql" {
		ddl += " ENGINE=InnoDB"
	}
	ddl += ";\n\n"
	f.WriteString(ddl)

	ddl = fmt.Sprintf(`CREATE TABLE projects (
    id %s PRIMARY KEY,
    organization_id %s NOT NULL,
    owner_id %s NOT NULL,
    name %s NOT NULL,
    description %s,
    created_at %s NOT NULL,
    updated_at %s NOT NULL
)`, uuid, uuid, uuid, text, text, ts, ts)
	if d.Name() == "mysql" {
		ddl += " ENGINE=InnoDB"
	}
	ddl += ";\n\n"
	f.WriteString(ddl)

	ddl = fmt.Sprintf(`CREATE TABLE tasks (
    id %s PRIMARY KEY,
    organization_id %s NOT NULL,
    project_id %s NOT NULL,
    assignee_id %s NOT NULL,
    name %s NOT NULL,
    description %s,
    status %s NOT NULL,
    created_at %s NOT NULL,
    updated_at %s NOT NULL
)`, uuid, uuid, uuid, uuid, text, text, text, ts, ts)
	if d.Name() == "mysql" {
		ddl += " ENGINE=InnoDB"
	}
	ddl += ";\n\n"
	f.WriteString(ddl)

	// MySQL-specific suffix
	if d.Name() == "mysql" {
		f.WriteString("SET FOREIGN_KEY_CHECKS = 1;\n")
	}
}

func generateOrganizations(r *rand.Rand) []Organization {
	orgs := make([]Organization, numOrganizations)
	for i := 0; i < numOrganizations; i++ {
		createdAt := randomTimeInPast(maxMonthsBack, r)
		orgs[i] = Organization{
			ID:             uuid.New().String(),
			Name:           gofakeit.Company(),
			BillingAddress: gofakeit.Address().Address,
			CreatedAt:      createdAt,
			UpdatedAt:      randomTimeAfter(createdAt, r),
		}
	}
	return orgs
}

func generateUsers(orgs []Organization, r *rand.Rand) []User {
	var users []User
	for i, org := range orgs {
		numUsers := r.Intn(maxUsersPerOrg-minUsersPerOrg+1) + minUsersPerOrg
		for j := 0; j < numUsers; j++ {
			createdAt := randomTimeInPast(maxMonthsBack, r)
			user := User{
				ID:             uuid.New().String(),
				OrganizationID: org.ID,
				Name:           gofakeit.Name(),
				Email:          gofakeit.Email(),
				CreatedAt:      createdAt,
				UpdatedAt:      randomTimeAfter(createdAt, r),
			}
			// Generate password hash using Argon2id with same parameters as translicator
			// Use a deterministic cleartext based on user info for consistent demo data
			cleartext := "password123" // Simple cleartext for demo
			user.Password = generatePasswordArgon2id(
				cleartext,  // cleartext password
				true,       // useSalt
				3,          // time (default from translicator)
				65536,      // memory (default: 64MB in KiB)
				4,          // threads (default)
				user.ID,    // original (for deterministic salt)
			)
			users = append(users, user)

			// Set the first user of each organization as its owner
			if j == 0 {
				orgs[i].OwnerID = user.ID
			}
		}
	}
	return users
}

func generateSubscriptions(orgs []Organization, r *rand.Rand) []Subscription {
	var subs []Subscription
	prices := []float64{5.00, 10.00, 20.00}
	for _, org := range orgs {
		createdAt := randomTimeInPast(maxMonthsBack, r)
		sub := Subscription{
			ID:                  uuid.New().String(),
			OrganizationID:      org.ID,
			PlanID:              fmt.Sprintf("plan_%d", r.Intn(3)+1),
			MonthlyPerUserPrice: prices[r.Intn(len(prices))],
			CreatedAt:           createdAt,
			UpdatedAt:           randomTimeAfter(createdAt, r),
		}
		subs = append(subs, sub)
	}
	return subs
}

func generateCreditCards(orgs []Organization, r *rand.Rand) []CreditCard {
	var cards []CreditCard
	for _, org := range orgs {
		numCards := r.Intn(maxCardsPerOrg-minCardsPerOrg+1) + minCardsPerOrg
		for i := 0; i < numCards; i++ {
			createdAt := randomTimeInPast(maxMonthsBack, r)
			expYear := time.Now().Year() + r.Intn(5) + 1
			card := CreditCard{
				ID:             uuid.New().String(),
				OrganizationID: org.ID,
				Number:         generateValidCardNumber(r),
				ExpMonth:       r.Intn(12) + 1,
				ExpYear:        expYear,
				CVV:            fmt.Sprintf("%03d", r.Intn(1000)),
				CreatedAt:      createdAt,
				UpdatedAt:      randomTimeAfter(createdAt, r),
			}
			cards = append(cards, card)
		}
	}
	return cards
}

func generateInvoices(orgs []Organization, users []User, subs []Subscription, r *rand.Rand) []Invoice {
	var invoices []Invoice
	for _, org := range orgs {
		// Find subscription for this org
		var sub Subscription
		for _, s := range subs {
			if s.OrganizationID == org.ID {
				sub = s
				break
			}
		}

		// Count users in this org
		var userCount int
		for _, u := range users {
			if u.OrganizationID == org.ID {
				userCount++
			}
		}

		// Generate monthly invoices
		monthsBack := r.Intn(maxMonthsBack)
		for i := 0; i < monthsBack; i++ {
			date := time.Now().AddDate(0, -i, 0)
			createdAt := date.Add(time.Duration(r.Intn(24)) * time.Hour)
			invoice := Invoice{
				ID:             uuid.New().String(),
				OrganizationID: org.ID,
				Date:           date,
				Cost:           float64(userCount) * sub.MonthlyPerUserPrice,
				CreatedAt:      createdAt,
				UpdatedAt:      randomTimeAfter(createdAt, r),
			}
			invoices = append(invoices, invoice)
		}
	}
	return invoices
}

func generateProjects(orgs []Organization, users []User, r *rand.Rand) []Project {
	var projects []Project
	for _, org := range orgs {
		numProjects := r.Intn(maxProjectsPerOrg-minProjectsPerOrg+1) + minProjectsPerOrg
		// Get users for this org
		var orgUsers []User
		for _, u := range users {
			if u.OrganizationID == org.ID {
				orgUsers = append(orgUsers, u)
			}
		}

		for i := 0; i < numProjects; i++ {
			createdAt := randomTimeInPast(maxMonthsBack, r)
			project := Project{
				ID:             uuid.New().String(),
				OrganizationID: org.ID,
				OwnerID:        orgUsers[r.Intn(len(orgUsers))].ID,
				Name:           gofakeit.AppName(),
				Description:    gofakeit.Sentence(10),
				CreatedAt:      createdAt,
				UpdatedAt:      randomTimeAfter(createdAt, r),
			}
			projects = append(projects, project)
		}
	}
	return projects
}

func generateTasks(orgs []Organization, projects []Project, users []User, r *rand.Rand) []Task {
	var tasks []Task
	statuses := []string{"todo", "in_progress", "done"}
	statusWeights := []int{50, 15, 35} // 50% todo, 15% in_progress, 35% done

	for _, org := range orgs {
		// Get projects and users for this org
		var orgProjects []Project
		var orgUsers []User
		for _, p := range projects {
			if p.OrganizationID == org.ID {
				orgProjects = append(orgProjects, p)
			}
		}
		for _, u := range users {
			if u.OrganizationID == org.ID {
				orgUsers = append(orgUsers, u)
			}
		}

		for _, project := range orgProjects {
			numTasks := r.Intn(maxTasksPerProject-minTasksPerProject+1) + minTasksPerProject
			for i := 0; i < numTasks; i++ {
				createdAt := randomTimeInPast(maxMonthsBack, r)
				task := Task{
					ID:             uuid.New().String(),
					OrganizationID: org.ID,
					ProjectID:      project.ID,
					AssigneeID:     orgUsers[r.Intn(len(orgUsers))].ID,
					Name:           gofakeit.BS(),
					Description:    gofakeit.Sentence(5),
					Status:         weightedRandomChoice(statuses, statusWeights, r),
					CreatedAt:      createdAt,
					UpdatedAt:      randomTimeAfter(createdAt, r),
				}
				tasks = append(tasks, task)
			}
		}
	}
	return tasks
}

func writeDML(f *os.File, d dialect.Dialect, orgs []Organization, users []User, subs []Subscription, cards []CreditCard, invoices []Invoice, projects []Project, tasks []Task) {
	// Write organizations
	f.WriteString("\n-- Organizations\n")
	for _, org := range orgs {
		f.WriteString(fmt.Sprintf("INSERT INTO organizations (id, name, billing_address, owner_id, created_at, updated_at) VALUES (%s, %s, %s, %s, %s, %s);\n",
			d.FormatString(org.ID), d.FormatString(org.Name), d.FormatString(org.BillingAddress), d.FormatString(org.OwnerID), d.FormatTimestamp(org.CreatedAt), d.FormatTimestamp(org.UpdatedAt)))
	}

	// Write users
	f.WriteString("\n-- Users\n")
	for _, user := range users {
		f.WriteString(fmt.Sprintf("INSERT INTO users (id, organization_id, name, email, password, created_at, updated_at) VALUES (%s, %s, %s, %s, %s, %s, %s);\n",
			d.FormatString(user.ID), d.FormatString(user.OrganizationID), d.FormatString(user.Name), d.FormatString(user.Email), d.FormatString(user.Password), d.FormatTimestamp(user.CreatedAt), d.FormatTimestamp(user.UpdatedAt)))
	}

	// Write subscriptions
	f.WriteString("\n-- Subscriptions\n")
	for _, sub := range subs {
		f.WriteString(fmt.Sprintf("INSERT INTO subscriptions (id, organization_id, plan_id, monthly_per_user_price, created_at, updated_at) VALUES (%s, %s, %s, %.2f, %s, %s);\n",
			d.FormatString(sub.ID), d.FormatString(sub.OrganizationID), d.FormatString(sub.PlanID), sub.MonthlyPerUserPrice, d.FormatTimestamp(sub.CreatedAt), d.FormatTimestamp(sub.UpdatedAt)))
	}

	// Write credit cards
	f.WriteString("\n-- Credit Cards\n")
	for _, card := range cards {
		f.WriteString(fmt.Sprintf("INSERT INTO credit_cards (id, organization_id, number, exp_month, exp_year, cvv, created_at, updated_at) VALUES (%s, %s, %s, %s, %s, %s, %s, %s);\n",
			d.FormatString(card.ID), d.FormatString(card.OrganizationID), d.FormatString(card.Number), d.FormatInt(int64(card.ExpMonth)), d.FormatInt(int64(card.ExpYear)), d.FormatString(card.CVV), d.FormatTimestamp(card.CreatedAt), d.FormatTimestamp(card.UpdatedAt)))
	}

	// Write invoices
	f.WriteString("\n-- Invoices\n")
	for _, invoice := range invoices {
		f.WriteString(fmt.Sprintf("INSERT INTO invoices (id, organization_id, date, cost, created_at, updated_at) VALUES (%s, %s, %s, %.2f, %s, %s);\n",
			d.FormatString(invoice.ID), d.FormatString(invoice.OrganizationID), d.FormatDate(invoice.Date), invoice.Cost, d.FormatTimestamp(invoice.CreatedAt), d.FormatTimestamp(invoice.UpdatedAt)))
	}

	// Write projects
	f.WriteString("\n-- Projects\n")
	for _, project := range projects {
		f.WriteString(fmt.Sprintf("INSERT INTO projects (id, organization_id, owner_id, name, description, created_at, updated_at) VALUES (%s, %s, %s, %s, %s, %s, %s);\n",
			d.FormatString(project.ID), d.FormatString(project.OrganizationID), d.FormatString(project.OwnerID), d.FormatString(project.Name), d.FormatString(project.Description), d.FormatTimestamp(project.CreatedAt), d.FormatTimestamp(project.UpdatedAt)))
	}

	// Write tasks
	f.WriteString("\n-- Tasks\n")
	for _, task := range tasks {
		f.WriteString(fmt.Sprintf("INSERT INTO tasks (id, organization_id, project_id, assignee_id, name, description, status, created_at, updated_at) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s);\n",
			d.FormatString(task.ID), d.FormatString(task.OrganizationID), d.FormatString(task.ProjectID), d.FormatString(task.AssigneeID), d.FormatString(task.Name), d.FormatString(task.Description), d.FormatString(task.Status), d.FormatTimestamp(task.CreatedAt), d.FormatTimestamp(task.UpdatedAt)))
	}
}

// Helper functions

func randomTimeInPast(maxMonthsBack int, r *rand.Rand) time.Time {
	monthsBack := r.Intn(maxMonthsBack)
	return time.Now().AddDate(0, -monthsBack, -r.Intn(30))
}

func randomTimeAfter(t time.Time, r *rand.Rand) time.Time {
	if r.Float32() < 0.5 {
		return t
	}
	return t.Add(time.Duration(r.Intn(30*24)) * time.Hour)
}

func weightedRandomChoice(choices []string, weights []int, r *rand.Rand) string {
	total := 0
	for _, w := range weights {
		total += w
	}
	rnd := r.Intn(total)
	sum := 0
	for i, w := range weights {
		sum += w
		if rnd < sum {
			return choices[i]
		}
	}
	return choices[len(choices)-1]
}

func generateValidCardNumber(r *rand.Rand) string {
	// Generate a random 16-digit number
	digits := make([]int, 16)
	for i := 0; i < 15; i++ {
		digits[i] = r.Intn(10)
	}

	// Calculate check digit using Luhn algorithm
	sum := 0
	for i := 0; i < 15; i++ {
		if i%2 == 0 {
			digits[i] *= 2
			if digits[i] > 9 {
				digits[i] -= 9
			}
		}
		sum += digits[i]
	}
	digits[15] = (10 - (sum % 10)) % 10

	// Convert to string
	var result strings.Builder
	for _, d := range digits {
		result.WriteString(fmt.Sprintf("%d", d))
	}
	return result.String()
}

// generateDeterministicSalt creates a deterministic salt based on the original value
// This matches the implementation in translicator
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

// generatePasswordArgon2id applies Argon2id hashing to the cleartext
// This matches the implementation in translicator exactly
func generatePasswordArgon2id(cleartext string, useSalt bool, time, memory uint32, threads uint8, original string) string {
	var salt []byte
	if useSalt {
		salt = generateDeterministicSalt(original, 16) // 16 bytes salt
	} else {
		salt = make([]byte, 16) // Empty salt
	}

	// Generate hash
	hash := argon2.IDKey([]byte(cleartext), salt, time, memory, threads, 32) // 32 bytes output

	// Format: salt$hash (both hex encoded)
	return fmt.Sprintf("%x$%x", salt, hash)
}
