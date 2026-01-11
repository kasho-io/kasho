# Kasho Console

This is the Kasho administrative console app.

## Getting Started

First, run the development server:

```bash
task dev:app:console
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

## Testing

The console app uses Playwright for end-to-end testing with mocked authentication configured automatically.

### Test Scripts

- `test:ci` - Runs linting, type checking, and formatting checks (for CI/CD)
- `test:e2e` - Runs Playwright end-to-end tests
- `test` - Alias for running Playwright tests
- `test:headed` - Runs tests with visible browser
- `test:ui` - Runs tests in interactive UI mode

### Running Tests

```bash
# Run CI checks (linting, type checking, formatting)
npm run test:ci

# Run end-to-end tests
npm run test:e2e

# Run specific tests
npm test -- --grep "profile"

# Run tests in headed mode (see browser)
npm run test:headed

# Run tests with UI mode
npm run test:ui
```

**Note**: Authentication mocking is automatically configured in `playwright.config.ts`, so you don't need to set any environment variables.

### Running Tests in Development Mode

If you need to run the app in development mode with mocked auth:

```bash
MOCK_AUTH=true npm run dev
```

**Note**: This is only needed for manual development testing. Playwright tests automatically configure mocked auth.

## Environment Variables

- `MOCK_AUTH=true` - Mocks WorkOS authentication (only needed for manual testing, Playwright sets this automatically)
- `BLOB_READ_WRITE_TOKEN` - Vercel Blob storage token for profile picture uploads
