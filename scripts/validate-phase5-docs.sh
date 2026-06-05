#!/bin/bash
# 验证 Phase 5 API 文档完整性脚本 (Linux/Mac)

echo "========================================"
echo "Phase 5 API 文档验证"
echo "========================================"
echo ""

ERROR_COUNT=0

echo "[1/5] 检查 OpenAPI 规范文件..."
if [ -f "docs/api-swagger.yaml" ]; then
    echo "  ✓ docs/api-swagger.yaml 存在"
else
    echo "  ✗ docs/api-swagger.yaml 不存在"
    ((ERROR_COUNT++))
fi

if [ -f "docs/api-swagger.json" ]; then
    echo "  ✓ docs/api-swagger.json 存在"
else
    echo "  ✗ docs/api-swagger.json 不存在"
    ((ERROR_COUNT++))
fi
echo ""

echo "[2/5] 检查 Postman 集合..."
if [ -f "docs/postman/phase5-api.json" ]; then
    echo "  ✓ docs/postman/phase5-api.json 存在"
else
    echo "  ✗ docs/postman/phase5-api.json 不存在"
    ((ERROR_COUNT++))
fi
echo ""

echo "[3/5] 检查文档指南..."
if [ -f "docs/phase5-api-documentation.md" ]; then
    echo "  ✓ docs/phase5-api-documentation.md 存在"
else
    echo "  ✗ docs/phase5-api-documentation.md 不存在"
    ((ERROR_COUNT++))
fi

if [ -f "docs/phase5-api-setup-guide.md" ]; then
    echo "  ✓ docs/phase5-api-setup-guide.md 存在"
else
    echo "  ✗ docs/phase5-api-setup-guide.md 不存在"
    ((ERROR_COUNT++))
fi

if [ -f "docs/generate-swagger.md" ]; then
    echo "  ✓ docs/generate-swagger.md 存在"
else
    echo "  ✗ docs/generate-swagger.md 不存在"
    ((ERROR_COUNT++))
fi
echo ""

echo "[4/5] 检查 Swagger 注释..."
if grep -q "@Summary" internal/controller/share_controller.go; then
    echo "  ✓ share_controller.go 包含 Swagger 注释"
else
    echo "  ✗ share_controller.go 缺少 Swagger 注释"
    ((ERROR_COUNT++))
fi

if grep -q "@Summary" internal/controller/search_controller.go; then
    echo "  ✓ search_controller.go 包含 Swagger 注释"
else
    echo "  ✗ search_controller.go 缺少 Swagger 注释"
    ((ERROR_COUNT++))
fi

if grep -q "@Summary" internal/controller/recycle_controller.go; then
    echo "  ✓ recycle_controller.go 包含 Swagger 注释"
else
    echo "  ✗ recycle_controller.go 缺少 Swagger 注释"
    ((ERROR_COUNT++))
fi

if grep -q "@Summary" internal/controller/version_controller.go; then
    echo "  ✓ version_controller.go 包含 Swagger 注释"
else
    echo "  ✗ version_controller.go 缺少 Swagger 注释"
    ((ERROR_COUNT++))
fi

if grep -q "@Summary" internal/controller/collaboration_controller.go; then
    echo "  ✓ collaboration_controller.go 包含 Swagger 注释"
else
    echo "  ✗ collaboration_controller.go 缺少 Swagger 注释"
    ((ERROR_COUNT++))
fi
echo ""

echo "[5/5] 检查 main.go Swagger 配置..."
if grep -q "@title" cmd/server/main.go; then
    echo "  ✓ main.go 包含 Swagger API 信息"
else
    echo "  ✗ main.go 缺少 Swagger API 信息"
    ((ERROR_COUNT++))
fi

if grep -q "ginSwagger" cmd/server/main.go; then
    echo "  ✓ main.go 已注册 Swagger UI 路由"
else
    echo "  ✗ main.go 未注册 Swagger UI 路由"
    ((ERROR_COUNT++))
fi
echo ""

echo "========================================"
if [ $ERROR_COUNT -eq 0 ]; then
    echo "✓ 所有文档检查通过！"
    echo ""
    echo "下一步:"
    echo "  1. 运行 ./scripts/generate-swagger.sh 生成 Swagger UI"
    echo "  2. 启动服务器: go run cmd/server/main.go"
    echo "  3. 访问 http://localhost:8080/swagger/index.html"
    echo ""
    exit 0
else
    echo "✗ 发现 $ERROR_COUNT 个问题"
    echo ""
    echo "请检查上述错误并修复"
    echo ""
    exit 1
fi
