# Console App Authentication Implementation Plan

## Overview

This document outlines the implementation plan for adding WorkOS AuthKit authentication and user management to the Kasho console application (currently named "homepage"). The integration will provide enterprise-grade authentication with SSO capabilities, user session management, and protected routes for administrative functions including billing and user management.

## Current State

- **Framework**: Next.js 15.3.2 with React 19
- **Styling**: TailwindCSS + DaisyUI
- **Location**: `/apps/homepage/` (to be renamed to `/apps/console/`)
- **Current Features**: Public landing page with Kasho wordmark and description
- **Port**: Runs on localhost:3000 in development

## Architecture Decision

### SDK Choice: `@workos-inc/authkit-nextjs`

We'll use the Next.js-specific SDK as it provides:
- Built-in middleware support for Next.js 15
- Server and client component compatibility
- Session management with encrypted cookies
- TypeScript support out of the box
- App Router compatibility

## Implementation Phases

### Phase 0: Application Rename

#### 0.1 Rename Directory
- Rename `/apps/homepage/` to `/apps/console/`
- Update all internal references to the new name

#### 0.2 Update Configuration
- Update `package.json` name field to `@kasho/console`
- Update any scripts in root `package.json` that reference the homepage app
- Update Taskfile.yml port mapping for console app (optional - will default to port 3000)

#### 0.3 Update Documentation
- Update README.md references
- Update CLAUDE.md references
- Update any other documentation mentioning the homepage app
- Update environment configuration files

#### 0.4 Update Git References
- Ensure git properly tracks the rename
- Update any CI/CD pipelines if they reference the homepage path

### Phase 1: Infrastructure Setup

#### 1.1 Install Dependencies
```bash
npm install @workos-inc/authkit-nextjs
```

#### 1.2 Environment Configuration
Create `.env` with:
```
# WorkOS API Keys
WORKOS_API_KEY='sk_...'
WORKOS_CLIENT_ID='client_...'
WORKOS_COOKIE_PASSWORD='<32+ char secure password>'

# Redirect URIs (configured in WorkOS Dashboard)
NEXT_PUBLIC_WORKOS_REDIRECT_URI='http://localhost:3000/callback'
WORKOS_LOGIN_URI='http://localhost:3000/login'
WORKOS_LOGOUT_REDIRECT_URI='http://localhost:3000'
```

#### 1.3 WorkOS Dashboard Configuration
- Enable AuthKit in WorkOS Dashboard
- Configure redirect URIs:
  - Default redirect: `http://localhost:3000/callback`
  - Login endpoint: `http://localhost:3000/login`
  - Logout redirect: `http://localhost:3000`
- Production URIs will follow pattern: `https://kasho.io/...`

### Phase 2: Core Authentication Implementation

#### 2.1 AuthKit Provider Setup
- Wrap app layout with `AuthKitProvider` in `/app/layout.tsx`
- Configure provider with environment variables

#### 2.2 Middleware Configuration
Create `/middleware.ts` with two possible approaches:
- **Option A**: Page-based auth (recommended initially)
  - Protected routes determined by `withAuth` usage
  - More flexible for mixed public/private content
- **Option B**: Middleware auth
  - All routes protected by default
  - Explicit allowlist for public routes

Initial recommendation: **Page-based auth** for flexibility

#### 2.3 Authentication Routes

##### `/app/callback/route.ts`
- Handle OAuth callback from WorkOS
- Exchange authorization code for user session
- Set encrypted session cookie
- Redirect to intended destination or dashboard

##### `/app/login/route.ts`
- Generate AuthKit authorization URL
- Support `returnTo` parameter for post-login redirect
- Handle PKCE flow initialization

##### `/app/logout/route.ts`
- Clear session cookie
- Redirect to WorkOS logout URL
- Return user to homepage after logout

### Phase 3: User Interface Components

#### 3.1 Navigation Updates

Since there is no existing navigation, we'll build one from scratch using DaisyUI components and WorkOS AuthKit for authentication.

**Navigation Component Structure:**
- Create a `Navigation.tsx` component that will:
  - Display at the top of all pages
  - Show "Sign In" button for unauthenticated users
  - Show user dropdown menu for authenticated users with:
    - User email/name display
    - Account settings link (future feature)
    - Sign Out option

**Authentication Integration:**
Using WorkOS AuthKit for Next.js:
- Install `@workos-inc/authkit-nextjs`
- Set up AuthKitProvider in the layout
- Configure middleware for session management
- Create callback and login routes

**Component Hierarchy:**
```
layout.tsx (with AuthKitProvider)
  └── Navigation.tsx (uses useAuth hook)
       ├── Sign In button (unauthenticated)
       └── User dropdown (authenticated)
            ├── User info display
            └── Sign Out link
```

**Styling Approach:**
- Use DaisyUI's navbar and dropdown components
- Maintain consistent theming with existing dark/light mode support
- Responsive design for mobile/desktop

**Key Implementation Details:**
- Navigation will be a client component using the `useAuth` hook
- Sign In redirects to WorkOS hosted auth page
- Sign Out clears session and redirects to homepage
- User info fetched from WorkOS session

#### 3.2 User Dashboard (`/app/dashboard`)
Create authenticated user area:
- User profile information
- Account settings
- Organization management (if applicable)

