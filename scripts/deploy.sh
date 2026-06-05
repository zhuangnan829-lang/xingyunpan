#!/bin/bash

set -e

echo "=== Deploying Xingyunpan V2 ==="
echo "Deployment started at $(date)"
echo ""

# Configuration
APP_DIR="/opt/xingyunpan-v2"
BACKUP_DIR="/backup/xingyunpan"
LOG_FILE="/data/xingyunpan/logs/deploy.log"

# Create log directory if not exists
mkdir -p "$(dirname "$LOG_FILE")"

# Log function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log "=== Starting Deployment ==="

# Step 1: Backup current version
log "Step 1: Backing up current version..."
if [ -f "$APP_DIR/bin/server" ]; then
    cp "$APP_DIR/bin/server" "$APP_DIR/bin/server.backup"
    log "✓ Server binary backed up"
else
    log "⚠ No existing server binary found"
fi

if [ -f "$APP_DIR/bin/worker" ]; then
    cp "$APP_DIR/bin/worker" "$APP_DIR/bin/worker.backup"
    log "✓ Worker binary backed up"
else
    log "⚠ No existing worker binary found"
fi

# Step 2: Backup database
log "Step 2: Backing up database..."
mkdir -p "$BACKUP_DIR"
BACKUP_FILE="$BACKUP_DIR/pre-deploy-$(date +%Y%m%d-%H%M%S).sql"

if command -v mysqldump &> /dev/null; then
    if mysqldump -u xingyunpan -p"${DB_PASSWORD}" xingyunpan > "$BACKUP_FILE" 2>/dev/null; then
        gzip "$BACKUP_FILE"
        log "✓ Database backed up to $BACKUP_FILE.gz"
    else
        log "⚠ Database backup failed (continuing anyway)"
    fi
else
    log "⚠ mysqldump not found, skipping database backup"
fi

# Step 3: Stop services
log "Step 3: Stopping services..."
systemctl stop xingyunpan-server || log "⚠ Failed to stop server service"
systemctl stop xingyunpan-worker || log "⚠ Failed to stop worker service"
log "✓ Services stopped"

# Wait for services to fully stop
sleep 2

# Step 4: Deploy new binaries (already copied by GitHub Actions)
log "Step 4: Verifying new binaries..."
if [ -f "$APP_DIR/bin/server" ] && [ -f "$APP_DIR/bin/worker" ]; then
    chmod +x "$APP_DIR/bin/server"
    chmod +x "$APP_DIR/bin/worker"
    log "✓ New binaries verified and made executable"
else
    log "✗ New binaries not found!"
    log "Rolling back..."
    
    if [ -f "$APP_DIR/bin/server.backup" ]; then
        cp "$APP_DIR/bin/server.backup" "$APP_DIR/bin/server"
        cp "$APP_DIR/bin/worker.backup" "$APP_DIR/bin/worker"
    fi
    
    systemctl start xingyunpan-server
    systemctl start xingyunpan-worker
    
    log "✗ Deployment failed - rolled back to previous version"
    exit 1
fi

# Step 5: Run database migrations (if needed)
log "Step 5: Running database migrations..."
cd "$APP_DIR"
if [ -f "scripts/migrate.go" ]; then
    if go run scripts/migrate.go 2>&1 | tee -a "$LOG_FILE"; then
        log "✓ Database migrations completed"
    else
        log "⚠ Database migrations failed (continuing anyway)"
    fi
else
    log "⚠ Migration script not found, skipping"
fi

# Step 6: Start services
log "Step 6: Starting services..."
systemctl start xingyunpan-server
systemctl start xingyunpan-worker
log "✓ Services started"

# Step 7: Wait for services to start
log "Step 7: Waiting for services to start..."
sleep 5

# Step 8: Check service status
log "Step 8: Checking service status..."
if systemctl is-active --quiet xingyunpan-server; then
    log "✓ Server service is running"
else
    log "✗ Server service is not running!"
    systemctl status xingyunpan-server | tee -a "$LOG_FILE"
    
    log "Rolling back..."
    systemctl stop xingyunpan-server
    systemctl stop xingyunpan-worker
    
    cp "$APP_DIR/bin/server.backup" "$APP_DIR/bin/server"
    cp "$APP_DIR/bin/worker.backup" "$APP_DIR/bin/worker"
    
    systemctl start xingyunpan-server
    systemctl start xingyunpan-worker
    
    log "✗ Deployment failed - rolled back to previous version"
    exit 1
fi

if systemctl is-active --quiet xingyunpan-worker; then
    log "✓ Worker service is running"
else
    log "✗ Worker service is not running!"
    systemctl status xingyunpan-worker | tee -a "$LOG_FILE"
    
    log "Rolling back..."
    systemctl stop xingyunpan-server
    systemctl stop xingyunpan-worker
    
    cp "$APP_DIR/bin/server.backup" "$APP_DIR/bin/server"
    cp "$APP_DIR/bin/worker.backup" "$APP_DIR/bin/worker"
    
    systemctl start xingyunpan-server
    systemctl start xingyunpan-worker
    
    log "✗ Deployment failed - rolled back to previous version"
    exit 1
fi

# Step 9: Health check
log "Step 9: Performing health check..."
sleep 5  # Wait a bit more for application to fully start

if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    log "✓ Health check passed"
else
    log "✗ Health check failed!"
    
    log "Rolling back..."
    systemctl stop xingyunpan-server
    systemctl stop xingyunpan-worker
    
    cp "$APP_DIR/bin/server.backup" "$APP_DIR/bin/server"
    cp "$APP_DIR/bin/worker.backup" "$APP_DIR/bin/worker"
    
    systemctl start xingyunpan-server
    systemctl start xingyunpan-worker
    
    log "✗ Deployment failed - rolled back to previous version"
    exit 1
fi

# Step 10: Cleanup old backups
log "Step 10: Cleaning up old backups..."
find "$BACKUP_DIR" -name "pre-deploy-*.sql.gz" -mtime +7 -delete
log "✓ Old backups cleaned up (kept last 7 days)"

log "=== Deployment Complete ==="
log "Deployment finished at $(date)"
echo ""
echo "✅ Deployment successful!"
echo ""
echo "Service status:"
systemctl status xingyunpan-server --no-pager | grep "Active:"
systemctl status xingyunpan-worker --no-pager | grep "Active:"
echo ""
echo "Deployment log: $LOG_FILE"
