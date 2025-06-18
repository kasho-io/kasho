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
    name: Name
    email: Email
    
    # Object notation (equivalent to above)
    company:
      type: Company
    
    # Regex transforms require object notation
    phone:
      type: Regex
      pattern: '\(?\d{3}\)?[-.\\s]?\d{3}[-.\\s]?\d{4}'
      replacement: '(XXX) XXX-XXXX'
```

## Available Transform Types

**Personal Information:**
- `Name` - Full name generation
- `FirstName`, `LastName` - Individual name components
- `Email` - Email address generation
- `Phone` - Phone number generation
- `SSN` - Social Security Number (XXX-XX-XXXX format)
- `DateOfBirth` - Date of birth (YYYY-MM-DD format)
- `Username`, `Password` - Account credentials

**Address Information:**
- `StreetAddress` - Full street address
- `City`, `State`, `StateAbbr` - Location components
- `Zip` - ZIP codes (XXXXX or XXXXX-XXXX format)
- `Country` - Country names
- `Latitude`, `Longitude` - Geographic coordinates

**Business Information:**
- `Company` - Company names
- `JobTitle` - Job/position titles
- `Industry` - Industry names
- `Product`, `ProductName` - Product information

**Text and Content:**
- `Paragraph`, `Sentence`, `Word` - Text generation
- `Characters`, `Digits` - String generation

**Other Types:**
- `Bool` - Boolean values
- `CreditCardType`, `CreditCardNum` - Financial data
- `Currency` - Currency codes

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
       email: Email
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
     email: Email
   public.user_profiles:
     user_email: Email  # Same transform maintains relationship
   ```

## Example Configurations

**E-commerce Example:**
```yaml
version: v1
tables:
  public.customers:
    first_name: FirstName
    last_name: LastName
    email: Email
    phone: Phone
    street_address: StreetAddress
    city: City
    state: StateAbbr
    zip_code: Zip
  public.orders:
    customer_email: Email
    billing_address: StreetAddress
  public.payments:
    cardholder_name: Name
    card_number: CreditCardNum
```

**Comprehensive Example with Regex:**
```yaml
version: v1
tables:
  public.users:
    # Simple string transforms (shorthand format)
    name: Name
    email: Email
    
    # Object notation (equivalent to above)
    company:
      type: Company
    
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
    cardholder_name: Name
    
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