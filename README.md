|       |       | **a** |       | **l** | **a** |       | **c** | **a** | **r** | **t** | **e** |       |       |
| ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- |
| **c** | **h** | **e** | **z** | **m** | **o** | **i** |       | **p** | **l** | **u** | **g** | **i** | **n** |

---

Software provisioning and configuration for host and container environments.

---

## Project Overview

**chezmoi-a-la-carte** is a Go-based plugin for [chezmoi](https://chezmoi.io) that provides an advanced, beautiful terminal user interface (TUI) for software provisioning, powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea). It leverages a YAML manifest (`software.yml`) to present and manage available software packages.

- **Language:** Go
- **TUI Framework:** Bubble Tea, Bubbles, Lip Gloss
- **Manifest:** By default, `software.yml`. You can override this by setting the `SOFTWARE_MANIFEST_PATH` environment variable or by creating a `.env` file with `SOFTWARE_MANIFEST_PATH=yourfile.yml`.
- **Binary Name:** `chezmoi-a-la-carte`
- **CI/CD:** GitHub Actions, release-please
- **Commit Style:** [Conventional Commits v1](https://www.conventionalcommits.org/en/v1.0.0/)

## Features

- Interactive TUI for browsing and selecting software
- Integration with chezmoi for seamless provisioning
- Search, filtering, and advanced navigation
- Automated changelog and release management

## Usage

```sh
chezmoi-a-la-carte
```

## Development

1. Install Go (>=1.23)
2. Clone the repo and run:
   ```sh
   go run ./cmd/chezmoi-a-la-carte
   ```
3. The TUI will launch. Press `q` or `Ctrl+C` to quit.
4. Commit messages must follow [Conventional Commits v1](https://www.conventionalcommits.org/en/v1.0.0/), enforced by commitlint and Husky.
5. **All code must pass `golangci-lint run` before commit.** This is enforced by a Husky pre-commit hook. VS Code is pre-configured to show lint errors on save.

## Manifest

- All available software is defined in a YAML manifest (default: `software.yml`). You can use a different file by setting the `SOFTWARE_MANIFEST_PATH` environment variable or in a `.env` file.

## CI/CD

- Automated with GitHub Actions (see `.github/workflows/`).
- Releases and changelogs are managed by [release-please](https://github.com/googleapis/release-please-action).
- Commit messages must follow [Conventional Commits v1](https://www.conventionalcommits.org/en/v1.0.0/), enforced by commitlint and Husky.

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines on commit messages, code quality, pull request process, and our linear history (no merge commits) policy.

## Roadmap

- [ ] Advanced TUI for browsing and selecting software
- [ ] Integration with chezmoi for provisioning
- [ ] Filtering, search, and more

## License

[MIT](LICENSE)

## Acknowledgements

Special thanks to the creators and maintainers of [chezmoi](https://chezmoi.io) and [install.doctor](https://install.doctor) for their amazing projects and inspiration. This project would not exist without their pioneering work in convenient dotfile management and system provisioning automation.
