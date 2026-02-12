import sys
import re
import os
import subprocess
import time

def parse_and_create_issues(filename):
    try:
        with open(filename, 'r', encoding='utf-8') as f:
            content = f.read()
    except FileNotFoundError:
        print(f"Error: File '{filename}' not found.")
        sys.exit(1)

    # 1. Generate a label from the filename (e.g., "sprint1_plan.md" -> "sprint1_plan")
    # We strip only the file extension and use the root filename as the sprint label.
    base_name = os.path.basename(filename)
    sprint_label = os.path.splitext(base_name)[0]

    # Split the content by lines starting with "## "
    chunks = re.split(r'^## ', content, flags=re.MULTILINE)

    if len(chunks) < 2:
        print("No issues found. Ensure titles start with '## '")
        sys.exit(0)

    # Skip the first chunk (Header/Sprint Goal)
    issues_to_create = chunks[1:]

    print(f"Found {len(issues_to_create)} issues to create from '{filename}'.")
    print(f"Applying global label: '{sprint_label}'\n")
    
    for i, chunk in enumerate(issues_to_create, 1):
        parts = chunk.split('\n', 1)
        title = parts[0].strip()
        body = parts[1].strip() if len(parts) > 1 else ""
        
        if body:
            lines = body.splitlines()
            for idx in range(len(lines) - 1, -1, -1):
                if re.fullmatch(r'---\s*', lines[idx]):
                    del lines[idx]
                    break
            body = "\n".join(lines).strip()

        if not title:
            continue

        # 2. Determine Type Label (Frontend vs Backend)
        type_labels = []
        if "[Backend]" in title:
            type_labels.append("backend")
        if "[Frontend]" in title:
            type_labels.append("frontend")
        
        # Combine all labels
        # We assume the sprint label and the type label are sufficient.
        all_labels = [sprint_label] + type_labels

        print(f"Creating Issue {i}: {title}")
        print(f" > Labels: {', '.join(all_labels)}")

        # Construct the GH CLI command
        cmd = ['gh', 'issue', 'create', '--title', title, '--body', body]
        
        # Add labels flag for each label found
        for label in all_labels:
            cmd.extend(['--label', label])

        try:
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                check=True
            )
            print(f" ✅ Success: {result.stdout.strip()}")
            time.sleep(0.1) # Rate limit protection
            
        except subprocess.CalledProcessError as e:
            print(f" ❌ Failed to create issue '{title}'")
            print(f"Error: {e.stderr}")

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python create_issues.py <filename.md>")
        sys.exit(1)
    
    parse_and_create_issues(sys.argv[1])