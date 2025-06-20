# Claude Code Preferences

## GitHub/Git Workflow Preferences
- **PR Merging**: Always use `--rebase` instead of `--squash` when merging pull requests
- **Branch Cleanup**: After merging PRs, automatically clean up local branches with `git branch -d <branch-name>`
- **PR Creation**: Include comprehensive summary of changes and note existing test coverage rather than generic test plans

## Development Preferences
- Follow existing code conventions and patterns in the codebase
- Check for existing libraries/frameworks before introducing new ones
- Run lint and typecheck commands after making changes (if available)
- Use idiomatic patterns for the languages and frameworks we are using

## Testing
- The project has comprehensive test coverage - acknowledge existing tests rather than creating generic test plans
- Test files are located in `*_test.go` files following Go conventions