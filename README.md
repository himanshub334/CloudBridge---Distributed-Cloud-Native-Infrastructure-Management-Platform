# CloudBridge — Distributed Cloud-Native Infrastructure Management Platform

A portfolio/demo project for infrastructure request intake, asynchronous job tracking, a cloud worker service, and an operations dashboard.

## Stack
Go/Gin API · Python/FastAPI + boto3 worker · AWS SQS/S3/EC2/IAM · Terraform · PostgreSQL · Redis · React/TypeScript · Docker · GitHub Actions

## Run locally
Requires Docker Desktop and Docker Compose.

```bash
cp .env.example .env
docker compose up --build
```

- Dashboard: http://localhost:5173
- Go API health: http://localhost:8080/health
- Python worker health: http://localhost:8001/health

Local queue and dry-run mode are defaults. The local simulation does not create cloud resources.

## AWS mode
This repository includes Terraform scaffolding for an SQS queue and dead-letter queue, a versioned S3 state bucket, DynamoDB lock table, and a CloudWatch queue-depth alarm. Review and configure all variables before running Terraform. Never commit AWS keys.

The Go SQS publisher is deliberately a placeholder, so AWS queue mode will not work until the AWS SDK integration is implemented. Python worker operations are opt-in with `DRY_RUN=false`; use only an isolated sandbox account and least-privilege IAM permissions.

## Project modules
- `api/`: Go/Gin REST API, local job queue, PostgreSQL persistence, Redis status cache, SSE snapshots
- `worker/`: FastAPI worker with dry-run default and boto3 handlers for selected EC2/S3/IAM actions
- `frontend/`: React dashboard with job search/filter and provisioning request form
- `infra/terraform/`: AWS infrastructure starter
- `docs/`: API notes, AI-assisted development log, product metrics and implementation caveats

## Validation
ZIP integrity was checked. Full Docker build, frontend build, Go tests, Python tests, Terraform plan, and AWS deployment have not been run in this environment. CI configuration is included as a starting point.

## Security scope
Demo only, not a production cloud control plane. Authentication, tenant isolation, approvals, a hardened policy engine, full audit controls, Slack digest delivery, and real CloudWatch custom metrics are not complete. Do not claim productivity improvements or performance numbers without measured evidence.
