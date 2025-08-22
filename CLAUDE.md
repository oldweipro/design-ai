# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go web application for DesignAI, an AI-driven design portfolio platform. It's a full-stack web application with SQLite database backend, REST API, and dynamic frontend that displays design portfolio works.

## Architecture

- **Main application**: `main.go` contains server setup, routing, and middleware
- **Database layer**: SQLite database with GORM ORM for data persistence
- **Models**: `models/portfolio.go` contains data structures and validation
- **Handlers**: `handlers/portfolio.go` contains API endpoint implementations  
- **Database setup**: `database/database.go` handles connection and migrations
- **Template-based frontend**: HTML templates with embedded JavaScript for dynamic behavior
- **Template structure**: HTML templates use Go's `{{define}}` syntax and are organized under `templates/pages/`
- **API-driven**: Frontend fetches data from REST API endpoints

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

## API Endpoints

### Portfolio Management
- `GET /api/v1/portfolios` - List portfolios with filtering and search
  - Query params: `category`, `search`, `page`, `page_size`, `sort_by`, `order`
- `GET /api/v1/portfolios/:id` - Get portfolio details (increments view count)
- `POST /api/v1/portfolios` - Create new portfolio
- `PUT /api/v1/portfolios/:id` - Update existing portfolio
- `DELETE /api/v1/portfolios/:id` - Soft delete portfolio
- `POST /api/v1/portfolios/:id/like` - Like a portfolio (increments like count)

### Categories
- `GET /api/v1/categories` - Get available portfolio categories

## Database Schema

### Portfolio Table
- `id` (UUID, primary key)
- `title` (string, required) - Portfolio title
- `author` (string, required) - Author name
- `description` (text) - Portfolio description
- `category` (string, required, indexed) - Portfolio category (ai, ui, web, mobile, brand, 3d)
- `tags` (JSON string) - Array of tags stored as JSON
- `image_url` (string) - URL to portfolio image
- `ai_level` (string) - AI involvement level (AI完全生成, AI辅助设计, 手工设计)
- `likes` (integer, default 0) - Like count
- `views` (integer, default 0) - View count  
- `status` (string, default 'published') - Portfolio status
- `created_at`, `updated_at` (timestamps)

## Dependencies
- `github.com/gin-gonic/gin` - Web framework
- `github.com/samber/lo` - Utility library (used for `lo.Must`)
- `gorm.io/gorm` - ORM for database operations
- `gorm.io/driver/sqlite` - SQLite driver (replaced with CGO-free version)
- `github.com/google/uuid` - UUID generation

## Important Notes

- Uses CGO-free SQLite driver (`github.com/glebarez/sqlite`) for cross-platform compatibility
- Database file `design_ai.db` is created automatically on first run
- Sample data is seeded automatically if database is empty
- The embed directive `//go:embed templates/**/*.html` requires templates to exist at build time
- Templates use Chinese language content by default
- Frontend dynamically loads data from API endpoints using JavaScript fetch
- CORS is enabled for all origins (not recommended for production)
- Application runs in Gin's debug mode by default

## Frontend Integration

- Frontend uses JavaScript fetch API to communicate with backend
- Search functionality includes debouncing (500ms delay)
- Real-time filtering by category and search terms
- Like functionality updates both database and UI immediately
- Portfolio upload form validates and categorizes content automatically
- Error handling with user-friendly toast notifications