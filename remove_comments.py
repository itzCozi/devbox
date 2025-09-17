#!/usr/bin/env python3

import os
import sys
import argparse
from typing import List, Tuple


class GoCommentRemover:
    def __init__(self, preserve_build_tags=True):
        self.preserve_build_tags = preserve_build_tags
        self.processed_files = 0
        self.total_comments_removed = 0
        
    def is_go_file(self, filepath: str) -> bool:
        """Check if file is a Go source file."""
        return filepath.endswith('.go')
    
    def find_go_files(self, directory: str) -> List[str]:
        """Recursively find all Go files in directory."""
        go_files = []
        for root, dirs, files in os.walk(directory):
            # Skip vendor and .git directories
            dirs[:] = [d for d in dirs if d not in ['vendor', '.git', '.vscode', '__pycache__']]
            
            for file in files:
                if self.is_go_file(file):
                    go_files.append(os.path.join(root, file))
        return go_files
    
    def should_preserve_comment(self, comment: str) -> bool:
        """Check if comment should be preserved (build tags, etc.)."""
        if not self.preserve_build_tags:
            return False
            
        comment_content = comment.strip('/* \t\n/')
        
        # Preserve build tags
        if comment_content.startswith('+build') or comment_content.startswith('go:'):
            return True

        return False
    
    def remove_comments_from_content(self, content: str) -> Tuple[str, int]:
        """
        Remove comments from Go source code while preserving string literals.
        Returns (cleaned_content, number_of_comments_removed).
        """
        lines = content.split('\n')
        cleaned_lines = []
        comments_removed = 0
        in_multiline_comment = False
        multiline_start_line = -1
        
        for i, line in enumerate(lines):
            cleaned_line = ""
            j = 0
            in_string = False
            in_raw_string = False
            string_delimiter = None
            
            while j < len(line):
                char = line[j]
                
                if char == '`' and not in_string:
                    in_raw_string = not in_raw_string
                    cleaned_line += char
                    j += 1
                    continue
                
                if in_raw_string:
                    cleaned_line += char
                    j += 1
                    continue
                
                if char in ['"', "'"] and not in_string:
                    in_string = True
                    string_delimiter = char
                    cleaned_line += char
                    j += 1
                    continue
                elif char == string_delimiter and in_string:
                    if j > 0 and line[j-1] == '\\':
                        backslash_count = 0
                        k = j - 1
                        while k >= 0 and line[k] == '\\':
                            backslash_count += 1
                            k -= 1
                        if backslash_count % 2 == 1:
                            cleaned_line += char
                            j += 1
                            continue
                    
                    in_string = False
                    string_delimiter = None
                    cleaned_line += char
                    j += 1
                    continue
                
                if in_string:
                    cleaned_line += char
                    j += 1
                    continue
                
                if in_multiline_comment:
                    if j < len(line) - 1 and line[j:j+2] == '*/':
                        in_multiline_comment = False
                        j += 2
                        if multiline_start_line != i:
                            comments_removed += 1
                        continue
                    else:
                        j += 1
                        continue
                
                if j < len(line) - 1 and line[j:j+2] == '/*':
                    end_pos = line.find('*/', j + 2)
                    if end_pos != -1:
                        comment_content = line[j:end_pos+2]
                        if not self.should_preserve_comment(comment_content):
                            comments_removed += 1
                            j = end_pos + 2
                            continue
                        else:
                            cleaned_line += comment_content
                            j = end_pos + 2
                            continue
                    else:
                        comment_content = line[j:]
                        if not self.should_preserve_comment(comment_content):
                            in_multiline_comment = True
                            multiline_start_line = i
                            break
                        else:
                            cleaned_line += line[j:]
                            break
                
                if j < len(line) - 1 and line[j:j+2] == '//':
                    comment_content = line[j:]
                    if not self.should_preserve_comment(comment_content):
                        comments_removed += 1
                        break
                    else:
                        cleaned_line += comment_content
                        break
                
                cleaned_line += char
                j += 1
            
            if not in_multiline_comment:
                if cleaned_line != line:
                    cleaned_line = cleaned_line.rstrip()
                cleaned_lines.append(cleaned_line)
        
        return '\n'.join(cleaned_lines), comments_removed
    
    def process_file(self, filepath: str, dry_run: bool = False) -> bool:
        """
        Process a single Go file to remove comments.
        Returns True if file was modified, False otherwise.
        """
        try:
            with open(filepath, 'r', encoding='utf-8') as f:
                original_content = f.read()
        except Exception as e:
            print(f"Error reading {filepath}: {e}")
            return False
        
        cleaned_content, comments_removed = self.remove_comments_from_content(original_content)
        
        if comments_removed == 0:
            return False
        
        if dry_run:
            print(f"Would remove {comments_removed} comments from {filepath}")
            return True
        
        try:
            with open(filepath, 'w', encoding='utf-8') as f:
                f.write(cleaned_content)
            print(f"Removed {comments_removed} comments from {filepath}")
            self.total_comments_removed += comments_removed
            return True
        except Exception as e:
            print(f"Error writing {filepath}: {e}")
            return False
    
    def process_directory(self, directory: str, dry_run: bool = False) -> None:
        """Process all Go files in directory."""
        if not os.path.isdir(directory):
            print(f"Error: {directory} is not a directory")
            return
        
        go_files = self.find_go_files(directory)
        
        if not go_files:
            print(f"No Go files found in {directory}")
            return
        
        print(f"Found {len(go_files)} Go files to process")
        
        if dry_run:
            print("DRY RUN - No files will be modified")
        
        modified_files = 0
        
        for filepath in go_files:
            if self.process_file(filepath, dry_run):
                modified_files += 1
            self.processed_files += 1
        
        print("\nSummary:")
        print(f"  Files processed: {self.processed_files}")
        print(f"  Files modified: {modified_files}")
        if not dry_run:
            print(f"  Total comments removed: {self.total_comments_removed}")


def main():
    parser = argparse.ArgumentParser(
        description="Remove comments from Go source files",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  python remove_comments.py                    # Process current directory
  python remove_comments.py /path/to/project  # Process specific directory
  python remove_comments.py --dry-run         # See what would be changed
  python remove_comments.py --no-preserve     # Remove all comments including build tags
        """
    )
    
    parser.add_argument(
        'directory',
        nargs='?',
        default='.',
        help='Directory to process (default: current directory)'
    )
    
    parser.add_argument(
        '--dry-run',
        action='store_true',
        help='Show what would be changed without modifying files'
    )
    
    parser.add_argument(
        '--no-preserve',
        action='store_true',
        help='Remove all comments including build tags and go directives'
    )
    
    args = parser.parse_args()
    
    directory = os.path.abspath(args.directory)
    
    if not os.path.exists(directory):
        print(f"Error: Directory {directory} does not exist")
        sys.exit(1)
    
    print(f"Processing Go files in: {directory}")
    
    remover = GoCommentRemover(preserve_build_tags=not args.no_preserve)
    
    try:
        remover.process_directory(directory, dry_run=args.dry_run)
    except KeyboardInterrupt:
        print("\nOperation cancelled by user")
        sys.exit(1)
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()