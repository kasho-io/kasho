# syntax=docker/dockerfile:1.7
# Kasho Demo App
#
# Multi-stage build for the apps/demo Next.js workspace. Produces a small
# runtime image based on Next.js standalone output. The build context is the
# repository root because the demo participates in an npm workspace.

ARG NODE_VERSION=22-alpine

# ----- deps: install workspace dependencies -----
FROM node:${NODE_VERSION} AS deps
WORKDIR /repo

# argon2 is a native module; compiling it requires python3 + build tools
RUN apk add --no-cache python3 make g++

# Copy workspace manifests first so this layer caches on lockfile-only changes
COPY package.json package-lock.json ./
COPY apps/demo/package.json ./apps/demo/

# Install only what the demo workspace needs, including dev deps for the build
RUN npm ci --workspace=apps/demo --include-workspace-root

# ----- builder: build the demo workspace -----
FROM node:${NODE_VERSION} AS builder
WORKDIR /repo

COPY --from=deps /repo/node_modules ./node_modules
COPY --from=deps /repo/apps/demo/node_modules ./apps/demo/node_modules
COPY package.json package-lock.json ./
COPY apps/demo ./apps/demo

ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build --workspace=apps/demo

# ----- runner: minimal runtime image -----
FROM node:${NODE_VERSION} AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3001
ENV HOSTNAME=0.0.0.0

RUN addgroup -g 1001 -S nodejs && \
    adduser -S nextjs -u 1001 -G nodejs

# Next.js standalone bundle: copies the minimal node_modules subset it needs
COPY --from=builder --chown=nextjs:nodejs /repo/apps/demo/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /repo/apps/demo/.next/static ./apps/demo/.next/static
COPY --from=builder --chown=nextjs:nodejs /repo/apps/demo/public ./apps/demo/public

USER nextjs

EXPOSE 3001

# The standalone server entrypoint preserves the workspace path layout
CMD ["node", "apps/demo/server.js"]
