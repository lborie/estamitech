# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

L'Estamitech is a French tech podcast website built with Go and deployed on Google App Engine. The site serves as a showcase for the podcast, featuring episode listings, individual episode pages with SEO optimization, and social sharing capabilities.

## Architecture

This is a simple Go web server using the standard library with server-side rendering:

- **RSS Integration**: Fetches podcast episodes from Zencastr RSS feed with 30-second caching
- **Server-Side Rendering**: Uses Go HTML templates (`web/*.gohtml`) for dynamic page generation
- **Dynamic OG Images**: Generates Open Graph images on-the-fly using the `github.com/fogleman/gg` library
- **SEO Optimization**: Includes sitemap.xml generation and optimized meta tags

### Key Files

- `main.go` - Main HTTP server with RSS caching, episode routing, and OG image generation
- `web/index.gohtml` - Homepage template with episode listings
- `web/episode.gohtml` - Individual episode page template with SEO meta tags
- `static/` - Static assets (logos, icons, background images)
- `app.yaml` - Google App Engine configuration with static file handling

### HTTP Endpoints

- `GET /` - Homepage with all episodes
- `GET /episode/{id}` - Individual episode page by GUID
- `GET /og-image/{id}` - Dynamic Open Graph image generation
- `GET /sitemap.xml` - SEO sitemap with all episodes

## Development Commands

### Local Development
```bash
# Start local development server
go run main.go

# Start with App Engine emulator
make dev-run
```

### Deployment
```bash
# Deploy to Google App Engine (requires environment variables)
export ESTAMITECH_ACCOUNT_ID="your-account@gmail.com"
export ESTAMITECH_PROJECT_ID="your-project-id"
make deploy
```

## Dependencies

- Go 1.23+ (minimum declared in `go.mod`); App Engine Standard runtime `go127` (`app.yaml`)
- `github.com/fogleman/gg` - For dynamic image generation
- Google Cloud SDK (for deployment)

## Development Notes

- No testing framework is currently configured
- No linting or formatting commands are set up
- RSS cache expires after 30 seconds to balance performance and freshness
- Static files are served directly by App Engine (configured in `app.yaml`)
- French language site with SEO optimized for French tech content