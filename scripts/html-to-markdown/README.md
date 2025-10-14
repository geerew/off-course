# HTML to Markdown Converter

A Python 3 script that converts HTML files to clean Markdown format. The converter extracts content while removing JavaScript, CSS, inline styles, and other non-content elements.

## Features

- Converts single HTML files or entire directories
- Removes JavaScript (`<script>` tags)
- Removes CSS (`<style>` tags and inline `style` attributes)
- Removes event handlers (onclick, onload, etc.)
- Preserves content structure (headers, paragraphs, images, links, lists)
- Outputs clean, readable Markdown files
- Supports both `.html` and `.htm` file extensions
- Recursive directory processing
- Optional backup: Move HTML files to a backup directory after conversion (absolute or relative)
- Optional deletion: Delete HTML files after successful conversion

## Installation

1. Make sure you have Python 3 installed:
```bash
python3 --version
```

2. Install required dependencies:
```bash
pip install -r requirements.txt
```

Or install manually:
```bash
pip install beautifulsoup4 html2text lxml
```

## Setting Up a Virtual Environment (Recommended)

Using a virtual environment keeps your project dependencies isolated from your system Python installation.

### On macOS/Linux:

```bash
# 1. Navigate to your project directory
cd path/to/html-to-markdown

# 2. Create a virtual environment
python3 -m venv venv

# 3. Activate the virtual environment
source venv/bin/activate

# 4. Install the required packages
pip install -r requirements.txt

# 5. Run the converter (see Usage section below)
python convert.py your_file.html

# 6. When you're done, deactivate the virtual environment
deactivate
```

### On Windows:

```bash
# 1. Navigate to your project directory
cd path\to\html-to-markdown

# 2. Create a virtual environment
python -m venv venv

# 3. Activate the virtual environment
venv\Scripts\activate

# 4. Install the required packages
pip install -r requirements.txt

# 5. Run the converter (see Usage section below)
python convert.py your_file.html

# 6. When you're done, deactivate the virtual environment
deactivate
```

**Note:** When the virtual environment is activated, you'll see `(venv)` at the beginning of your terminal prompt.

## Usage

### Convert a Single File

```bash
python convert.py path/to/file.html
```

This will create `file.md` in the same directory as the HTML file.

### Convert All HTML Files in a Directory

```bash
python convert.py path/to/html_files/
```

This will recursively find all `.html` and `.htm` files in the directory and convert each one to a `.md` file in the same location.

### Backup or Delete Original HTML Files

After successful conversion, you can automatically move the HTML files to a backup directory or delete them:

```bash
# Move HTML files to an absolute backup directory
python convert.py file.html --backup backup/

# Move HTML files to a directory relative to each file
python convert.py file.html --relative-backup old/

# Delete HTML files after conversion
python convert.py file.html --delete
```

**Note:** The `--backup`, `--relative-backup`, and `--delete` options are mutually exclusive - you can only use one at a time.

#### Difference Between --backup and --relative-backup

- **`--backup` (or `-b`)**: All HTML files are moved to one centralized backup directory
  - Example: `docs/page1.html` and `docs/sub/page2.html` both go to `backup/`
  
- **`--relative-backup` (or `-r`)**: Creates a backup directory relative to each HTML file
  - Example: `docs/page1.html` goes to `docs/old/page1.html`
  - Example: `docs/sub/page2.html` goes to `docs/sub/old/page2.html`
  - This preserves the directory structure and keeps backups near the original files

### Examples

```bash
# Convert a single file
python convert.py example.html

# Convert all HTML files in the current directory
python convert.py .

# Convert all HTML files in a specific folder
python convert.py path/to/html-files/

# Convert and backup HTML files (absolute backup)
python convert.py path/to/html-files/ --backup backup_html/

# Convert and backup HTML files (relative to each file)
python convert.py path/to/html-files/ --relative-backup old/

# Convert and delete HTML files (use with caution!)
python convert.py example.html --delete

# Short form: -b for backup, -r for relative-backup, -d for delete
python convert.py example.html -b backup/
python convert.py example.html -r old/
python convert.py example.html -d
```

## What Gets Converted

The script preserves:
- Headers (h1-h6) → Markdown headers (#, ##, ###, etc.)
- Paragraphs → Text with proper spacing
- Links → `[text](url)`
- Images → `![alt text](image-url)`
- Lists (ordered and unordered) → Markdown lists
- Bold and italic text
- Code blocks
- Blockquotes
- Tables

The script removes:
- JavaScript code (`<script>` tags)
- CSS styles (`<style>` tags)
- Inline style attributes
- Event handlers (onclick, onload, etc.)

## Output

Each HTML file is converted to a Markdown file with the same name but with a `.md` extension in the same directory as the source file.

For example:
- `document.html` → `document.md`
- `page.htm` → `page.md`

## Error Handling

- If a file cannot be read or parsed, an error message is displayed, and the script continues with other files
- The script reports the number of successful and failed conversions at the end
- Non-HTML files are skipped with a warning
- HTML files are only moved to backup or deleted after successful conversion
- If a file already exists in the backup directory, a number suffix is added to avoid overwriting (e.g., `file_1.html`, `file_2.html`)
- The `--backup`, `--relative-backup`, and `--delete` options cannot be used together

## Requirements

- Python 3.6 or higher
- beautifulsoup4 >= 4.12.0
- html2text >= 2020.1.16
- lxml >= 4.9.0

## License

This is free and unencumbered software released into the public domain.

