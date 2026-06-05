#!/bin/bash

# 星云盘 V2 性能压测脚本
# 使用 Apache Bench (ab) 进行 API 性能测试

set -e

# 配置
BASE_URL="http://localhost:8080"
API_BASE="${BASE_URL}/api/v1"
RESULTS_DIR="./benchmark-results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
REPORT_FILE="${RESULTS_DIR}/benchmark_${TIMESTAMP}.txt"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 创建结果目录
mkdir -p ${RESULTS_DIR}

# 检查 ab 是否安装
if ! command -v ab &> /dev/null; then
    echo -e "${RED}错误: Apache Bench (ab) 未安装${NC}"
    echo "请安装 Apache Bench:"
    echo "  Ubuntu/Debian: sudo apt-get install apache2-utils"
    echo "  CentOS/RHEL: sudo yum install httpd-tools"
    echo "  macOS: brew install httpd"
    exit 1
fi

# 检查服务器是否运行
echo -e "${YELLOW}检查服务器状态...${NC}"
if ! curl -s "${BASE_URL}/health" > /dev/null; then
    echo -e "${RED}错误: 服务器未运行或无法访问 ${BASE_URL}${NC}"
    exit 1
fi
echo -e "${GREEN}✓ 服务器运行正常${NC}"
echo ""

# 初始化报告文件
cat > ${REPORT_FILE} << EOF
========================================
星云盘 V2 性能压测报告
========================================
测试时间: $(date '+%Y-%m-%d %H:%M:%S')
服务器地址: ${BASE_URL}
测试工具: Apache Bench (ab)
========================================

EOF

echo -e "${GREEN}开始性能压测...${NC}"
echo "结果将保存到: ${REPORT_FILE}"
echo ""

# 函数: 执行压测并记录结果
run_benchmark() {
    local test_name=$1
    local url=$2
    local requests=$3
    local concurrency=$4
    local method=$5
    local data_file=$6
    local content_type=$7
    local headers=$8
    
    echo -e "${YELLOW}测试: ${test_name}${NC}"
    echo "  URL: ${url}"
    echo "  请求数: ${requests}, 并发数: ${concurrency}"
    
    # 记录到报告
    echo "" >> ${REPORT_FILE}
    echo "========================================" >> ${REPORT_FILE}
    echo "测试: ${test_name}" >> ${REPORT_FILE}
    echo "========================================" >> ${REPORT_FILE}
    echo "URL: ${url}" >> ${REPORT_FILE}
    echo "方法: ${method}" >> ${REPORT_FILE}
    echo "请求数: ${requests}" >> ${REPORT_FILE}
    echo "并发数: ${concurrency}" >> ${REPORT_FILE}
    echo "" >> ${REPORT_FILE}
    
    # 构建 ab 命令
    local ab_cmd="ab -n ${requests} -c ${concurrency}"
    
    if [ ! -z "${content_type}" ]; then
        ab_cmd="${ab_cmd} -T ${content_type}"
    fi
    
    if [ ! -z "${headers}" ]; then
        ab_cmd="${ab_cmd} -H \"${headers}\""
    fi
    
    if [ "${method}" == "POST" ] && [ ! -z "${data_file}" ]; then
        ab_cmd="${ab_cmd} -p ${data_file}"
    fi
    
    ab_cmd="${ab_cmd} ${url}"
    
    # 执行压测
    eval ${ab_cmd} >> ${REPORT_FILE} 2>&1
    
    # 提取关键指标
    local avg_time=$(grep "Time per request:" ${REPORT_FILE} | tail -1 | awk '{print $4}')
    local requests_per_sec=$(grep "Requests per second:" ${REPORT_FILE} | tail -1 | awk '{print $4}')
    local failed=$(grep "Failed requests:" ${REPORT_FILE} | tail -1 | awk '{print $3}')
    
    echo -e "  ${GREEN}✓ 完成${NC}"
    echo "    平均响应时间: ${avg_time} ms"
    echo "    吞吐量: ${requests_per_sec} req/s"
    echo "    失败请求: ${failed}"
    echo ""
}

# ========================================
# 测试 1: 健康检查 API
# ========================================
run_benchmark \
    "健康检查 API" \
    "${BASE_URL}/health" \
    1000 \
    10 \
    "GET"

# ========================================
# 测试 2: 用户注册 API
# ========================================
run_benchmark \
    "用户注册 API" \
    "${API_BASE}/user/register" \
    100 \
    5 \
    "POST" \
    "scripts/test-data/register.json" \
    "application/json"

# ========================================
# 测试 3: 用户登录 API
# ========================================
run_benchmark \
    "用户登录 API" \
    "${API_BASE}/user/login" \
    1000 \
    10 \
    "POST" \
    "scripts/test-data/login.json" \
    "application/json"

# ========================================
# 测试 4: 分片上传初始化 API
# ========================================
run_benchmark \
    "分片上传初始化 API" \
    "${API_BASE}/files/multipart/init" \
    100 \
    5 \
    "POST" \
    "scripts/test-data/multipart-init.json" \
    "application/json"

# ========================================
# 生成摘要
# ========================================
echo "" >> ${REPORT_FILE}
echo "========================================" >> ${REPORT_FILE}
echo "测试摘要" >> ${REPORT_FILE}
echo "========================================" >> ${REPORT_FILE}
echo "所有测试已完成" >> ${REPORT_FILE}
echo "详细结果请查看上方各项测试数据" >> ${REPORT_FILE}
echo "" >> ${REPORT_FILE}

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}性能压测完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo "报告文件: ${REPORT_FILE}"
echo ""
echo "查看报告:"
echo "  cat ${REPORT_FILE}"
echo ""
