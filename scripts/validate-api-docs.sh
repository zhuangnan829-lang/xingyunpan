#!/bin/bash

# API 文档验证脚本
# 用于验证 Swagger/OpenAPI 文档的有效性

echo "=========================================="
echo "API 文档验证脚本"
echo "=========================================="
echo ""

# 检查文件是否存在
echo "1. 检查文档文件..."
if [ -f "docs/api-swagger.yaml" ]; then
    echo "   ✅ docs/api-swagger.yaml 存在"
else
    echo "   ❌ docs/api-swagger.yaml 不存在"
    exit 1
fi

if [ -f "docs/api-swagger.json" ]; then
    echo "   ✅ docs/api-swagger.json 存在"
else
    echo "   ❌ docs/api-swagger.json 不存在"
    exit 1
fi

if [ -f "docs/API-DOCUMENTATION.md" ]; then
    echo "   ✅ docs/API-DOCUMENTATION.md 存在"
else
    echo "   ❌ docs/API-DOCUMENTATION.md 不存在"
    exit 1
fi

echo ""

# 检查文件大小
echo "2. 检查文档大小..."
yaml_size=$(wc -c < "docs/api-swagger.yaml")
json_size=$(wc -c < "docs/api-swagger.json")
doc_size=$(wc -c < "docs/API-DOCUMENTATION.md")

echo "   - api-swagger.yaml: $yaml_size bytes"
echo "   - api-swagger.json: $json_size bytes"
echo "   - API-DOCUMENTATION.md: $doc_size bytes"

if [ $yaml_size -gt 10000 ]; then
    echo "   ✅ YAML 文档大小正常"
else
    echo "   ⚠️  YAML 文档可能不完整"
fi

if [ $json_size -gt 5000 ]; then
    echo "   ✅ JSON 文档大小正常"
else
    echo "   ⚠️  JSON 文档可能不完整"
fi

echo ""

# 检查 YAML 语法（如果安装了 yamllint）
echo "3. 检查 YAML 语法..."
if command -v yamllint &> /dev/null; then
    if yamllint docs/api-swagger.yaml; then
        echo "   ✅ YAML 语法正确"
    else
        echo "   ❌ YAML 语法错误"
        exit 1
    fi
else
    echo "   ⚠️  yamllint 未安装，跳过语法检查"
    echo "   提示: 安装 yamllint 以验证 YAML 语法"
    echo "   命令: pip install yamllint"
fi

echo ""

# 检查 JSON 语法
echo "4. 检查 JSON 语法..."
if command -v jq &> /dev/null; then
    if jq empty docs/api-swagger.json 2>/dev/null; then
        echo "   ✅ JSON 语法正确"
    else
        echo "   ❌ JSON 语法错误"
        exit 1
    fi
else
    echo "   ⚠️  jq 未安装，跳过语法检查"
    echo "   提示: 安装 jq 以验证 JSON 语法"
    echo "   命令: brew install jq (Mac) 或 apt-get install jq (Linux)"
fi

echo ""

# 检查关键字段
echo "5. 检查关键字段..."
if grep -q "openapi: 3.0.3" docs/api-swagger.yaml; then
    echo "   ✅ OpenAPI 版本正确"
else
    echo "   ❌ OpenAPI 版本不正确"
fi

if grep -q "title: 星云盘 V2 API" docs/api-swagger.yaml; then
    echo "   ✅ API 标题正确"
else
    echo "   ❌ API 标题不正确"
fi

if grep -q "/api/v1/auth/register" docs/api-swagger.yaml; then
    echo "   ✅ 包含注册接口"
else
    echo "   ❌ 缺少注册接口"
fi

if grep -q "/api/v1/file/upload" docs/api-swagger.yaml; then
    echo "   ✅ 包含文件上传接口"
else
    echo "   ❌ 缺少文件上传接口"
fi

echo ""

# 统计接口数量
echo "6. 统计接口数量..."
api_count=$(grep -c "operationId:" docs/api-swagger.yaml)
echo "   - 总接口数: $api_count"

if [ $api_count -ge 13 ]; then
    echo "   ✅ 接口数量正常（预期 13 个）"
else
    echo "   ⚠️  接口数量可能不完整（预期 13 个）"
fi

echo ""

# 在线验证（可选）
echo "7. 在线验证建议..."
echo "   可以使用以下工具在线验证 API 文档:"
echo "   - Swagger Editor: https://editor.swagger.io/"
echo "   - Swagger Validator: https://validator.swagger.io/validator/debug"
echo ""
echo "   验证命令:"
echo "   curl -X POST https://validator.swagger.io/validator/debug \\"
echo "     -H 'Content-Type: application/yaml' \\"
echo "     --data-binary @docs/api-swagger.yaml"

echo ""
echo "=========================================="
echo "✅ API 文档验证完成!"
echo "=========================================="
echo ""
echo "查看文档:"
echo "  - YAML: docs/api-swagger.yaml"
echo "  - JSON: docs/api-swagger.json"
echo "  - 使用指南: docs/API-DOCUMENTATION.md"
echo ""
echo "在线查看:"
echo "  1. 访问 https://editor.swagger.io/"
echo "  2. 导入 docs/api-swagger.yaml"
echo "  3. 即可查看和测试 API"
echo ""
