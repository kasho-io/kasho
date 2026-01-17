# Kasho Documentation Site

This is the documentation site for Kasho, built with [Nextra](https://nextra.site/) (Next.js + MDX).

## Development

```bash
# From repo root
task dev:app:docs

# Or directly
cd apps/docs && npm run dev
```

Runs on [http://localhost:3002](http://localhost:3002).

## Special Configuration

### localStorage Polyfill

The `NODE_OPTIONS='--require ./scripts/polyfill-localstorage.js'` in dev/build scripts is required due to a compatibility issue between Node.js v25+ and `@typescript/vfs` (used by Nextra's code highlighting).

Node.js v25 adds experimental localStorage with an implementation that breaks `@typescript/vfs` during SSR. The polyfill provides a working Map-based implementation.

See: https://github.com/shikijs/twoslash/issues/191

### Pagefind Search

After building, the `postbuild` script runs [Pagefind](https://pagefind.app/) to generate a static search index. This enables the search functionality in the docs site without requiring a backend.

The search index is output to `public/_pagefind/`.

## Content Structure

Documentation content lives in `content/`:

```
content/
├── _meta.tsx           # Navigation configuration
├── index.mdx           # Homepage
├── getting-started/    # Getting started guides
├── configuration/      # Configuration reference
└── ...
```

See the [Nextra documentation](https://nextra.site/docs) for content authoring guidelines.
