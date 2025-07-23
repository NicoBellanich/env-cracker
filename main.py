import re
import os

def parse_env_file(file_path):
    # Initialize list to store file entries
    files = []
    current_file = None
    current_content = []
    in_content = False
    metadata = {}

    # Regular expressions for metadata and markers
    metadata_pattern = re.compile(r'^(\w+)/(.+)$')
    marker_pattern = re.compile(r'^(DC%%STAM4|\*\*%%KEYB4|\*\*%%DOCU)$')
    sig_pattern = re.compile(r'^_SIG/D\.C\.$')
    xml_start_pattern = re.compile(r'^\<\?xml')

    with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
        for line in f:
            line = line.strip()
            if not line:
                continue

            # Check for file start markers
            if marker_pattern.match(line):
                if current_file:
                    # Save previous file
                    current_file['content'] = '\n'.join(current_content).strip()
                    files.append(current_file)
                    current_content = []
                    in_content = False
                    metadata = {}
                current_file = {'marker': line, 'metadata': metadata}
                continue

            # Check for metadata
            if current_file and not in_content:
                metadata_match = metadata_pattern.match(line)
                if metadata_match:
                    key, value = metadata_match.groups()
                    metadata[key] = value
                    continue
                if sig_pattern.match(line):
                    in_content = True
                    continue

            # Collect content
            if current_file and in_content:
                current_content.append(line)

        # Save the last file
        if current_file:
            current_file['content'] = '\n'.join(current_content).strip()
            files.append(current_file)

    return files

def save_files(files, output_dir='extracted_files'):
    # Create output directory if it doesn't exist
    os.makedirs(output_dir, exist_ok=True)

    for file_entry in files:
        metadata = file_entry['metadata']
        filename = metadata.get('FILENAME', 'unknown_file')
        ext = metadata.get('EXT', '')
        content = file_entry['content']

        # Construct output file path
        output_path = os.path.join(output_dir, filename)

        # Write content to file
        if 'TYPE' in metadata and 'IMAGE' in metadata['TYPE']:
            # For images, write as text due to strings output limitations
            with open(output_path + '.txt', 'w', encoding='utf-8') as f:
                f.write(content)
        else:
            with open(output_path, 'w', encoding='utf-8') as f:
                f.write(content)

        # Save metadata as a separate .meta.txt file
        with open(os.path.join(output_dir, f"{filename}.meta.txt"), 'w', encoding='utf-8') as f:
            for key, value in metadata.items():
                f.write(f"{key}: {value}\n")
            f.write(f"MARKER: {file_entry['marker']}\n")

def main():
    file_path = 'strings.txt'
    output_dir = 'extracted_files'

    # Parse the file
    files = parse_env_file(file_path)
    
    # Save extracted files and metadata
    save_files(files, output_dir)

    # Print summary
    print(f"Extracted {len(files)} files:")
    for file_entry in files:
        filename = file_entry['metadata'].get('FILENAME', 'unknown_file')
        file_type = file_entry['metadata'].get('TYPE', 'unknown_type')
        print(f"- {filename} ({file_type})")

if __name__ == "__main__":
    main()