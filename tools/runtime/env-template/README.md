# env-template

A simple tool to process template files using environment variables.

## Usage

```bash
go run main.go --dirs "dir1,dir2,dir3"
```

The tool will:
1. Look for all files ending in `.template` in the specified directories
2. Process each template file using environment variables
3. Create a new file without the `.template` extension containing the processed content

## Example

Given a template file `example.sql.template`:
```sql
CREATE ROLE {{.PRIMARY_DATABASE_KASHO_USER}} WITH PASSWORD '{{.PRIMARY_DATABASE_KASHO_PASSWORD}}';
```

And environment variables:
```bash
PRIMARY_DATABASE_KASHO_USER=kasho
PRIMARY_DATABASE_KASHO_PASSWORD=secret
```

Running:
```bash
go run main.go --dirs "environments/development/primary-init.d"
```

Will create `example.sql`:
```sql
CREATE ROLE kasho WITH PASSWORD 'secret';
```

## Building

```bash
go build -o env-template
```

Then use it:
```bash
./env-template --dirs "dir1,dir2,dir3"
``` 