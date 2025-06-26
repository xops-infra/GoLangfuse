# GoLangfuse Documentation

This directory contains the production-ready documentation for the GoLangfuse library, designed to be deployed using GitHub Pages.

## Documentation Structure

- `index.html` - Main documentation page with comprehensive API reference
- `styles.css` - Custom styling for professional appearance
- `_config.yml` - Jekyll configuration for GitHub Pages
- `.nojekyll` - Disables Jekyll processing for pure HTML deployment
- `CNAME` - Custom domain configuration (golangfuse.dev)

## Features

The documentation includes:

### 📚 Complete API Reference
- All public interfaces and methods
- Parameter descriptions with types
- Default values and configuration options
- Return value documentation

### 💡 Comprehensive Examples
- Quick start guide
- Complete LLM workflow examples
- Batch processing patterns
- Error handling and monitoring
- Custom HTTP client configuration

### ⚙️ Configuration Guide
- Environment variable reference
- Programmatic configuration examples
- Performance tuning guidelines
- Production deployment settings

### 🎯 Event Types Documentation
- TraceEvent for user interactions
- GenerationEvent for LLM calls
- SpanEvent for processing steps
- ScoreEvent for quality metrics
- Usage tracking structures

## Deployment

The documentation is designed for GitHub Pages deployment:

1. **Automatic Deployment**: Push to the `docs/` directory triggers GitHub Pages build
2. **Custom Domain**: Configured for `golangfuse.dev` (update CNAME as needed)
3. **Mobile Responsive**: Optimized for all device sizes
4. **Search Engine Friendly**: Includes proper meta tags and structured content

## Local Development

To preview locally:

```bash
# Simple HTTP server
python -m http.server 8000 -d docs/

# Or using Node.js
npx serve docs/
```

Visit `http://localhost:8000` to view the documentation.

## Customization

### Styling
- Modify `styles.css` for visual customization
- Uses CSS custom properties for easy theming
- Includes dark mode support for code blocks

### Content
- Update `index.html` for content changes
- Add new sections by extending the existing structure
- Include additional examples in the examples section

### Domain
- Update `CNAME` file for custom domain
- Modify `_config.yml` for site configuration

## Features Included

- 🎨 Professional, modern design
- 📱 Fully responsive layout
- 🔍 Syntax-highlighted code examples
- 📋 Copy-paste ready code snippets
- 🚀 Fast loading with optimized assets
- 🔗 Deep-linkable sections
- 📖 Comprehensive API documentation
- 💡 Real-world usage examples
- ⚙️ Complete configuration reference

The documentation provides everything developers need to successfully integrate and use the GoLangfuse library in production applications.