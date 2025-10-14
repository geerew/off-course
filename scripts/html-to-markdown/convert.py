#!/usr/bin/env python3
"""
HTML to Markdown Converter
Converts HTML files to Markdown format, removing JavaScript, CSS, and styles.
"""

import os
import sys
import argparse
import shutil
from pathlib import Path
from bs4 import BeautifulSoup
import html2text


def clean_html(html_content):
    """
    Remove JavaScript, CSS, and style attributes from HTML.
    
    Args:
        html_content (str): Raw HTML content
        
    Returns:
        str: Cleaned HTML content
    """
    soup = BeautifulSoup(html_content, 'lxml')
    
    # Remove script tags
    for script in soup.find_all('script'):
        script.decompose()
    
    # Remove style tags
    for style in soup.find_all('style'):
        style.decompose()
    
    # Remove inline style attributes
    for tag in soup.find_all(True):
        if tag.has_attr('style'):
            del tag['style']
        # Also remove event handlers (onclick, onload, etc.)
        attrs_to_remove = [attr for attr in tag.attrs if attr.startswith('on')]
        for attr in attrs_to_remove:
            del tag[attr]
    
    return str(soup)


def html_to_markdown(html_content):
    """
    Convert HTML to Markdown format.
    
    Args:
        html_content (str): HTML content
        
    Returns:
        str: Markdown formatted text
    """
    # Clean the HTML first
    cleaned_html = clean_html(html_content)
    
    # Configure html2text
    h = html2text.HTML2Text()
    h.ignore_links = False
    h.ignore_images = False
    h.ignore_emphasis = False
    h.body_width = 0  # Don't wrap lines
    h.skip_internal_links = False
    h.inline_links = True
    h.protect_links = True
    h.wrap_links = False
    
    # Convert to markdown
    markdown = h.handle(cleaned_html)
    
    return markdown.strip()


def convert_file(input_path, output_path=None, backup_dir=None, relative_backup_dir=None, delete_after=False):
    """
    Convert a single HTML file to Markdown.
    
    Args:
        input_path (Path): Path to input HTML file
        output_path (Path, optional): Path to output MD file. 
                                     If None, saves in same directory with .md extension
        backup_dir (Path, optional): Directory to move HTML file to after conversion
        relative_backup_dir (str, optional): Directory name for backup relative to each HTML file
        delete_after (bool): Whether to delete HTML file after successful conversion
    
    Returns:
        bool: True if successful, False otherwise
    """
    try:
        # Read HTML file
        with open(input_path, 'r', encoding='utf-8') as f:
            html_content = f.read()
        
        # Convert to markdown
        markdown_content = html_to_markdown(html_content)
        
        # Determine output path
        if output_path is None:
            output_path = input_path.with_suffix('.md')
        
        # Write markdown file
        with open(output_path, 'w', encoding='utf-8') as f:
            f.write(markdown_content)
        
        print(f"✓ Converted: {input_path} → {output_path}")
        
        # Handle backup or deletion after successful conversion
        if backup_dir:
            # Absolute backup: all files go to the same backup directory
            backup_dir.mkdir(parents=True, exist_ok=True)
            
            # Preserve directory structure in backup
            backup_path = backup_dir / input_path.name
            
            # If file already exists in backup, add number suffix
            counter = 1
            original_backup_path = backup_path
            while backup_path.exists():
                stem = original_backup_path.stem
                suffix = original_backup_path.suffix
                backup_path = backup_dir / f"{stem}_{counter}{suffix}"
                counter += 1
            
            shutil.move(str(input_path), str(backup_path))
            print(f"  → Moved to backup: {backup_path}")
            
        elif relative_backup_dir:
            # Relative backup: create backup directory relative to each HTML file
            file_parent = input_path.parent
            backup_path = file_parent / relative_backup_dir
            backup_path.mkdir(parents=True, exist_ok=True)
            
            # Target file in the relative backup directory
            target_path = backup_path / input_path.name
            
            # If file already exists in backup, add number suffix
            counter = 1
            original_target_path = target_path
            while target_path.exists():
                stem = original_target_path.stem
                suffix = original_target_path.suffix
                target_path = backup_path / f"{stem}_{counter}{suffix}"
                counter += 1
            
            shutil.move(str(input_path), str(target_path))
            print(f"  → Moved to backup: {target_path}")
            
        elif delete_after:
            os.remove(input_path)
            print(f"  → Deleted: {input_path}")
        
        return True
        
    except Exception as e:
        print(f"✗ Error converting {input_path}: {str(e)}", file=sys.stderr)
        return False


