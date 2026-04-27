import os
import re
import json

api_dir = r"d:\开源项目\GoBlog\go-blog\api"
models_dir = r"d:\开源项目\GoBlog\go-blog\models"

def parse_structs(directory):
    structs = {}
    struct_pattern = re.compile(r'type\s+([a-zA-Z0-9_]+)\s+struct\s*\{([^}]*)\}', re.MULTILINE)
    
    for root, _, files in os.walk(directory):
        for file in files:
            if file.endswith('.go'):
                filepath = os.path.join(root, file)
                with open(filepath, 'r', encoding='utf-8') as f:
                    content = f.read()
                    matches = struct_pattern.findall(content)
                    for name, body in matches:
                        fields = []
                        lines = body.strip().split('\n')
                        for line in lines:
                            line = line.strip()
                            if not line or line.startswith('//'):
                                continue
                            # Handle inline comments
                            comment = ""
                            if '//' in line:
                                parts = line.split('//', 1)
                                line = parts[0].strip()
                                comment = parts[1].strip()
                            
                            parts = line.split()
                            if len(parts) >= 1:
                                if len(parts) == 1: # embedded struct
                                    fields.append({"name": parts[0], "type": parts[0], "tags": "", "comment": comment})
                                else:
                                    # find tags
                                    tag_match = re.search(r'`([^`]+)`', line)
                                    tags = tag_match.group(1) if tag_match else ""
                                    fields.append({
                                        "name": parts[0],
                                        "type": parts[1],
                                        "tags": tags,
                                        "comment": comment
                                    })
                        structs[name] = fields
    return structs

api_structs = parse_structs(api_dir)
models_structs = parse_structs(models_dir)
api_structs.update(models_structs)

with open('structs_dump.json', 'w', encoding='utf-8') as f:
    json.dump(api_structs, f, ensure_ascii=False, indent=2)
print(f"Extracted {len(api_structs)} structs.")
