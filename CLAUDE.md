# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go web application for DesignAI, an AI-driven design portfolio platform. It's a single-binary web server that serves both HTML templates and static assets.

## Architecture

- **Single-file application**: The entire backend is contained in `main.go`
- **Template-based rendering**: Uses Go's `html/template` with embedded file system via `//go:embed`
- **Web framework**: Built with Gin framework for HTTP routing and middleware
- **Template structure**: HTML templates use Go's `{{define}}` syntax and are organized under `templates/pages/`
- **Static assets**: Served from `/assets` route (though no assets directory exists currently)

## Development Commands

### Running the Application
```bash
go run main.go
```
The server starts on port 8080. Access via:
- Main page: http://localhost:8080
- About page: http://localhost:8080/about

### Building
```bash
go build -o design-ai main.go
```

### Dependency Management
```bash
go mod tidy          # Clean up dependencies
go mod download      # Download dependencies
```

## Template System

The application uses Go's embedded file system with a specific template structure:

- Templates must be placed in `templates/pages/` directory
- Each template file must use `{{define "pages/filename"}}...{{end}}` syntax
- Template names in routes correspond to the define names (e.g., `"pages/home"` maps to `home.html` with `{{define "pages/home"}}`)
- Template functions available: `upper` (strings.ToUpper) and `yearNow` (current year)

### Adding New Pages
1. Create HTML file in `templates/pages/`
2. Wrap content with `{{define "pages/pagename"}}...{{end}}`
3. Add corresponding route in `main.go` using `c.HTML(http.StatusOK, "pages/pagename", data)`

## Current Routes
- `GET /` → renders `pages/home` template
- `GET /about` → renders `pages/about` template  
- `GET /assets/*filepath` → serves static files (directory doesn't exist yet)

## Dependencies
- `github.com/gin-gonic/gin` - Web framework
- `github.com/samber/lo` - Utility library (used for `lo.Must`)

## Important Notes

- The embed directive `//go:embed templates/**/*.html` requires templates to exist at build time
- Templates use Chinese language content by default
- No database or external services currently configured
- Application runs in Gin's debug mode by default