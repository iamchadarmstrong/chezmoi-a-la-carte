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

## UI Architecture

The TUI (Terminal User Interface) is built using the Bubble Tea framework and is organized into three main packages within `internal/ui/`:

- **`core`**: This package provides the foundational elements for the UI. It includes:

  - **Theme Management**: Defines the `Theme` interface and `DefaultTheme` implementation, along with functions for managing themes (`CurrentTheme`, `RegisterTheme`, `SetTheme`).
  - **Styling**: Contains the `Styles` struct holding various `lipgloss.Style` definitions, functions to build and access current styles (`BuildStyles`, `CurrentStyles`), and layout constants (`PanelWidth`, `ListHeight`, etc.).
  - **Color Helpers**: Utility functions for color manipulation, like `colorToAdaptive`.
  - **Basic UI Models**: Simple, reusable Bubble Tea models like `StringModel` and `EmptyModel`.
  - **Emoji Handling**: Logic for selecting and normalizing emojis for display (`EmojiForEntry`, `NormalizeEmoji`).

- **`components`**: This package contains individual, self-contained UI components that are used to build the TUI. Examples include:

  - `DetailsPanelModel`: Renders the detailed view of a selected software item.
  - `HelpDialogModel`: Displays the help dialog.
  - `ListPaneModel`: Manages and displays lists of software items.
  - `SearchBarModel`: Provides search functionality.
    These components directly use elements from the `core` package for styling and theming.

- **`patterns`**: This package offers more complex UI patterns and layouts composed of core elements and components. Examples include:
  - `Dialog`: A function to create a standardized dialog box.
  - `Card`: A function to create a card-like container.
  - `SplitPaneLayout`: An interface and implementation for creating split-pane views (e.g., list/details).
  - `PlaceOverlay`: A utility to position an overlay (like a dialog) on top of existing content.
  - Container helpers like `NewEnhancedContainer`, `GetListPanelStyle`, and `GetDetailPanelStyle`.

This structure promotes a clear separation of concerns, making it easier to manage and extend the UI. Application code (like in `cmd/chezmoi-a-la-carte/main.go`) primarily interacts with these three packages to construct and manage the user interface.

## Features

- Interactive TUI for browsing and selecting software
- Integration with chezmoi for seamless provisioning
- Search, filtering, and advanced navigation
- Automated changelog and release management

## Usage

```sh
chezmoi-a-la-carte [options]
```

### Options

| Argument          | Short | Description                                        |
| ----------------- | ----- | -------------------------------------------------- |
| `--config FILE`   | `-c`  | Path to configuration file                         |
| `--manifest FILE` | `-m`  | Path to software manifest file                     |
| `--debug`         | `-d`  | Enable debug mode                                  |
| `--version`       | `-v`  | Show version and exit                              |
| `--help`          | `-h`  | Show help message                                  |
| `--output FORMAT` | `-o`  | Output format (text, json) for non-interactive use |
| `--quiet`         | `-q`  | Suppress non-essential output                      |
| `--no-emojis`     | `-E`  | Disable emojis in the UI                           |

For detailed information about the configuration system, see [Configuration System](docs/configuration-system.md).

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
