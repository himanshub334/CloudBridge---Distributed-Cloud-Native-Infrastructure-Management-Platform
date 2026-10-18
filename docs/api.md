# API endpoints

- `GET /health` — API health and queue mode
- `GET /api/jobs` — list jobs currently held by the API process
- `POST /api/jobs` — submit a job (resource type: ec2, s3, iam, rds; action: create, update, delete)
- `GET /api/events` — Server-Sent Events for job updates and periodic snapshots

Example body:
```json
{"resourceType":"ec2","action":"create","config":{"image_id":"ami-sandbox-image","instance_type":"t3.micro"}}
```

The local queue is a demo simulation. AWS SQS publishing is not implemented yet; AWS queue mode returns an explicit not-implemented response.
