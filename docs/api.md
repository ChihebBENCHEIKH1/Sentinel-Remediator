# API Reference

## Base URL

```
http://localhost:8080
```

## Endpoints

### Health Check

```http
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "service": "sentinel-remediator"
}
```

---

### Submit Remediation Job

```http
POST /api/remediate
Content-Type: application/json
```

**Request Body:**
```json
{
  "scan_result": {
    "scan_id": "scan-001",
    "image_name": "myapp",
    "image_tag": "latest",
    "repo_url": "https://github.com/owner/repo",
    "branch": "main",
    "vulnerabilities": [
      {
        "id": "VULN-001",
        "severity": "HIGH",
        "type": "RUN_AS_ROOT",
        "title": "Container runs as root",
        "description": "...",
        "file_path": "Dockerfile",
        "line_number": 1,
        "suggestion": "Add non-root user"
      }
    ]
  }
}
```

**Response:**
```json
{
  "job_id": "uuid",
  "status": "PENDING",
  "message": "Remediation job started"
}
```

---

### List Jobs

```http
GET /api/jobs
```

**Response:**
```json
[
  {
    "id": "uuid",
    "status": "REASONING",
    "progress": 0.33,
    "fixed_count": 1,
    "failed_count": 0,
    "total_count": 3,
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

---

### Get Job Details

```http
GET /api/jobs/:id
```

**Response:**
```json
{
  "id": "uuid",
  "status": "SUCCESS",
  "progress": 1.0,
  "thought_trace": [
    {
      "timestamp": "...",
      "type": "thought",
      "content": "Analyzing vulnerability...",
      "iteration": 0
    }
  ],
  "fix_attempts": [...],
  "pr_url": "https://github.com/..."
}
```

---

### Stream Job Events (SSE)

```http
GET /api/jobs/:id/stream
Accept: text/event-stream
```

**Events:**
```
event: init
data: {"job_id":"...","status":"PENDING"}

event: thought
data: {"type":"thought","content":"Analyzing...","iteration":0}

event: complete
data: {"status":"SUCCESS","pr_url":"..."}
```

---

### Cancel Job

```http
DELETE /api/jobs/:id
```

**Response:**
```json
{
  "message": "Job cancelled",
  "job_id": "uuid"
}
```

## Vulnerability Types

| Type | Description |
|------|-------------|
| `RUN_AS_ROOT` | Container runs as root user |
| `NO_HEALTHCHECK` | Missing HEALTHCHECK instruction |
| `OUTDATED_BASE_IMAGE` | Base image has known CVEs |
| `HARDCODED_SECRET` | Secrets in Dockerfile |
| `PRIVILEGED_CONTAINER` | Privileged mode enabled |

## Job Statuses

| Status | Description |
|--------|-------------|
| `PENDING` | Job created, waiting to start |
| `REASONING` | Agent is thinking about fix |
| `APPLYING_FIX` | Modifying files |
| `BUILDING` | Running Docker build |
| `PUSHING` | Pushing to Git |
| `SUCCESS` | All fixes applied |
| `FAILED` | Could not fix vulnerabilities |
