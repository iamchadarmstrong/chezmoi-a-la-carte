# UI Module Organization

This directory contains the user interface (UI) components for the a-la-carte application.

## Directory Structure

- **core/**: Foundational UI primitives and interfaces

  - `container.go`: Base container component for UI elements
  - `theme.go`: Theme definitions and management
  - `styles.go`: Shared styles and layout constants

- **components/**: Interactive UI components

  - `helpdialog.go`: Help dialog component
  - `toggle.go`: Toggle switch component
  - `form.go`: Form component for collecting user input
  - (more components to be added)

- **patterns/**: Reusable UI design patterns

  - `containers.go`: Common container patterns (Panel, Dialog, etc.)
  - (more patterns to be added)

- **util/**: Helper utilities for the UI
  - (utility files to be added)

## Usage Guidelines

### Core Containers

Use the core Container interface for building UI elements:

```go
import "a-la-carte/internal/ui/core"

container := core.NewContainer(
    content,
    core.WithBorderAll(),
    core.WithPaddingAll(1),
)
```

### Container Patterns

For commonly used container styles, use the patterns package:

```go
import "a-la-carte/internal/ui/patterns"

dialog := patterns.Dialog(content)
panel := patterns.Panel(content)
```

### Creating New Components

When creating new UI components:

1. Use the core Container as a building block
2. Place component logic in the `components/` directory
3. Follow the existing patterns for state management and rendering

### For New Code

For new code, use the more specific imports:

```go
import (
    "a-la-carte/internal/ui/core"
    "a-la-carte/internal/ui/components"
    "a-la-carte/internal/ui/patterns"
)

// Use core for foundational elements
container := core.NewContainer(...)
theme := core.CurrentTheme()

// Use layouts for layout algorithms
result := layouts.PlaceOverlay(...)

// Use patterns for common UI patterns
panel := patterns.Panel(content)
dialog := patterns.Dialog(content)

// Use components for interactive elements
toggle := components.NewToggleModel()
form := components.NewFormModel()
```
