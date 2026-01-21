#!/bin/bash

# 开发模式启动脚本 - 前后端分离运行
# 前端: http://localhost:5173 (支持热重载)
# 后端: http://localhost:9090

echo "🚀 启动开发环境..."
echo ""

# 加载环境变量
if [ -f .env ]; then
    echo "📝 加载环境变量..."
    export $(cat .env | grep -v '^#' | xargs)
else
    echo "⚠️  警告: 未找到 .env 文件"
    echo "请复制 .env.example 为 .env 并配置数据库连接信息："
    echo "  cp .env.example .env"
    echo ""
    echo "或者确保本地 PostgreSQL 已启动，使用默认配置："
    export POSTGRES_HOST=localhost
    export POSTGRES_PORT=5432
    export POSTGRES_USER=fasaxi
    export POSTGRES_PASSWORD=fasaxi_password
    export POSTGRES_DB=fasaxi_linker
    echo "  数据库: postgresql://$POSTGRES_USER@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB"
    echo ""
fi

# 检查是否安装了必要的工具
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未安装 Go"
    exit 1
fi

if ! command -v pnpm &> /dev/null; then
    echo "❌ 错误: 未安装 pnpm"
    echo "请运行: npm install -g pnpm"
    exit 1
fi

# 检查 PostgreSQL 是否可访问
echo "🔍 检查数据库连接..."
if command -v psql &> /dev/null; then
    if ! PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER -d postgres -c '\q' 2>/dev/null; then
        echo "⚠️  警告: 无法连接到 PostgreSQL"
        echo "请确保 PostgreSQL 已启动并且配置正确"
        echo ""
        echo "快速启动 PostgreSQL (使用 Docker):"
        echo "  docker run -d --name fasaxi-postgres \\"
        echo "    -e POSTGRES_USER=$POSTGRES_USER \\"
        echo "    -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \\"
        echo "    -e POSTGRES_DB=$POSTGRES_DB \\"
        echo "    -p $POSTGRES_PORT:5432 \\"
        echo "    postgres:17"
        echo ""
        read -p "是否继续启动？(y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    else
        echo "✅ 数据库连接正常"
    fi
else
    echo "💡 提示: 未安装 psql，跳过数据库连接检查"
fi
echo ""

# 启动后端服务器
echo "📦 启动后端服务器 (端口 9090)..."
cd server
go run cmd/server/main.go &
BACKEND_PID=$!
cd ..

# 等待后端启动
sleep 2

# 启动前端开发服务器
echo "🎨 启动前端开发服务器 (端口 5173)..."
cd web
pnpm install
pnpm dev &
FRONTEND_PID=$!
cd ..

echo ""
echo "✅ 开发环境已启动！"
echo ""
echo "📍 访问地址:"
echo "   前端: http://localhost:5173"
echo "   后端: http://localhost:9090"
echo ""
echo "💡 提示:"
echo "   - 前端支持热重载，修改代码会自动刷新"
echo "   - 按 Ctrl+C 停止所有服务"
echo ""

# 捕获 Ctrl+C 信号
trap "echo ''; echo '🛑 正在停止服务...'; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit 0" INT

# 等待进程
wait
