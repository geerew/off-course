# Off Course

## Overview

Off Course is a local course management application that enables you to view, organize, and track progress through educational content stored on your local filesystem. It provides a web-based interface for browsing courses, tracking learning progress, and managing course materials without requiring an internet connection.

The application automatically scans course directories to identify assets (videos, HTML files, PDFs) and attachments, organizing them into structured lessons with progress tracking capabilities.

## Architecture

### Frontend

- **SvelteKit** with TypeScript for the web interface
- Modern, responsive UI with Tailwind CSS
- Real-time progress tracking and course management
- Built-in media player for video content

### Backend

- **Go** application with RESTful API
- **SQLite** database for data persistence
- File system scanning and course processing
- Background job processing for course availability monitoring

## Development Setup

### Prerequisites

**Frontend**

- Node.js >= 22.12.0
- pnpm >= 8

**Backend**

- Go >= 1.20
- [Air](https://github.com/cosmtrek/air) for hot reloading (recommended)

### Running in Development Mode

1. **Clone the repository**

   ```bash
   git clone https://github.com/geerew/off-course.git
   cd off-course
   ```

2. **Start the backend** (Terminal 1)

   ```bash
   # Install dependencies
   go mod download

   # Run with air for hot reloading
   air

   # Or run directly
   go run main.go

   # To change the port (default is 9080)
   air -- --http 0.0.0.0:8080
   ```

3. **Start the frontend** (Terminal 2)

   ```bash
   cd ui

   # Install dependencies
   pnpm install

   # Start development server
   pnpm run dev
   ```

4. **Access the application**
   - Visit `http://localhost:9080` (not the SvelteKit dev server port)
   - The backend serves the built frontend and handles API requests

### Database

The application uses SQLite databases stored in the `oc_data` directory, created automatically when the application is first launched:

- `oc_data/data.db` - Main application data (courses, users, progress)
- `oc_data/logs.db` - Application logs

The database is created relative to where the application is launched, so the `oc_data` directory will appear in your current working directory.

## Docker Setup

For production deployment, see the [Docker README](docker/README.md) for complete setup instructions including:

- Building Docker images
- Docker Compose configuration
- Volume mounting for data persistence
- Environment variable configuration

## Adding Courses

### Course Structure

Courses are organized as directories containing assets and attachments. The application automatically scans these directories to identify and organize content.

#### Course Cards

Place an image named `card.xxx` at the root of your course directory (where `xxx` is a supported image extension: `.jpg`, `.png`, `.webp`, `.tiff`).

#### Assets vs Attachments

- **Assets**: Primary course materials (videos, HTML files, PDFs, text files)
- **Attachments**: Supplementary materials linked to specific assets

#### File Organization

```
My Course/
├── card.jpg                    # Course card image
├── 01 Introduction.mp4         # Main asset
├── 01 Notes.txt               # Attachment to '01 Introduction.mp4'
├── Chapter 1/                 # Subdirectory (chapter/section)
│   ├── 01 Overview.mp4         # Main asset for this chapter
│   ├── 01 Overview Notes.txt   # Attachment
│   └── 02 Example.html         # Another asset
└── Chapter 2/
    ├── 01 Deep Dive.pdf       # Main asset
    └── 01 Source Links.txt     # Attachment
```

#### Filename Structure

**Assets** must follow this pattern:

```
{number} {title}.{extension}
```

Examples:

- `01 Introduction.mp4`
- `02 Getting Started.html`
- `03 Advanced Concepts.pdf`

**Attachments** are linked to assets by numerical prefix:

```
{number} {optional title}.{extension}
```

Examples:

- `01` (becomes attachment to `01 Introduction.mp4`)
- `01 Notes.txt` (becomes attachment to `01 Introduction.mp4`)
- `01 Introduction Notes.pdf` (becomes attachment to `01 Introduction.mp4`)

#### Multiple Assets with Sub-Prefix

For lessons with multiple assets (e.g., multiple video parts), use the sub-prefix syntax:

```
01 Introduction {1 Part 1}.mp4
01 Introduction {2 Part 2}.mp4
01 Introduction {3 Part 3}.mp4
```

This creates three separate assets for the same lesson, each with their own sub-prefix and optional sub-title.

The assets will be rendered in the order of the sub-prefix on the same lesson page.

#### Asset Priority

When multiple assets share the same prefix (without sub-prefix), the system uses priority to determine the primary asset:

1. **Video** (highest priority) - `.mp4`, `.avi`, `.mkv`, `.webm`, etc.
2. **HTML** - `.html`, `.htm`
3. **PDF** - `.pdf`
4. **Markdown** - `.md`
5. **Text** (lowest priority) - `.txt`

Example: If you have both `01 Introduction.mp4` and `01 Introduction.html`, the video will be the primary asset and the HTML file will become an attachment.

#### Supported Asset Types

**Video/Audio:**

- `.mp4`, `.avi`, `.mkv`, `.webm`, `.ogv`
- `.mp3`, `.m4a`, `.ogg`, `.wav`, `.flac`

**Documents:**

- `.html`, `.htm`
- `.pdf`
- `.md`, `.txt`

### Adding Courses via UI

1. Navigate to **Settings** > **Courses**
2. Click **Add Courses**
3. Select course directories from your file system
4. The application will automatically scan and organize the content

### Scanning and Availability

- **Automatic Scanning**: Courses are scanned when first added
- **Manual Scanning**: Use the "Scan" option in the course menu to refresh content
- **Availability Monitoring**: Background jobs check course availability
- **Maintenance Mode**: Courses are locked during scanning to prevent conflicts

## Contributing

We welcome contributions! Here's how you can help:

### Development

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature-name`
3. Make your changes
4. Run tests: `make test`
5. Run quality checks: `make audit`
6. Commit your changes: `git commit -m "Add your feature"`
7. Push to your fork: `git push origin feature/your-feature-name`
8. Create a Pull Request

### Code Quality

- Follow Go best practices and the existing code style
- Write tests for new functionality
- Ensure all tests pass: `make test`
- Run the full audit: `make audit`
- Format code: `make tidy`

### Areas for Contribution

- **Frontend**: UI/UX improvements, new features, accessibility
- **Backend**: API enhancements, performance optimizations
- **Course Processing**: Enhanced file type support, better scanning algorithms
- **Documentation**: Improved guides, examples, API documentation
- **Testing**: Additional test coverage, integration tests

### Reporting Issues

- Use GitHub Issues for bug reports and feature requests
- Include steps to reproduce for bugs
- Provide system information (OS, Go version, etc.)

## License

[Add your license information here]
