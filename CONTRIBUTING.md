# Contributing to chezmoi-a-la-carte

Thank you for your interest in contributing! Please follow these guidelines to help us keep the project clean, maintainable, and easy to use for everyone.

## Commit Messages

- Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for all commit messages.
- Commit messages are linted and enforced by commitlint and Husky.

## Linting & Code Quality

- All code must pass `golangci-lint run` before commit. This is enforced by a Husky pre-commit hook.
- You can run `golangci-lint run` manually to check for issues before committing.

## Pull Requests

- Pull requests are welcome! For major changes, please open an issue first to discuss what you would like to change.
- All pull requests must target the `main` branch unless otherwise specified.

### Linear History Required (No Merge Commits)

- **All pull requests must have a linear commit history (no merge commits).**
- Our CI will **fail any pull request that contains a merge commit**. Please use `git rebase` to update your branch instead of merging `main`.
- If you see a CI failure about merge commits, run:
  ```sh
  git fetch origin
  git rebase origin/main
  # Resolve any conflicts, then push with --force
  git push --force-with-lease
  ```
- This policy keeps our history clean and easy to follow.

## Development Quickstart

1. Install Go (>=1.23)
2. Clone the repo and run:
   ```sh
   go run ./cmd/chezmoi-a-la-carte
   ```
3. The TUI will launch. Press `q` or `Ctrl+C` to quit.

## Additional Information

- See the [README.md](README.md) for project overview, features, and more details.
- For questions or help, please open an issue.
