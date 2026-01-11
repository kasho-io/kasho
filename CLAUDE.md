# Claude Code Preferences

## Project Overview
@README.md

## GitHub/Git Workflow Preferences
- **PR Merging**: Always use `--rebase` instead of `--squash` when merging pull requests
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
