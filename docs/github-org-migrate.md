# GitHub Organization Migration Guide

This guide walks through migrating the kasho repository from a personal GitHub account to a new GitHub organization.

## Overview

Moving to a GitHub organization provides better collaboration features and professional separation while maintaining all code history. However, several integrations will need reconfiguration.

## Pre-Migration Checklist

### 1. Document Current Configuration

**GitHub Secrets (❗ Critical):**
- [ ] List all repository secrets: `Settings > Secrets and variables > Actions`
- [ ] Document secret names and values (store securely)
- [ ] Current secrets include:
  - `DROPLET_HOST`, `DROPLET_USERNAME`, `DROPLET_SSH_KEY`
  - `KV_URL`, `CHANGE_STREAM_SERVICE`
  - `PRIMARY_DATABASE_URL`, `REPLICA_DATABASE_URL`
  - `PRIMARY_DATABASE_SU_USER/PASSWORD/DB`
  - `REPLICA_DATABASE_SU_USER/PASSWORD/DB`
  - `REPO_URL`

**GitHub Settings:**
- [ ] Note branch protection rules
- [ ] Document any webhooks or integrations
- [ ] Check collaborators and their permissions

**Vercel Projects:**
- [ ] List all connected Vercel projects (demo app, homepage app)
- [ ] Note custom domains for each project
- [ ] Document environment variables for each project

**DigitalOcean:**
- [ ] Note the current `REPO_URL` being used in deployment
- [ ] Verify SSH keys and access

## Migration Steps

### Phase 1: Create GitHub Organization

1. **Create Organization:**
   - Go to GitHub.com → Your profile → "Your organizations" → "New organization"
   - Choose "Create a free organization"
   - Pick organization name and billing email
   - Skip inviting members for now

2. **Organization Settings:**
   - Configure organization profile
   - Set member permissions as needed
   - Consider enabling two-factor authentication requirement

### Phase 2: Transfer Repository

1. **Transfer Repository:**
   - Go to original repo → Settings → scroll to "Transfer ownership"
   - Enter new organization name
   - Type repository name to confirm
   - Click "I understand, transfer this repository"

2. **Verify Transfer:**
   - [ ] Repository appears in new organization
   - [ ] All code, history, and branches transferred
   - [ ] Issues and PRs transferred
   - [ ] GitHub Actions workflows present (but secrets are missing)

### Phase 3: Reconfigure GitHub Actions

1. **Recreate Secrets:**
   - Go to new repo → Settings → Secrets and variables → Actions
   - Add all documented secrets with their values
   - **Update `REPO_URL`** to point to new organization repo

2. **Test Workflows:**
   - [ ] Push a small change to trigger test workflow
   - [ ] Verify all tests pass
   - [ ] Check that deployment workflow can access all secrets

### Phase 4: Update Local Development

1. **Update Git Remote:**
   ```bash
   cd /path/to/kasho
   
   # Check current remote
   git remote -v
   
   # Update origin to new organization
   git remote set-url origin git@github.com:NEW_ORG_NAME/kasho.git
   
   # Verify change
   git remote -v
   
   # Test connection
   git fetch origin
   ```

2. **Update Documentation:**
   - [ ] Update README.md with new repository URLs
   - [ ] Update any other docs referencing old repo URL
   - [ ] Commit and push changes

### Phase 5: Reconfigure Vercel

**For each Vercel project (demo app, homepage app):**

1. **Option A: Reconnect Existing Project**
   - Go to Vercel project → Settings → Git
   - Click "Disconnect" from old repository
   - Click "Connect Git Repository"
   - Select new organization and repository
   - Configure build settings if needed
   - Test auto-deployment

2. **Option B: Import Fresh Project** (if reconnection fails)
   - Go to Vercel dashboard → "Add New" → "Project"
   - Import from new organization repository
   - Configure build settings:
     - **Demo app:** Root directory: `apps/demo`
     - **Homepage app:** Root directory: `apps/homepage`
   - Copy environment variables from old project
   - Update custom domains to point to new project
   - Delete old project after verification

3. **Verify Vercel Integration:**
   - [ ] Test auto-deployment by pushing a change to apps
   - [ ] Verify custom domains still work
   - [ ] Check environment variables are present
   - [ ] Confirm build settings are correct

### Phase 6: Update DigitalOcean Deployment

The deployment will automatically use the new repository URL from the updated `REPO_URL` secret. However, you may need to:

1. **Clean Up DigitalOcean:**
   ```bash
   # SSH into droplet
   ssh user@your-droplet
   
   # Remove old repository and re-clone
   cd ~
   rm -rf kasho
   git clone NEW_REPO_URL kasho
   
   # Rebuild and restart
   cd kasho
   docker build -t kasho --target production .
   cd environments/demo
   docker compose down -v
   docker compose up --build -d
   ```

2. **Verify Deployment:**
   - [ ] Check that services start successfully
   - [ ] Verify auto-deployment works on next push

### Phase 7: Update Branch Protection

1. **Configure Branch Protection:**
   - Go to new repo → Settings → Branches
   - Add rule for `main` branch
   - Configure same protections as before:
     - Require pull request reviews
     - Require status checks (GitHub Actions)
     - Restrict pushes to main

### Phase 8: Clean Up

1. **Old Repository:**
   - Archive or delete old repository (after confirming everything works)
   - Update any external links pointing to old repo

2. **Team Access:**
   - Invite collaborators to new organization
   - Set appropriate permissions

## Post-Migration Verification

### Test Full Pipeline:

1. **Code Changes:**
   - [ ] Make a small change to a service
   - [ ] Push to main branch
   - [ ] Verify GitHub Actions run successfully
   - [ ] Confirm auto-deployment to DigitalOcean works

2. **App Deployments:**
   - [ ] Make a change to demo app
   - [ ] Verify Vercel auto-deployment works
   - [ ] Check that custom domains still work

3. **Infrastructure:**
   - [ ] Verify demo environment is accessible
   - [ ] Test that all services are running
   - [ ] Confirm database replication is working

## Troubleshooting

### Common Issues:

**GitHub Actions failing:**
- Check that all secrets are recreated in new repo
- Verify `REPO_URL` points to new organization
- Ensure SSH keys are still valid

**Vercel not deploying:**
- Reconnect git integration
- Check that webhook URLs are updated
- Verify build settings and environment variables

**DigitalOcean deployment issues:**
- Update `REPO_URL` secret
- May need to re-clone repository on server
- Check SSH keys and permissions

**Local git issues:**
- Update remote URL: `git remote set-url origin NEW_URL`
- May need to re-authenticate with GitHub

## Organization Benefits

After migration, you'll have:
- ✅ Professional separation of personal and project code
- ✅ Better collaboration features
- ✅ Improved permissions management
- ✅ Same free tier benefits (3,000 Actions minutes/month)
- ✅ Ability to add team members easily

## Cost Considerations

- **Free organization** should be sufficient for current usage
- Monitor GitHub Actions minutes usage
- Consider upgrading only if you need more than 3,000 minutes/month

---

**⚠️ Important:** Keep old repository accessible until you've verified everything works in the new organization. The migration can be reversed if needed.