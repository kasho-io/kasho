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

### Phase 3: User Profile Management

#### 3.1 Overview

Implement comprehensive user profile management features allowing users to update their personal information, including email addresses, profile pictures, and custom metadata. Since WorkOS doesn't provide pre-built UI components for profile management, we'll build custom forms that interact with the WorkOS API directly.

#### 3.2 Technical Architecture

##### 3.2.1 Dependencies
- `@workos-inc/node` - WorkOS Node SDK for server-side API calls
- `@workos-inc/authkit-nextjs` - Already installed for authentication
- `@vercel/blob` - Vercel Blob Storage for profile picture uploads
- Next.js API routes for secure server-side operations
- DaisyUI components for consistent UI

##### 3.2.2 API Integration Pattern
```typescript
// Server-side only operations
WorkOS Client → API Route → WorkOS API → Update Session → UI Update
```

##### 3.2.3 Data Storage Strategy
- **Profile Pictures**: 
  - Files stored in Vercel Blob Storage
  - URLs saved in WorkOS user metadata
  - Served through Vercel's global CDN
  - Works in both local development and production (requires BLOB_READ_WRITE_TOKEN)
- **User Preferences**: Store as JSON in metadata
- **Email Updates**: Direct field update via WorkOS API
- **Metadata Limits**: 10 key-value pairs, 40 char keys, structured data as JSON strings

##### 3.2.4 Vercel Blob Storage Configuration
- **Local Development**: Requires Vercel project and BLOB_READ_WRITE_TOKEN
- **Production**: Automatically configured in Vercel deployments
- **Storage**: Files uploaded to Vercel's servers, not stored locally
- **CDN**: Automatic edge caching for fast global delivery
- **Costs**: Pay-per-use pricing (1GB free tier for storage and bandwidth)

#### 3.3 Implementation Components

##### 3.3.1 WorkOS Client Setup (`/lib/workos-client.ts`)
```typescript
import { WorkOS } from '@workos-inc/node';

export const workosClient = new WorkOS(process.env.WORKOS_API_KEY);
```

##### 3.3.2 Profile Management Page (`/app/account/profile/page.tsx`)
**Features:**
- Display current user information
- Edit form for email address
- Profile picture upload/URL input
- Custom fields for additional metadata
- Save/Cancel actions with optimistic updates

**UI Components:**
- DaisyUI card for profile section
- Form inputs with validation
- Avatar preview component
- Loading states and error handling
- Success notifications

##### 3.3.3 API Routes

**Update User Profile (`/app/api/user/update/route.ts`)**
- POST endpoint for profile updates
- Validates session before allowing updates
- Calls WorkOS Update User API
- Returns updated user data
- Handles partial updates

**Upload Profile Picture (`/app/api/user/upload-avatar/route.ts`)**
- POST endpoint for image uploads
- Validates file type and size (max 4.5MB per Vercel Blob limits)
- Uploads to Vercel Blob Storage using `@vercel/blob`
- Returns permanent CDN URL
- Updates user metadata with picture URL
- Example implementation:
  ```typescript
  import { put } from '@vercel/blob';
  
  const blob = await put(filename, file, {
    access: 'public',
  });
  // Returns URL like: https://[...].public.blob.vercel-storage.com/[...]
  ```

##### 3.3.4 Type Definitions (`/types/user.ts`)
```typescript
interface UserProfile {
  id: string;
  email: string;
  firstName?: string;
  lastName?: string;
  profilePictureUrl?: string;
  metadata?: Record<string, any>;
}

interface UpdateUserRequest {
  email?: string;
  firstName?: string;
  lastName?: string;
  metadata?: Record<string, any>;
}
```

#### 3.4 User Flow

##### 3.4.1 Viewing Profile
1. User navigates to `/account/profile`
2. Page fetches current user data from session
3. Displays current information in read-only format
4. Shows "Edit" button to enable form

##### 3.4.2 Editing Profile
1. User clicks "Edit" button
2. Form becomes editable with current values pre-filled
3. User modifies desired fields
4. Client-side validation provides immediate feedback
5. User clicks "Save" or "Cancel"

