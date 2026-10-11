#!/bin/sh
# Consistent SQLite copy plus the photos, one dated folder per night, 14 kept.
set -eu
day=$(date +%F)
mkdir -p "/backups/$day"
sqlite3 /data/sidequestar.db ".backup /backups/$day/sidequestar.db"
tar -czf "/backups/$day/photos.tar.gz" -C /data photos 2>/dev/null || true
ls -1d /backups/*/ | sort -r | tail -n +15 | xargs -r rm -rf