#### 3.3 Protected Content Areas
- `/app/account/*` - User account management
- `/app/settings/*` - Application settings
- `/app/admin/*` - Admin panel (role-based)

### Phase 4: Session Management

#### 4.1 Server Components
Use `withAuth` for server-side auth:
```typescript
const { user } = await withAuth({ ensureSignedIn: true });
```

#### 4.2 Client Components
Use `useAuth` hook for client-side:
```typescript
const { user, loading, signOut } = useAuth();
```

#### 4.3 Session Refresh
- Automatic token refresh handled by SDK
- Configure session timeout preferences
- Implement "Remember me" functionality

### Phase 5: Advanced Features

#### 5.1 Role-Based Access Control (RBAC)
- Define user roles (admin, user, viewer)
- Implement role checking middleware
- Create role-specific UI components

#### 5.2 Organization Management
- Multi-tenant support
- Organization switching
- Invitation flows

#### 5.3 User Management Admin Panel
- User list with search/filter
- User details and activity logs
- Manual user creation/deletion
- Role assignment interface

### Phase 6: Security Considerations

#### 6.1 CSRF Protection
- Leverage Next.js built-in CSRF protection
- Validate state parameter in OAuth flow

#### 6.2 Session Security
- HTTP-only cookies for session storage
- Secure flag for HTTPS environments
- SameSite attribute configuration

#### 6.3 Rate Limiting
- Implement rate limiting on auth endpoints
- Consider using Vercel Edge Functions rate limiting

## File Structure

```
apps/console/
├── .env                          # Environment variables
├── middleware.ts                 # Auth middleware
├── app/
│   ├── layout.tsx               # AuthKitProvider wrapper
│   ├── page.tsx                 # Public homepage
│   ├── login/
│   │   └── route.ts            # Login endpoint
│   ├── logout/
│   │   └── route.ts            # Logout endpoint
│   ├── callback/
│   │   └── route.ts            # OAuth callback
│   ├── dashboard/
│   │   ├── layout.tsx          # Protected layout
│   │   └── page.tsx            # User dashboard
│   ├── account/
│   │   ├── page.tsx            # Account settings
│   │   └── profile/
│   │       └── page.tsx        # User profile
│   └── admin/
│       ├── layout.tsx          # Admin-only layout
│       └── users/
│           └── page.tsx        # User management
├── components/
│   ├── auth/
│   │   ├── LoginButton.tsx     # Sign in CTA
│   │   ├── UserMenu.tsx        # User dropdown
│   │   └── ProtectedRoute.tsx  # Route guard
│   └── layout/
│       └── Navigation.tsx      # Updated nav with auth
└── lib/
    ├── auth.ts                 # Auth utilities
    └── workos.ts              # WorkOS client config
```

## Testing Strategy

### Development Testing
1. Create test WorkOS account
2. Test login flow with email/password
3. Test SSO with Google/GitHub
4. Verify session persistence
5. Test logout flow
6. Validate protected route access

### Integration Testing
- Test session refresh mechanism
- Validate CORS configuration
- Test redirect URI handling
- Verify error states

### User Acceptance Testing
- Multi-browser testing
- Mobile responsiveness
- SSO provider testing
- Session timeout scenarios

## Deployment Considerations

### Environment-Specific Config
- Development: localhost:3000
- Staging: staging.kasho.io
- Production: kasho.io

### Environment Variables
- Use Vercel environment variables
- Separate WorkOS environments (dev/staging/prod)
- Rotate cookie passwords periodically

### Monitoring
- Track authentication metrics
- Monitor failed login attempts
- Set up alerts for auth service issues

## Migration Path

### Phase 0: Application Rename (Day 1)
- Rename homepage to console
- Update all references
- Verify development environment

### Phase 1: Basic Auth (Week 1)
- Core authentication flow
- Session management
- Basic protected routes

### Phase 2: Enhanced UI (Week 2)
- User dashboard
- Account settings
- Improved navigation

### Phase 3: Advanced Features (Week 3)
- RBAC implementation
- Organization support
- Admin panel

### Phase 4: Production Ready (Week 4)
- Performance optimization
- Security audit
- Documentation
- Deployment

## Success Criteria

- [ ] Users can sign up/sign in via WorkOS
- [ ] Sessions persist across page refreshes
- [ ] Protected routes redirect unauthenticated users
- [ ] Users can sign out successfully
- [ ] SSO providers work correctly
- [ ] Session refresh works seamlessly
- [ ] Error states handled gracefully
- [ ] Mobile-responsive auth flows
- [ ] Compliant with security best practices
- [ ] Performance metrics meet targets (<200ms auth checks)

## Related Documentation

- [WorkOS AuthKit Quick Start](./workos-auth.md)
- [Next.js Authentication Best Practices](https://nextjs.org/docs/app/building-your-application/authentication)
- [WorkOS Dashboard](https://dashboard.workos.com)

## Open Questions

1. Do we need organization/team support initially?
2. Should we implement social login providers (Google, GitHub)?
3. What user metadata should we store locally vs. in WorkOS?
4. Do we need audit logging for authentication events?
5. Should we implement 2FA/MFA support?
6. What's the session timeout strategy?
7. Do we need impersonation features for support?

## Next Steps

1. Review and approve implementation plan
2. Set up WorkOS account and configure AuthKit
3. Generate secure cookie password
4. Begin Phase 1 implementation
5. Schedule weekly progress reviews