def process_path(path_str, backup_dir=None, relative_backup_dir=None, delete_after=False):
    """
    Process a file or directory path, converting all HTML files found.
    
    Args:
        path_str (str): Path to file or directory
        backup_dir (Path, optional): Directory to move HTML files to after conversion
        relative_backup_dir (str, optional): Directory name for backup relative to each HTML file
        delete_after (bool): Whether to delete HTML files after successful conversion
        
    Returns:
        tuple: (success_count, failure_count)
    """
    path = Path(path_str)
    
    if not path.exists():
        print(f"Error: Path does not exist: {path}", file=sys.stderr)
        return 0, 0
    
    success_count = 0
    failure_count = 0
    
    if path.is_file():
        # Process single file
        if path.suffix.lower() in ['.html', '.htm']:
            if convert_file(path, backup_dir=backup_dir, relative_backup_dir=relative_backup_dir, delete_after=delete_after):
                success_count += 1
            else:
                failure_count += 1
        else:
            print(f"Error: File must have .html or .htm extension: {path}", file=sys.stderr)
            failure_count += 1
    
    elif path.is_dir():
        # Process all HTML files in directory
        html_files = list(path.glob('**/*.html')) + list(path.glob('**/*.htm'))
        
        if not html_files:
            print(f"No HTML files found in directory: {path}")
            return 0, 0
        
        print(f"Found {len(html_files)} HTML file(s) to convert...\n")
        
        for html_file in html_files:
            if convert_file(html_file, backup_dir=backup_dir, relative_backup_dir=relative_backup_dir, delete_after=delete_after):
                success_count += 1
            else:
                failure_count += 1
    
    return success_count, failure_count


def main():
    """Main entry point for the script."""
    parser = argparse.ArgumentParser(
        description='Convert HTML files to Markdown format, removing JavaScript and CSS.',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  python convert.py file.html                           # Convert a single file
  python convert.py /path/to/folder/                    # Convert all HTML files in a folder
  python convert.py file.html --backup backup/          # Convert and move HTML to backup folder
  python convert.py file.html --relative-backup old/    # Convert and move HTML to ./old/ (relative to file)
  python convert.py file.html --delete                  # Convert and delete HTML file
        """
    )
    
    parser.add_argument(
        'path',
        help='Path to an HTML file or directory containing HTML files'
    )
    
    # Create mutually exclusive group for backup and delete options
    cleanup_group = parser.add_mutually_exclusive_group()
    
    cleanup_group.add_argument(
        '-b', '--backup',
        metavar='DIR',
        type=str,
        help='Move HTML files to this backup directory after successful conversion (absolute path)'
    )
    
    cleanup_group.add_argument(
        '-r', '--relative-backup',
        metavar='DIR',
        type=str,
        help='Move HTML files to this directory relative to each file after conversion'
    )
    
    cleanup_group.add_argument(
        '-d', '--delete',
        action='store_true',
        help='Delete HTML files after successful conversion'
    )
    
    parser.add_argument(
        '-v', '--version',
        action='version',
        version='HTML to Markdown Converter 1.2'
    )
    
    args = parser.parse_args()
    
    # Prepare backup directory if specified
    backup_dir = Path(args.backup) if args.backup else None
    relative_backup_dir = args.relative_backup if args.relative_backup else None
    
    # Process the path
    print("HTML to Markdown Converter\n" + "="*50 + "\n")
    
    if backup_dir:
        print(f"Backup directory: {backup_dir}\n")
    elif relative_backup_dir:
        print(f"Relative backup directory: {relative_backup_dir}\n")
    elif args.delete:
        print("HTML files will be deleted after conversion\n")
    
    success_count, failure_count = process_path(
        args.path,
        backup_dir=backup_dir,
        relative_backup_dir=relative_backup_dir,
        delete_after=args.delete
    )
    
    # Print summary
    print("\n" + "="*50)
    print(f"Conversion complete!")
    print(f"  Successful: {success_count}")
    print(f"  Failed: {failure_count}")
    
    # Exit with appropriate code
    sys.exit(0 if failure_count == 0 else 1)


if __name__ == '__main__':
    main()

