# Claude Code Preferences

## Project Overview
@README.md
@AGENTS.md

## GitHub/Git Workflow Preferences
- **Feature Branches**: Always work on feature branches for large changes (new services, tools, multi-file refactors). Never commit large changes directly to main.
- **PR Workflow**:
  1. Create a feature branch (e.g., `feature/mysql-support`, `feature/kasho-md7-bootstrap-sync`)
  2. Commit work to the branch
  3. Create a PR with comprehensive summary of changes
  4. **Wait for human review** - do not merge PRs automatically. The user will review all code changes before approval.
  5. After approval, merge with `--rebase` (not `--squash`)
- **Branch Cleanup**: After merging PRs, automatically clean up local branches with `git branch -d <branch-name>`
- **PR Creation**: Include comprehensive summary of changes and note existing test coverage rather than generic test plans

## Development Preferences
- Follow existing code conventions and patterns in the codebase
- Check for existing libraries/frameworks before introducing new ones
- Run lint and typecheck commands after making changes (if available)
- Use idiomatic patterns for the languages and frameworks we are using
- **IMPORTANT**: Always run Prettier and ESLint before committing changes to apps/* directories:
  - For homepage app: `cd apps/homepage && npm run prettier:write && npm run lint`
  - For demo app: `cd apps/demo && npm run prettier:write && npm run lint`
  - For docs app: `cd apps/docs && npm run prettier:write && npm run lint`
  - Note: Git pre-commit hooks will automatically run these checks, but running them manually helps catch issues earlier

## Testing
- The services in the project have comprehensive test coverage - acknowledge existing tests rather than creating generic test plans
- Service test files are located in `*_test.go` files following Go conventions

## Documentation
- As the codebase evolves, review the documentation in apps/docs to ensure that none of it is out of date, and if it is out of date, update it.
- The audience for the documenation are customers of Kasho. These will typically be SRE / Infra / Platform engineers who are technical, understand Docker, containers, and DevOps.

## Releasing New Versions

When cutting a new release, **always** update versions in these locations:

1. **Go services**: Update `pkg/version/version.go` - change the `Version` variable
2. **Apps**: Update the top-level `"version"` field in each package.json:
   - `apps/demo/package.json`
   - `apps/homepage/package.json`
   - `apps/docs/package.json`

   **IMPORTANT**: Only edit the `"version"` field at the top of each file. Do NOT use sed/regex replacement that might accidentally change dependency versions.

3. **Git tag**: After committing version updates, tag and push:
   ```bash
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```

### Version Numbering (Semantic Versioning)
- **Breaking changes**: Bump minor version (0.X.0) while pre-1.0
- **New features**: Bump minor version (0.X.0)
- **Bug fixes only**: Bump patch version (0.0.X)