##### 3.4.3 Saving Changes
1. Form submission triggers API call to `/api/user/update`
2. API route validates session and input
3. Calls WorkOS Update User API with partial update
4. On success: Updates local session, shows success message
5. On failure: Shows error message, preserves form state

##### 3.4.4 Profile Picture Update
1. User clicks "Change Picture" button
2. File input limited to image types (jpg, png, gif, webp)
3. File sent to `/api/user/upload-avatar`
4. API route uploads to Vercel Blob Storage
5. Blob Storage returns permanent CDN URL
6. Preview shown using the CDN URL
7. URL saved in WorkOS metadata as `profile_picture_url`
8. Old image optionally deleted from Blob Storage

#### 3.5 WorkOS API Integration Details

##### 3.5.1 Update User Endpoint
- **Method**: PATCH (partial update pattern)
- **Endpoint**: `/users/{userId}`
- **Capabilities**:
  - Update email (with re-verification if needed)
  - Update metadata (partial, preserves existing)
  - Update external_id for system integration

##### 3.5.2 Metadata Structure
```json
{
  "metadata": {
    "profile_picture_url": "https://...",
    "first_name": "John",
    "last_name": "Doe",
    "preferences": "{\"theme\":\"dark\",\"notifications\":true}",
    "bio": "Software engineer",
    "location": "San Francisco, CA"
  }
}
```

##### 3.5.3 Session Synchronization
- After successful update, refresh the session
- Use `refreshUser` from AuthKit to get latest data
- Update UI components that display user info

#### 3.6 Security Considerations

##### 3.6.1 Authentication
- All profile routes require authenticated session
- Use `withAuth` for server components
- Validate session in API routes before processing

##### 3.6.2 Input Validation
- Sanitize all user inputs
- Validate email format
- Check file types and sizes for uploads
- Prevent XSS in metadata fields

##### 3.6.3 Rate Limiting
- Implement rate limiting on update endpoints
- Prevent rapid successive updates
- Consider using middleware for rate limiting

#### 3.7 UI/UX Guidelines

##### 3.7.1 Form Design
- Use DaisyUI form components
- Clear labels and help text
- Inline validation messages
- Disabled state while saving
- Clear success/error feedback

##### 3.7.2 Responsive Design
- Mobile-first approach
- Stack form fields on small screens
- Appropriate touch targets
- Consider drawer pattern for mobile

##### 3.7.3 Accessibility
- Proper ARIA labels
- Keyboard navigation support
- Screen reader friendly
- Focus management

#### 3.8 Error Handling

##### 3.8.1 Client-Side
- Form validation before submission
- Network error recovery
- Optimistic updates with rollback
- Clear error messages

##### 3.8.2 Server-Side
- Try-catch blocks in API routes
- Proper HTTP status codes
- Detailed error logging
- User-friendly error messages

#### 3.9 Testing Approach

##### 3.9.1 Unit Tests
- Form validation logic
- API route handlers
- Utility functions

##### 3.9.2 Integration Tests
- Full update flow
- Session synchronization
- Error scenarios

##### 3.9.3 E2E Tests
- Complete user journey
- Profile picture upload
- Email change flow

#### 3.10 Future Enhancements

##### 3.10.1 Phase 3.5 (Optional)
- Email verification for email changes
- Two-factor authentication setup
- Password change (if using password auth)
- Account deletion

##### 3.10.2 Phase 3.6 (Optional)
- Social account linking
- Multiple profile pictures/gallery
- Rich text bio editor
- Privacy settings

#### 3.11 Navigation Component Updates

The navigation component (implemented in Phase 2) needs the following updates for Phase 3:

**Current Implementation:**
- User avatar with initial
- Dropdown menu with email display
- Sign out functionality
- Responsive design with DaisyUI components

**Phase 3 Updates Required:**
1. **Add "Profile" menu item** - Link to `/account/profile` in the dropdown
2. **Update avatar to show profile picture** - Display user's uploaded image instead of initial when `profile_picture_url` exists in metadata
3. **Menu structure:**
   ```
   Dropdown Menu:
   - [User Email] (header)
   - Profile (link to /account/profile)
   - ─────────── (divider)
   - Sign Out
   ```

**Future enhancements (Phase 4+):**
- Add "Settings" link when settings page is created
- Add organization switcher if multi-org support is added

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