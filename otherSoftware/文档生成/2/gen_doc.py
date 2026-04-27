import os
import re

router_dir = "router"
api_dir = "api"

routes = []
# Parse routers
for filename in os.listdir(router_dir):
    if not filename.endswith("_router.go"):
        continue
    filepath = os.path.join(router_dir, filename)
    with open(filepath, "r", encoding="utf-8") as f:
        content = f.read()
        for line in content.split('\n'):
            match = re.search(r'r\.(GET|POST|PUT|DELETE|PATCH)\s*\(\s*"([^"]+)"\s*,(.*)\)', line)
            if match:
                method = match.group(1)
                path = match.group(2)
                handlers_str = match.group(3)
                
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

# Map handlers to structs
for root, dirs, files in os.walk(api_dir):
    for file in files:
        if not file.endswith(".go"):
            continue
        filepath = os.path.join(root, file)
        with open(filepath, "r", encoding="utf-8") as f:
            content = f.read()
            for route in routes:
                handler_name = route["handler"]
                # Match func (xxx) HandlerName(c *gin.Context)
                func_pattern = re.compile(r'func\s*\([^)]+\)\s*' + handler_name + r'\s*\(')
                if func_pattern.search(content):
                    route["file"] = filepath.replace('\\', '/')
                    func_body_match = re.search(r'func\s*\([^)]+\)\s*' + handler_name + r'\s*\([\s\S]*?(?=\nfunc|$)', content)
                    if func_body_match:
                        func_body = func_body_match.group(0)
                        
                        # Find ShouldBind or ShouldBindJSON
                        bind_match = re.search(r'ShouldBind[a-zA-Z]*\s*\(&([a-zA-Z0-9_]+)\)', func_body)
                        if bind_match:
                            var_name = bind_match.group(1)
                            # Find var declaration: var var_name StructName
                            var_pattern = re.search(r'var\s+' + var_name + r'\s+([a-zA-Z0-9_\.]+)', func_body)
                            if var_pattern:
                                struct_name = var_pattern.group(1)
                                if "." in struct_name:
                                    struct_name = struct_name.split(".")[-1]
                                route["req_struct"] = struct_name
                                
                                # Search for struct definition in the same file or in common models
                                struct_def_pattern = re.search(r'type\s+' + struct_name + r'\s+struct\s*{([^}]+)}', content)
                                if struct_def_pattern:
                                    fields_str = struct_def_pattern.group(1)
                                    fields = [l.strip() for l in fields_str.split('\n') if l.strip() and not l.strip().startswith('//')]
                                    route["req_fields"] = fields
                                else:
                                    route["req_fields"] = [f"// 结构体 {struct_name} 定义在其他文件 (可能在 models 包中)"]
                        
                        # Find response calls: response.OkWithData(..., c) or response.OkWithMsg(..., c)
                        resp_match = re.search(r'response\.(Ok[a-zA-Z0-9_]*)\s*\(([^,]+)', func_body)
                        if resp_match:
                            resp_type = resp_match.group(1)
                            resp_data = resp_match.group(2).strip()
                            route["response_info"] = f"{resp_type}({resp_data})"

# Generate Markdown
md = "# 前端接口文档\n\n"
md += "> 本文档由分析代码路由和控制层自动生成，包含了当前所有的 API 接口、请求参数与响应。\n\n"

grouped = {}
for route in routes:
    parts = [p for p in route['path'].split('/') if p]
    group = parts[0] if parts else "root"
    if group not in grouped:
        grouped[group] = []
    grouped[group].append(route)

for group, rts in grouped.items():
    md += f"## 模块: `{group}`\n\n"
    for route in rts:
        md += f"### {route['method']} `/api{route['path']}`\n"
        md += f"- **处理函数**: `{route['handler']}`\n"
        if route['file']:
            md += f"- **所在文件**: `{route['file']}`\n"
        
        md += "- **请求参数**:\n"
        if route["req_struct"]:
            md += f"  - 绑定结构体: `{route['req_struct']}`\n"
            md += "  ```go\n"
            for f in route["req_fields"]:
                md += "  " + f + "\n"
            md += "  ```\n"
        else:
            md += "  - *无特定结构体绑定（可能是无参或路径参数/Query参数）*\n"
        
        md += "- **成功响应**:\n"
        if route["response_info"]:
            md += f"  - `{route['response_info']}`\n"
        else:
            md += "  - *返回默认成功或未检测到标准 response 响应*\n"
        
        md += "\n---\n"

with open("docs/前端接口文档-整理版.md", "w", encoding="utf-8") as f:
    f.write(md)

print("Markdown generated successfully!")
