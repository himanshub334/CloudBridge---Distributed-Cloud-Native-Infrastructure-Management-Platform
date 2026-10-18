import os
from typing import Any
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
import boto3

app = FastAPI(title="CloudBridge Worker", version="1.0.0")
DRY_RUN = os.getenv("DRY_RUN", "true").lower() != "false"
REGION = os.getenv("AWS_REGION", "us-east-1")

class ProvisionRequest(BaseModel):
    job_id: str
    resource_type: str
    action: str
    config: dict[str, Any] = Field(default_factory=dict)

@app.get("/health")
def health():
    return {"status": "UP", "dryRun": DRY_RUN}

@app.post("/process")
def process(req: ProvisionRequest):
    kind, action = req.resource_type.lower(), req.action.lower()
    if kind not in {"ec2", "s3", "iam", "rds"} or action not in {"create", "update", "delete"}:
        raise HTTPException(400, "unsupported resource action")
    if DRY_RUN:
        return {"jobId": req.job_id, "status": "SUCCEEDED",
                "message": f"Dry-run accepted {action} {kind}; no AWS resources changed", "dryRun": True}
    try:
        session = boto3.Session(region_name=REGION)
        if kind == "s3":
            bucket = req.config.get("bucket")
            if not bucket: raise HTTPException(400, "config.bucket is required")
            client = session.client("s3")
            if action == "create": client.create_bucket(Bucket=bucket)
            elif action == "delete": client.delete_bucket(Bucket=bucket)
            else: raise HTTPException(400, "S3 update is not implemented")
            result = {"bucket": bucket}
        elif kind == "ec2":
            client = session.client("ec2")
            if action == "create":
                image = req.config.get("image_id")
                if not image: raise HTTPException(400, "config.image_id is required")
                result = client.run_instances(ImageId=image, InstanceType=req.config.get("instance_type","t3.micro"), MinCount=1, MaxCount=1)
            elif action == "delete":
                iid = req.config.get("instance_id")
                if not iid: raise HTTPException(400, "config.instance_id is required")
                client.terminate_instances(InstanceIds=[iid]); result = {"instanceId": iid}
            else: raise HTTPException(400, "EC2 update is not implemented")
        elif kind == "iam":
            client = session.client("iam")
            if action == "create":
                name, doc = req.config.get("policy_name"), req.config.get("policy_document")
                if not name or not doc: raise HTTPException(400, "policy_name and policy_document are required")
                result = client.create_policy(PolicyName=name, PolicyDocument=doc)
            elif action == "delete":
                arn = req.config.get("policy_arn")
                if not arn: raise HTTPException(400, "policy_arn is required")
                client.delete_policy(PolicyArn=arn); result = {"policyArn": arn}
            else: raise HTTPException(400, "IAM update is not implemented")
        else:
            raise HTTPException(400, "RDS provisioning is not implemented; add subnet/security-group configuration")
        return {"jobId": req.job_id, "status": "SUCCEEDED", "result": str(result)[:800], "dryRun": False}
    except HTTPException:
        raise
    except Exception as exc:
        raise HTTPException(502, f"AWS operation failed: {type(exc).__name__}") from exc
