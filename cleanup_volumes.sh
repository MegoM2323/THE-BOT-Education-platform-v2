#!/bin/bash
set -e

SERVER="mg@5.129.249.206"

echo "=== Docker Volumes Cleanup Script ==="
echo "Server: $SERVER"
echo ""

# Volumes to DELETE
VOLUMES_TO_DELETE=(
  "the_bot_v3_postgres_data"
  "the_bot_v3_backend_uploads"
  "the_bot_platform_celery_beat_schedule"
  "the_bot_platform_celery_worker_1"
  "the_bot_platform_celery_worker_2"
  "the_bot_platform_redis_data"
  "the_bot_platform_nginx_cache"
  "the_bot_platform_nginx_logs"
  "the_bot_platform_nginx_temp"
  "the_bot_platform_frontend_build"
  "the_bot_platform_frontend_node_modules"
  "the_bot_platform_backend_logs"
  "the_bot_platform_backend_media"
  "the_bot_platform_backend_static"
  "the_bot_platform_backend_temp"
  "the_bot_platform_postgres_temp"
)

# Volumes to PRESERVE
PRESERVE_VOLUMES=(
  "the_bot_platform_postgres_data"
  "the_bot_platform_backend_uploads"
  "backend_uploads"
)

echo "Volumes marked for deletion:"
for vol in "${VOLUMES_TO_DELETE[@]}"; do
  echo "  - $vol"
done
echo ""

echo "Volumes to PRESERVE (will NOT be deleted):"
for vol in "${PRESERVE_VOLUMES[@]}"; do
  echo "  - $vol (PROTECTED)"
done
echo ""

read -p "Continue with cleanup? (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
  echo "Cleanup cancelled."
  exit 0
fi

echo ""
echo "=== Listing existing volumes on server ==="
ssh "$SERVER" "docker volume ls"

echo ""
echo "=== Checking for containers using volumes ==="
for vol in "${VOLUMES_TO_DELETE[@]}"; do
  echo "Checking volume: $vol"
  ssh "$SERVER" "docker ps -a --filter volume=$vol --format '{{.Names}}' | head -5" || true
done

echo ""
echo "=== Starting volume removal ==="

DELETED_COUNT=0
FAILED_COUNT=0

for vol in "${VOLUMES_TO_DELETE[@]}"; do
  echo "Attempting to remove: $vol"

  # Check if volume exists
  if ssh "$SERVER" "docker volume inspect $vol >/dev/null 2>&1"; then
    # Try to remove
    if ssh "$SERVER" "docker volume rm $vol 2>&1"; then
      echo "  ✓ Removed: $vol"
      ((DELETED_COUNT++))
    else
      echo "  ✗ Failed to remove: $vol (may be in use)"
      ((FAILED_COUNT++))
    fi
  else
    echo "  ⊘ Volume does not exist: $vol"
  fi
  echo ""
done

echo "=== Cleanup Summary ==="
echo "Deleted: $DELETED_COUNT"
echo "Failed: $FAILED_COUNT"
echo ""

echo "=== Remaining volumes on server ==="
ssh "$SERVER" "docker volume ls"

echo ""
echo "=== Protected volumes verification ==="
for vol in "${PRESERVE_VOLUMES[@]}"; do
  if ssh "$SERVER" "docker volume inspect $vol >/dev/null 2>&1"; then
    echo "  ✓ Protected volume still exists: $vol"
  else
    echo "  ⚠ WARNING: Protected volume NOT found: $vol"
  fi
done

echo ""
echo "=== Cleanup complete ==="
