#!/bin/bash

# Count lines in .go files
echo "Counting lines in .go files..."

total_lines=0
file_count=0

# Use find to locate all .go files
while IFS= read -r file; do
    if [[ -f "$file" ]]; then
        lines=$(wc -l < "$file")
        echo "$file: $lines lines"
        total_lines=$((total_lines + lines))
        file_count=$((file_count + 1))
    fi
done < <(find . -name "*.go" -type f)

echo "====================================="
echo "Total .go files: $file_count"
echo "Total lines in .go files: $total_lines"
