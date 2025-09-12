# Kasho Console

This is the Kasho administrative console app.

## Getting Started

First, run the development server:

```bash
task dev:app:console
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

## Testing

The console app uses Playwright for end-to-end testing. Tests require authentication to be mocked to avoid needing real WorkOS credentials.

### Test Scripts

- `test:ci` - Runs linting, type checking, and formatting checks (for CI/CD)
- `test:e2e` - Runs Playwright end-to-end tests (requires `MOCK_AUTH=true`)
- `test` - Alias for running Playwright tests
- `test:headed` - Runs tests with visible browser
- `test:ui` - Runs tests in interactive UI mode

### Running Tests

**IMPORTANT: Always use MOCK_AUTH=true when running Playwright tests**

```bash
# Run CI checks (linting, type checking, formatting)
npm run test:ci

# Run end-to-end tests with mocked authentication
MOCK_AUTH=true npm run test:e2e

# Run specific tests
MOCK_AUTH=true npm test -- --grep "profile"

# Run tests in headed mode (see browser)
MOCK_AUTH=true npm test:headed

# Run tests with UI mode
MOCK_AUTH=true npm test:ui
```

### Running Tests in Development Mode

If you need to run the app in development mode with mocked auth:

```bash
MOCK_AUTH=true npm run dev
```

## Environment Variables

- `MOCK_AUTH=true` - Mocks WorkOS authentication for testing (required for tests)
- `BLOB_READ_WRITE_TOKEN` - Vercel Blob storage token for profile picture uploads
