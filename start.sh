#!/bin/bash
# InterviewHub 服务管理脚本
#
# 用法:
#   ./start.sh         启动前后端
#   ./start.sh stop    停止前后端
#   ./start.sh restart 重启前后端
#   ./start.sh backend  仅启动后端
#   ./start.sh frontend 仅启动前端

set -e

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
BACKEND_DIR="$PROJECT_DIR/backend"
FRONTEND_DIR="$PROJECT_DIR/frontend"
BACKEND_PORT=17001
FRONTEND_PORT=17000
LOG_DIR="$PROJECT_DIR/.logs"
BACKEND_LOG="$LOG_DIR/backend.log"
FRONTEND_LOG="$LOG_DIR/frontend.log"
BACKEND_PID_FILE="$LOG_DIR/backend.pid"
FRONTEND_PID_FILE="$LOG_DIR/frontend.pid"

mkdir -p "$LOG_DIR"

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()  { echo -e "${GREEN}[INFO]${NC}  $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC}  $1"; }
err()  { echo -e "${RED}[ERROR]${NC} $1"; }

# 检查端口是否被占用
port_in_use() {
    lsof -i ":$1" -sTCP:LISTEN -t 2>/dev/null || return 1
}

# 停止后端
stop_backend() {
    if [ -f "$BACKEND_PID_FILE" ]; then
        PID=$(cat "$BACKEND_PID_FILE")
        if kill -0 "$PID" 2>/dev/null; then
            kill "$PID" 2>/dev/null
            sleep 1
            if kill -0 "$PID" 2>/dev/null; then
                kill -9 "$PID" 2>/dev/null
            fi
            log "后端已停止 (PID: $PID)"
        fi
        rm -f "$BACKEND_PID_FILE"
    fi
    # 兜底：通过端口杀
    PID=$(lsof -i ":$BACKEND_PORT" -sTCP:LISTEN -t 2>/dev/null)
    if [ -n "$PID" ]; then
        kill -9 "$PID" 2>/dev/null
        log "后端端口 $BACKEND_PORT 已释放 (PID: $PID)"
    fi
}

# 停止前端
stop_frontend() {
    if [ -f "$FRONTEND_PID_FILE" ]; then
        PID=$(cat "$FRONTEND_PID_FILE")
        if kill -0 "$PID" 2>/dev/null; then
            kill "$PID" 2>/dev/null
            sleep 1
            if kill -0 "$PID" 2>/dev/null; then
                kill -9 "$PID" 2>/dev/null
            fi
            log "前端已停止 (PID: $PID)"
        fi
        rm -f "$FRONTEND_PID_FILE"
    fi
    PID=$(lsof -i ":$FRONTEND_PORT" -sTCP:LISTEN -t 2>/dev/null)
    if [ -n "$PID" ]; then
        kill -9 "$PID" 2>/dev/null
        log "前端端口 $FRONTEND_PORT 已释放 (PID: $PID)"
    fi
}

# 启动后端
start_backend() {
    if port_in_use "$BACKEND_PORT"; then
        warn "端口 $BACKEND_PORT 已被占用"
        return 1
    fi

    log "启动后端..."
    cd "$BACKEND_DIR"
    nohup go run cmd/http-server/main.go >> "$BACKEND_LOG" 2>&1 &
    echo $! > "$BACKEND_PID_FILE"

    # 等待启动
    for i in $(seq 1 30); do
        if curl -s http://localhost:$BACKEND_PORT/api/health > /dev/null 2>&1; then
            log "后端启动成功 http://localhost:$BACKEND_PORT"
            return 0
        fi
        sleep 1
    done
    err "后端启动超时，查看日志: tail -f $BACKEND_LOG"
    return 1
}

# 启动前端
start_frontend() {
    if port_in_use "$FRONTEND_PORT"; then
        warn "端口 $FRONTEND_PORT 已被占用"
        return 1
    fi

    log "启动前端..."
    cd "$FRONTEND_DIR"
    nohup npx vite --port "$FRONTEND_PORT" >> "$FRONTEND_LOG" 2>&1 &
    echo $! > "$FRONTEND_PID_FILE"

    for i in $(seq 1 15); do
        if curl -s http://localhost:$FRONTEND_PORT > /dev/null 2>&1; then
            log "前端启动成功 http://localhost:$FRONTEND_PORT"
            return 0
        fi
        sleep 1
    done
    err "前端启动超时，查看日志: tail -f $FRONTEND_LOG"
    return 1
}

# 状态
status_all() {
    echo "=== InterviewHub 服务状态 ==="
    if port_in_use "$BACKEND_PORT"; then
        echo -e "  后端:  ${GREEN}运行中${NC}  http://localhost:$BACKEND_PORT"
    else
        echo -e "  后端:  ${RED}未启动${NC}"
    fi
    if port_in_use "$FRONTEND_PORT"; then
        echo -e "  前端:  ${GREEN}运行中${NC}  http://localhost:$FRONTEND_PORT"
    else
        echo -e "  前端:  ${RED}未启动${NC}"
    fi
    echo ""
    echo "日志: $LOG_DIR/"
}

# 主逻辑
case "${1:-start}" in
    start)
        start_backend
        start_frontend
        sleep 1
        status_all
        ;;
    stop)
        stop_backend
        stop_frontend
        log "所有服务已停止"
        ;;
    restart)
        stop_backend
        stop_frontend
        sleep 1
        start_backend
        start_frontend
        sleep 1
        status_all
        ;;
    backend)
        case "${2:-start}" in
            start)   start_backend ;;
            stop)    stop_backend ;;
            restart) stop_backend; sleep 1; start_backend ;;
            *)       err "用法: $0 backend {start|stop|restart}" ;;
        esac
        ;;
    frontend)
        case "${2:-start}" in
            start)   start_frontend ;;
            stop)    stop_frontend ;;
            restart) stop_frontend; sleep 1; start_frontend ;;
            *)       err "用法: $0 frontend {start|stop|restart}" ;;
        esac
        ;;
    status)
        status_all
        ;;
    logs)
        echo "=== 后端日志 ($BACKEND_LOG) ==="
        tail -20 "$BACKEND_LOG" 2>/dev/null || echo "(空)"
        echo ""
        echo "=== 前端日志 ($FRONTEND_LOG) ==="
        tail -20 "$FRONTEND_LOG" 2>/dev/null || echo "(空)"
        ;;
    *)
        echo "用法: $0 {start|stop|restart|status|logs|backend|frontend}"
        echo ""
        echo "  start          启动前后端"
        echo "  stop           停止前后端"
        echo "  restart        重启前后端"
        echo "  status         查看服务状态"
        echo "  logs           查看最近日志"
        echo "  backend  {start|stop|restart}  单独管理后端"
        echo "  frontend {start|stop|restart}  单独管理前端"
        ;;
esac
