# Off Course

## Overview

Off Course is a local course management application that enables you to view, organize, and track progress through educational content
stored on your local filesystem. It provides a web-based interface for browsing courses, tracking learning progress, and managing course
materials without requiring an internet connection.

The application automatically scans course directories to identify assets (videos, PDFs, markdown files and text files) and attachments,
organizing them into structured lessons with progress tracking capabilities.

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

## Installation & Running

### Prerequisites

**For Manual Installation:**

- Node.js >= 22.12.0
- pnpm >= 8
- Go >= 1.20
- FFmpeg and FFProbe (for video processing and HLS transcoding)

**For Docker:**

- Docker
- Docker Compose

### Building the Application

The application consists of a Go backend that embeds a SvelteKit frontend. You must build the frontend first, then build the Go application
which will embed the UI.

1. **Build the Frontend**

   ```bash
   cd ui
   pnpm install
   pnpm run build
   cd ..
   ```

2. **Build the Go Application**

   ```bash
   # Install Go dependencies
   go mod download

   # Build the application (embeds the UI)
   go build -o off-course .

   # Or build for specific platform
   GOOS=linux GOARCH=amd64 go build -o off-course-linux .
   ```

The built application will include the frontend UI embedded within the Go binary.

### Running Off Course

#### Option 1: Docker (Recommended for Production)

For production deployment, see the [Docker README](docker/README.md) for complete setup instructions including:

- Building Docker images
- Docker Compose configuration
- Volume mounting for data persistence
- Environment variable configuration

Quick Docker start:

```bash
# Using Docker Compose
docker-compose up -d

# Or using Docker directly
docker run -p 9081:9081 -v $(pwd)/oc_data:/app/oc_data off-course
```

#### Option 2: Manual Installation

1. **Clone and Build**

   ```bash
   git clone https://github.com/geerew/off-course.git
   cd off-course

   # Build the application (see Building section above)
   cd ui && pnpm install && pnpm run build && cd ..
   go build -o off-course .
   ```

2. **Run the Application**

   ```bash
   # Basic run
   ./off-course serve

   # With custom settings
   ./off-course serve --http 0.0.0.0:8080 --data-dir ./my-data

   # Enable user signup
   ./off-course serve --enable-signup
   ```

3. **Access the Application**
   - Visit `http://localhost:9081` (or your configured port)
   - Create your first admin account when prompted

### Development Mode

For development with hot reloading:

1. **Start the backend** (Terminal 1)

   ```bash
   # Install dependencies
   go mod download

   # Run with air for hot reloading (recommended)
   air

   # Or run directly
   go run main.go

   # To change the port (default is 9081)
   air -- --http 0.0.0.0:8080
   ```

2. **Start the frontend** (Terminal 2)

   ```bash
   cd ui

   # Install dependencies
   pnpm install

   # Start development server
   pnpm run dev
   ```

3. **Access the application**
   - Visit `http://localhost:9081` (not the SvelteKit dev server port)
   - The backend serves the built frontend and handles API requests

### Database

The application uses SQLite databases stored in the `oc_data` directory, created automatically when the application is first launched:

- `oc_data/data.db` - Main application data (courses, users, progress)
- `oc_data/logs.db` - Application logs

The database is created relative to where the application is launched, so the `oc_data` directory will appear in your current working
directory.

### Admin Password Reset

If you get locked out of your admin account, you can reset the password using the CLI:

```bash
# Reset password for an admin user
./off-course admin reset-password <username>

# Example
./off-course admin reset-password admin
```

This command will:

1. Verify the user exists and is an admin
2. Prompt for a new password
3. Generate a secure recovery token
4. Communicate with the running application to reset the password
5. Clean up the recovery token automatically

**Requirements:**

- The application must be running
- You need filesystem access to the data directory
- The user must have admin role

### CLI Commands

Off Course includes several CLI commands for administration:

```bash
# Start the application
./off-course serve [options]

# Admin commands
./off-course admin reset-password <username>  # Reset admin password

# Get help for any command
./off-course --help
./off-course admin --help
./off-course serve --help
```

**Available serve options:**

- `--http <address>` - HTTP server address (default: 127.0.0.1:9081)
- `--data-dir <path>` - Data directory path (default: ./oc_data)
- `--enable-signup` - Allow user registration
- `--dev` - Run in development mode

## Adding Courses

### Course Structure

Courses are organized as directories containing assets and attachments. The application automatically scans these directories to identify
and organize content.

#### Course Cards

Place an image named `card.xxx` at the root of your course directory (where `xxx` is a supported image extension: `.jpg`, `.png`, `.webp`, `.tiff`).

#### Assets vs Attachments

- **Assets**: Primary course materials (videos, PDFs, markdown files and text files)
- **Attachments**: Supplementary materials linked to specific assets

#### File Organization

```
My Course/
├── card.jpg                   # Course card image
├── Chapter 1/                 # Module (chapter)
│   ├── 01 Overview.mp4        # First asset
│   ├── 01 Overview Notes.txt  # Attachment for first asset
│   └── 02 Example.md          # Second asset
└── Chapter 2/
    ├── 01 Deep Dive.pdf       # First asset
    └── 01 Source Links.txt    # Attachment for first asset
```

#### Filename Structure

**Assets** must follow this pattern:

```
{number} {title}.{extension}
```

Examples:

- `01 Introduction.mp4`
- `02 Advanced Concepts.pdf`
- `3 Getting Started.md`

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
2. **PDF** - `.pdf`
3. **Markdown** - `.md`
4. **Text** (lowest priority) - `.txt`

Example: If you have both `01 Introduction.mp4` and `01 Introduction.md`, the video will be the primary asset and the markdown file
will become an attachment

#### Supported Asset Types

**Video/Audio:**

- `.mp4`, `.avi`, `.mkv`, `.webm`, `.ogv`
- `.mp3`, `.m4a`, `.ogg`, `.wav`, `.flac`

**Documents:**

- `.pdf`
- `.md`,
- `.txt`

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

**Note**: When running tests outside of `make test`, use `-tags dev`. This bypasses the need for the ui to be built

```bash
# Run all tests with dev tag
go test -tags dev ./...

# Run specific package tests
go test -tags dev ./api/...
```

