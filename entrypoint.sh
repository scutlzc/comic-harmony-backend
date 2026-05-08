#!/bin/sh
# entrypoint.sh - 后端容器启动入口
# 自动运行数据库迁移后启动服务

set -e

echo "[entrypoint] running migrations..."
for f in /migrations/*.sql; do
  echo "  - $(basename $f)"
  PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f "$f" 2>/dev/null || true
done

echo "[entrypoint] starting server..."
exec /server
