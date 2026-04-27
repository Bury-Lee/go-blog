import os
import re
import json

router_dir = "router"

routes = []
for filename in os.listdir(router_dir):
    if not filename.endswith("_router.go"):
        continue
    filepath = os.path.join(router_dir, filename)
    with open(filepath, "r", encoding="utf-8") as f:
        content = f.read()
        # Find all routes
        lines = content.split('\n')
        for line in lines:
            # simple match
            match = re.search(r'r\.(GET|POST|PUT|DELETE|PATCH)\s*\(\s*"([^"]+)"\s*,(.*)\)', line)
            if match:
                method = match.group(1)
                path = match.group(2)
                handlers_str = match.group(3)
                
                # Split by comma but ignore commas inside parens/brackets
                # Actually, the last argument is usually the handler
                parts = handlers_str.split(',')
                handler_full = parts[-1].strip().split(')')[0].strip()
                handler_name = handler_full.split('.')[-1]
                
                routes.append({
                    "method": method,
                    "path": path,
                    "handler": handler_name,
                    "file": "",
                    "req_struct": "",
                    "req_fields": [],
                    "response_info": ""
                })

with open("debug_routes.json", "w") as f:
    json.dump(routes, f, indent=2)

print("Done")
