# chezmoi-a-la-carte

|       | **a** | **l** | **a** |       |
| ----- | ----- | ----- | ----- | ----- |
| **C** | **A** | **R** | **T** | **E** |

[chezmoi](https://chezmoi.io) plugin for software provisioning on new systems and container environments.

---

## Project Overview

**chezmoi-a-la-carte** is a Go-based plugin for chezmoi that provides an advanced, beautiful terminal user interface (TUI) for software provisioning, powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea). It leverages a YAML manifest (`software.yml`) to present and manage available software packages.

- **Language:** Go
- **TUI Framework:** Bubble Tea, Bubbles, Lip Gloss
- **Manifest:** `software.yml`
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

## Manifest

- All available software is defined in `software.yml` (YAML format).

## CI/CD

- Automated with GitHub Actions (see `.github/workflows/`).
- Releases and changelogs are managed by [release-please](https://github.com/googleapis/release-please-action).
- Commit messages must follow [Conventional Commits v1](https://www.conventionalcommits.org/en/v1.0.0/), enforced by commitlint and Husky.

## Contributing

- Please use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for all commit messages.
- Pull requests are welcome! For major changes, please open an issue first to discuss what you would like to change.
- See the [CONTRIBUTING.md](CONTRIBUTING.md) for more details (if available).

## Roadmap

- [ ] Advanced TUI for browsing and selecting software
- [ ] Integration with chezmoi for provisioning
- [ ] Filtering, search, and more

## License

[MIT](LICENSE)

## Acknowledgements

Special thanks to the creators and maintainers of [chezmoi](https://chezmoi.io) and [install.doctor](https://install.doctor) for their amazing projects and inspiration. This project would not exist without their pioneering work in convenient dotfile management and system provisioning automation.
