from fastapi.testclient import TestClient
import app
client = TestClient(app.app)

def test_health():
    response = client.get("/health")
    assert response.status_code == 200
    assert "dryRun" in response.json()

def test_unknown_resource_rejected(monkeypatch):
    monkeypatch.setattr(app, "DRY_RUN", True)
    r = client.post("/process", json={"job_id":"t1","resource_type":"other","action":"create","config":{}})
    assert r.status_code == 400

def test_dry_run_no_aws_change(monkeypatch):
    monkeypatch.setattr(app, "DRY_RUN", True)
    r = client.post("/process", json={"job_id":"t2","resource_type":"ec2","action":"create","config":{}})
    assert r.status_code == 200 and r.json()["dryRun"] is True
