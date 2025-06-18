# Transform Configuration Guide

pg-translicator requires a `transforms.yml` file mounted at `/app/config/transforms.yml`. This file defines table and column transformations.

## transforms.yml File Format

The `transforms.yml` file controls how data is transformed during replication. It uses YAML format with the following structure:

**Basic Structure:**
```yaml
version: v1
tables:
  schema.table_name:
    column_name: TransformationType
    another_column: AnotherTransformationType
```

**Simple vs Object Notation:**

You can use either simple string format or object notation:

```yaml
version: v1
tables:
  public.users:
    # Simple string transforms (shorthand format)
    name: FakeName
    email: FakeEmail
    
    # Object notation (equivalent to above)
    company:
      type: FakeCompany
    
    # Regex transforms require object notation
    phone:
      type: Regex
      pattern: '\(?\d{3}\)?[-.\\s]?\d{3}[-.\\s]?\d{4}'
      replacement: '(XXX) XXX-XXXX'
```

## Available Transform Types

**Personal Information (Gofakeit-based):**
- `FakeName` - Full name generation
- `FakeFirstName`, `FakeLastName` - Individual name components
- `FakeEmail` - Email address generation
- `FakePhone` - Phone number generation
- `FakeSSN` - Social Security Number (XXX-XX-XXXX format)
- `FakeDateOfBirth` - Date of birth (YYYY-MM-DD format)
- `FakeUsername`, `FakePassword` - Account credentials

**Address Information (Gofakeit-based):**
- `FakeStreetAddress` - Full street address
- `FakeCity`, `FakeState`, `FakeStateAbbr` - Location components
- `FakeZip` - ZIP codes (XXXXX or XXXXX-XXXX format)
- `FakeCountry` - Country names
- `FakeLatitude`, `FakeLongitude` - Geographic coordinates

**Business Information (Gofakeit-based):**
- `FakeCompany` - Company names
- `FakeJobTitle` - Job/position titles
- `FakeIndustry` - Industry names
- `FakeProduct`, `FakeProductName` - Product information

**Text and Content (Gofakeit-based):**
- `FakeParagraph`, `FakeSentence`, `FakeWord` - Text generation
- `FakeCharacters`, `FakeDigits` - String generation

**Financial Information (Gofakeit-based):**
- `FakeCreditCardType`, `FakeCreditCardNum` - Financial data
- `FakeCurrency` - Currency codes

**Date and Time (Gofakeit-based):**
- `FakeMonth`, `FakeMonthNum`, `FakeWeekDay`, `FakeYear` - Date/time components

**Custom Transforms:**
- `Bool` - Boolean values (deterministic custom implementation)

**Pattern-Based Transforms:**
- `Regex` - Apply custom regular expression patterns and replacements

## Regex Transform Details

The Regex transform allows custom pattern-based data transformation:

```yaml
column_name:
  type: Regex
  pattern: 'regex_pattern'
  replacement: 'replacement_string'
```

**Features:**
- Uses Go's RE2 regex syntax (safe subset of Perl regex)
- Supports capture groups with `$1`, `$2`, etc. in replacements
- No lookahead/lookbehind assertions
- Linear time complexity guaranteed

**Examples:**
```yaml
# Phone number standardization
phone:
  type: Regex
  pattern: '\+?1?\s*\(?\(\d{3}\)\)?[-.\s]*(\d{3})[-.\s]*(\d{4})'
  replacement: '+1 (XXX) XXX-XXXX'

# SSN partial masking (keep last 4 digits)
ssn:
  type: Regex
  pattern: '(\d{3})-(\d{2})-(\d{4})'
  replacement: 'XXX-XX-$3'

# IP address masking
ip_address:
  type: Regex
  pattern: '\d+\.\d+\.\d+\.\d+'
  replacement: 'XXX.XXX.XXX.XXX'

# Credit card partial masking (keep last 4 digits)
card_number:
  type: Regex
  pattern: '(\d{4})[\s-]?(\d{4})[\s-]?(\d{4})[\s-]?(\d{4})'
  replacement: 'XXXX-XXXX-XXXX-$4'
```

## Key Features

1. **Deterministic Transformations**: The same input always produces the same output, ensuring data consistency
2. **Type-Safe**: Transforms are validated against column data types
3. **Selective Processing**: Only specified tables/columns are transformed
4. **Referential Integrity**: Consistent transformations preserve relationships

## Configuration Guidelines

**Creating Your transforms.yml:**

1. **Start Simple**: Begin with a minimal configuration and add tables gradually
   ```yaml
   version: v1
   tables:
     public.users:
       email: FakeEmail
   ```

2. **Identify Sensitive Data**: Focus on columns containing:
   - Personal identifiers (names, emails, phone numbers)
   - Addresses and location data
   - Financial information
   - Any data subject to privacy regulations

3. **Test Transformations**: Verify transforms work with your data types:
   - String columns → String transforms (Name, Email, etc.)
   - Integer columns → Integer transforms (Year, MonthNum, etc.)
   - Boolean columns → Bool transform

4. **Consider Relationships**: Use consistent transforms for related data:
   ```yaml
   public.users:
     email: FakeEmail
   public.user_profiles:
     user_email: FakeEmail  # Same transform maintains relationship
   ```

## Example Configurations

**E-commerce Example:**
```yaml
version: v1
tables:
  public.customers:
    first_name: FakeFirstName
    last_name: FakeLastName
    email: FakeEmail
    phone: FakePhone
    street_address: FakeStreetAddress
    city: FakeCity
    state: FakeStateAbbr
    zip_code: FakeZip
  public.orders:
    customer_email: FakeEmail
    billing_address: FakeStreetAddress
  public.payments:
    cardholder_name: FakeName
    card_number: FakeCreditCardNum
```

**Comprehensive Example with Regex:**
```yaml
version: v1
tables:
  public.users:
    # Simple string transforms (shorthand format)
    name: FakeName
    email: FakeEmail
    
    # Object notation (equivalent to above)
    company:
      type: FakeCompany
    
    # Regex transforms require object notation
    phone:
      type: Regex
      pattern: '\(?\(\d{3}\)\)?[-.\\s]?\(\d{3}\)[-.\\s]?\(\d{4}\)'
      replacement: '(XXX) XXX-XXXX'
    
    ssn:
      type: Regex
      pattern: '(\d{3})-(\d{2})-(\d{4})'
      replacement: 'XXX-XX-$3'  # Preserves last 4 digits
    
    ip_address:
      type: Regex
      pattern: '\d+\.\d+\.\d+\.\d+'
      replacement: 'XXX.XXX.XXX.XXX'
  
  public.credit_cards:
    cardholder_name: FakeName
    
    # Partial masking - keeps last 4 digits visible
    card_number:
      type: Regex
      pattern: '(\d{4})[\s-]?(\d{4})[\s-]?(\d{4})[\s-]?(\d{4})'
      replacement: 'XXXX-XXXX-XXXX-$4'
    
    cvv:
      type: Regex
      pattern: '\d+'
      replacement: 'XXX'
```

## Validation

- Missing `/app/config/transforms.yml` → Service fails to start
- Invalid YAML syntax → Parsing error at startup
- Unknown transform types → Runtime error during processing
- Type mismatches → Processing error for affected columns

## Troubleshooting

**"Required config file /app/config/transforms.yml not found"**
- Mount the config directory with transforms.yml to /app/config

**Transform errors during processing**
- Verify transform types match column data types
- Check YAML syntax is valid
- Ensure all referenced tables exist in your database