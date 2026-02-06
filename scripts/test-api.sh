#!/bin/bash
# scripts/test-api.sh - Quick API test script

set -e

BASE_URL="${API_URL:-http://localhost:8080}"

echo "🧪 Testing Sentinel-Remediator API at $BASE_URL"
echo ""

# Health check
echo "1. Health Check..."
HEALTH=$(curl -s "$BASE_URL/health")
echo "   Response: $HEALTH"
if [[ "$HEALTH" == *"ok"* ]]; then
    echo "   ✅ Health check passed"
else
    echo "   ❌ Health check failed"
    exit 1
fi
echo ""

# List jobs
echo "2. List Jobs..."
JOBS=$(curl -s "$BASE_URL/api/jobs")
echo "   Response: $JOBS"
echo "   ✅ Jobs endpoint working"
echo ""

# Create test scan
echo "3. Submit Test Scan..."
RESPONSE=$(curl -s -X POST "$BASE_URL/api/remediate" \
    -H "Content-Type: application/json" \
    -d '{
        "scan_result": {
            "scan_id": "test-001",
            "image_name": "test",
            "image_tag": "latest",
            "repo_url": "https://github.com/example/test",
            "branch": "main",
            "vulnerabilities": [
                {
                    "id": "TEST-001",
                    "severity": "HIGH",
                    "type": "RUN_AS_ROOT",
                    "title": "Test vulnerability",
                    "description": "This is a test",
                    "file_path": "Dockerfile"
                }
            ]
        }
    }')
echo "   Response: $RESPONSE"

if [[ "$RESPONSE" == *"job_id"* ]]; then
    JOB_ID=$(echo "$RESPONSE" | grep -o '"job_id":"[^"]*"' | cut -d'"' -f4)
    echo "   ✅ Job created: $JOB_ID"
    
    # Get job status
    echo ""
    echo "4. Get Job Status..."
    sleep 1
    STATUS=$(curl -s "$BASE_URL/api/jobs/$JOB_ID")
    echo "   Status: $(echo "$STATUS" | grep -o '"status":"[^"]*"')"
    echo "   ✅ Job status endpoint working"
else
    echo "   ❌ Job creation failed"
fi

echo ""
echo "🎉 All API tests completed!"
