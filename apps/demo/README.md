# kasho demo

The web UI for the Kasho demo. To just try it, see [Try the Demo](https://docs.kasho.io/demo), which needs only Docker.

## Development

Start a development environment and bootstrap it (see [Getting Started](../../README.md#getting-started)), then from the repository root:

```bash
cp apps/demo/.env.example apps/demo/.env.local
task dev:app:demo
```

Open [http://localhost:3001](http://localhost:3001) with your browser to see the result.
