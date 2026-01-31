#!/bin/bash
# Open Shelley 启动脚本

cd /home/exedev/002/openshelley

exec /home/exedev/002/shelley_linux_amd64 \
    -db ./shelley.db \
    -config ./shelley.json \
    -default-model claude-sonnet-4-20250514 \
    serve -port 9001
