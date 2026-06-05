#!/bin/bash

# Deployment Notification Script
# This script sends deployment notifications via various channels

# Configuration
DEPLOYMENT_STATUS="$1"  # success or failure
DEPLOYMENT_TIME="$(date '+%Y-%m-%d %H:%M:%S')"
COMMIT_SHA="${GITHUB_SHA:-unknown}"
COMMIT_MESSAGE="${GITHUB_COMMIT_MESSAGE:-unknown}"
DEPLOYER="${GITHUB_ACTOR:-unknown}"

# Colors for terminal output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to send webhook notification
send_webhook() {
    local webhook_url="$1"
    local status="$2"
    
    if [ -z "$webhook_url" ]; then
        return
    fi
    
    local color
    local emoji
    if [ "$status" == "success" ]; then
        color="good"
        emoji="✅"
    else
        color="danger"
        emoji="❌"
    fi
    
    local payload=$(cat <<EOF
{
    "text": "${emoji} Deployment ${status}",
    "attachments": [
        {
            "color": "${color}",
            "fields": [
                {
                    "title": "Status",
                    "value": "${status}",
                    "short": true
                },
                {
                    "title": "Time",
                    "value": "${DEPLOYMENT_TIME}",
                    "short": true
                },
                {
                    "title": "Deployer",
                    "value": "${DEPLOYER}",
                    "short": true
                },
                {
                    "title": "Commit",
                    "value": "${COMMIT_SHA:0:7}",
                    "short": true
                },
                {
                    "title": "Message",
                    "value": "${COMMIT_MESSAGE}",
                    "short": false
                }
            ]
        }
    ]
}
EOF
)
    
    curl -X POST -H 'Content-type: application/json' \
        --data "$payload" \
        "$webhook_url" 2>/dev/null
}

# Function to send email notification
send_email() {
    local email="$1"
    local status="$2"
    
    if [ -z "$email" ]; then
        return
    fi
    
    local subject
    if [ "$status" == "success" ]; then
        subject="✅ Deployment Successful - Xingyunpan V2"
    else
        subject="❌ Deployment Failed - Xingyunpan V2"
    fi
    
    local body=$(cat <<EOF
Deployment Status: ${status}
Time: ${DEPLOYMENT_TIME}
Deployer: ${DEPLOYER}
Commit: ${COMMIT_SHA}
Message: ${COMMIT_MESSAGE}

Server: $(hostname)
Application: Xingyunpan V2

---
This is an automated notification from the CI/CD pipeline.
EOF
)
    
    # Send email using mail command (requires mailutils)
    if command -v mail &> /dev/null; then
        echo "$body" | mail -s "$subject" "$email"
    fi
}

# Function to log to file
log_deployment() {
    local status="$1"
    local log_file="/data/xingyunpan/logs/deployment-history.log"
    
    mkdir -p "$(dirname "$log_file")"
    
    echo "[${DEPLOYMENT_TIME}] Status: ${status} | Deployer: ${DEPLOYER} | Commit: ${COMMIT_SHA:0:7} | Message: ${COMMIT_MESSAGE}" >> "$log_file"
}

# Function to update deployment status file
update_status_file() {
    local status="$1"
    local status_file="/opt/xingyunpan-v2/deployment-status.json"
    
    cat > "$status_file" <<EOF
{
    "status": "${status}",
    "timestamp": "${DEPLOYMENT_TIME}",
    "deployer": "${DEPLOYER}",
    "commit": "${COMMIT_SHA}",
    "message": "${COMMIT_MESSAGE}"
}
EOF
}

# Main notification logic
main() {
    if [ -z "$DEPLOYMENT_STATUS" ]; then
        echo "Usage: $0 <success|failure>"
        exit 1
    fi
    
    # Print to console
    if [ "$DEPLOYMENT_STATUS" == "success" ]; then
        echo -e "${GREEN}✅ Deployment Successful${NC}"
        echo "Time: $DEPLOYMENT_TIME"
        echo "Deployer: $DEPLOYER"
        echo "Commit: ${COMMIT_SHA:0:7}"
        echo "Message: $COMMIT_MESSAGE"
    else
        echo -e "${RED}❌ Deployment Failed${NC}"
        echo "Time: $DEPLOYMENT_TIME"
        echo "Deployer: $DEPLOYER"
        echo "Commit: ${COMMIT_SHA:0:7}"
        echo "Message: $COMMIT_MESSAGE"
    fi
    
    # Log to file
    log_deployment "$DEPLOYMENT_STATUS"
    
    # Update status file
    update_status_file "$DEPLOYMENT_STATUS"
    
    # Send webhook notification (if configured)
    if [ -n "$WEBHOOK_URL" ]; then
        send_webhook "$WEBHOOK_URL" "$DEPLOYMENT_STATUS"
    fi
    
    # Send email notification (if configured)
    if [ -n "$NOTIFICATION_EMAIL" ]; then
        send_email "$NOTIFICATION_EMAIL" "$DEPLOYMENT_STATUS"
    fi
    
    # Send to monitoring system (if configured)
    # Example: Send metric to Prometheus Pushgateway
    if [ -n "$PUSHGATEWAY_URL" ]; then
        local metric_value
        if [ "$DEPLOYMENT_STATUS" == "success" ]; then
            metric_value=1
        else
            metric_value=0
        fi
        
        cat <<EOF | curl --data-binary @- "$PUSHGATEWAY_URL/metrics/job/deployment"
# TYPE deployment_status gauge
deployment_status{status="$DEPLOYMENT_STATUS"} $metric_value
# TYPE deployment_timestamp gauge
deployment_timestamp $(date +%s)
EOF
    fi
}

# Run main function
main
