import json

api_dir = r"d:\开源项目\GoBlog\go-blog\api"
models_dir = r"d:\开源项目\GoBlog\go-blog\models"

api_structs = parse_structs(api_dir)
models_structs = parse_structs(models_dir)

api_structs.update(models_structs)

with open('structs_dump.json', 'w', encoding='utf-8') as f:
    json.dump(api_structs, f, ensure_ascii=False, indent=2)
print(f"Extracted {len(api_structs)} structs.")
