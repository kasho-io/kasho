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

### Running Tests

**IMPORTANT: Always use MOCK_AUTH=true when running tests**

```bash
# Run all tests with mocked authentication
MOCK_AUTH=true npm test

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
