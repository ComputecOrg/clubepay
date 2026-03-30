#!/bin/bash
# scripts/backup.sh — Automated PostgreSQL backup for ClubePay
#
# Usage: ./scripts/backup.sh
# Cron:  0 3 * * * /path/to/clubepay/scripts/backup.sh
#
# Requires: pg_dump, gzip
# Env vars: DB_USER, DB_PASSWORD, DB_HOST, DB_NAME, BACKUP_DIR

set -euo pipefail

DB_USER="${DB_USER:-clubepay}"
DB_HOST="${DB_HOST:-localhost}"
DB_NAME="${DB_NAME:-clubepay_prod}"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/clubepay}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.sql.gz"

mkdir -p "$BACKUP_DIR"

echo "[$(date)] Starting backup of ${DB_NAME}..."

PGPASSWORD="${DB_PASSWORD}" pg_dump \
  -h "$DB_HOST" \
  -U "$DB_USER" \
  -d "$DB_NAME" \
  --format=custom \
  --compress=9 \
  --file="$BACKUP_FILE"

echo "[$(date)] Backup saved to ${BACKUP_FILE}"
echo "[$(date)] Size: $(du -h "$BACKUP_FILE" | cut -f1)"

# Clean old backups
find "$BACKUP_DIR" -name "${DB_NAME}_*.sql.gz" -mtime +${RETENTION_DAYS} -delete
echo "[$(date)] Cleaned backups older than ${RETENTION_DAYS} days"

echo "[$(date)] Backup complete."
