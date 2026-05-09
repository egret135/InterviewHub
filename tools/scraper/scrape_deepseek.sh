#!/bin/bash
# 便捷命令行工具：抓取 DeepSeek 聊天分享页面
#
# 用法:
#   ./scrape_deepseek.sh <deepseek_share_url> [output_dir]
#
# 示例:
#   ./scrape_deepseek.sh https://chat.deepseek.com/share/abc123
#   ./scrape_deepseek.sh https://chat.deepseek.com/share/abc123 ./my_output

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
URL="${1:?Usage: $0 <deepseek_share_url> [output_dir]}"
OUTDIR="${2:-$SCRIPT_DIR/deepseek_output}"

node "$SCRIPT_DIR/scrape_deepseek.mjs" --url "$URL" --outdir "$OUTDIR"

echo ""
echo "=== 输出文件 ==="
ls -lh "$OUTDIR"/
