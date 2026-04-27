import json

def go_type_to_ts(go_type):
    go_type = go_type.replace('*', '').replace('[]', 'Array<')
    mapping = {
        'string': 'string',
        'int': 'number',
        'int8': 'number',
        'int16': 'number',
        'int32': 'number',
        'int64': 'number',
        'uint': 'number',
        'uint8': 'number',
        'uint16': 'number',
        'uint32': 'number',
        'uint64': 'number',
        'float32': 'number',
        'float64': 'number',
        'bool': 'boolean',
        'time.Time': 'string (ISO8601)',
        'any': 'any',
        'interface{}': 'any'
    }
    for k, v in mapping.items():
        if go_type.startswith(k):
            return v + go_type[len(k):]
    
    # if it's Array<something>, try to convert the something
    if go_type.startswith('Array<'):
        inner = go_type[6:]
        return f"Array<{mapping.get(inner, inner)}>"
        
    return go_type

def main():
    with open('structs_dump.json', 'r', encoding='utf-8') as f:
        structs = json.load(f)
    
    md_lines = ["\n\n## 8. 数据结构详解 (Data Structures)\n\n", "以下是前端交互中主要的请求和响应结构体定义（含字段、类型及校验规则）。\n"]
    
    # filter out only those ending with Request, Response, Model, etc. if needed
    # but let's just print all that have fields
    for name, fields in sorted(structs.items()):
        if not fields:
            continue
        # Skip pure controllers
        if name.endswith('Api') and len(fields) == 0:
            continue
            
        md_lines.append(f"### {name}\n")
        md_lines.append("| 字段名 (Field) | 类型 (Type) | 标签 (Tags/JSON) | 描述说明 |\n")
        md_lines.append("| --- | --- | --- | --- |\n")
        
        for field in fields:
            f_name = field['name']
            f_type = go_type_to_ts(field['type'])
            f_tags = field['tags'].replace('`', '')
            # Try to extract json or form tag for the column
            json_tag = ""
            for tag in f_tags.split():
                if tag.startswith('json:"'):
                    json_tag = tag.split('"')[1]
                elif tag.startswith('form:"'):
                    json_tag = tag.split('"')[1]
            
            display_name = json_tag if json_tag else f_name
            if json_tag == "-":
                continue # ignored in json
                
            f_comment = field['comment']
            md_lines.append(f"| `{display_name}` | `{f_type}` | `{f_tags}` | {f_comment} |\n")
            
        md_lines.append("\n")
        
    with open(r'd:\开源项目\GoBlog\go-blog\docs\前端接口文档.md', 'a', encoding='utf-8') as f:
        f.writelines(md_lines)
        
    print("Done generating markdown structures.")

if __name__ == "__main__":
    main()
