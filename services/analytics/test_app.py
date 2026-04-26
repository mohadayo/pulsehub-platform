import json

import pytest

from app import app, compute_summary


@pytest.fixture
def client():
    app.config["TESTING"] = True
    with app.test_client() as c:
        yield c


def test_health(client):
    resp = client.get("/health")
    assert resp.status_code == 200
    data = resp.get_json()
    assert data["status"] == "ok"
    assert data["service"] == "analytics"


def test_analyze_valid(client):
    payload = {
        "metrics": [
            {"service": "web", "name": "cpu", "value": 40},
            {"service": "web", "name": "cpu", "value": 60},
            {"service": "db", "name": "memory", "value": 2048},
        ]
    }
    resp = client.post("/analyze", json=payload)
    assert resp.status_code == 200
    data = resp.get_json()
    assert data["total_metrics"] == 3
    assert "web/cpu" in data["services"]
    assert data["services"]["web/cpu"]["mean"] == 50.0
    assert data["services"]["web/cpu"]["count"] == 2
    assert data["services"]["db/memory"]["count"] == 1


def test_analyze_missing_metrics(client):
    resp = client.post("/analyze", json={"data": []})
    assert resp.status_code == 400


def test_analyze_empty_body(client):
    resp = client.post("/analyze", data="", content_type="application/json")
    assert resp.status_code == 400


def test_compute_summary_empty():
    result = compute_summary([])
    assert result["total_metrics"] == 0
    assert result["services"] == {}


def test_compute_summary_single():
    metrics = [{"service": "api", "name": "latency", "value": 120}]
    result = compute_summary(metrics)
    assert result["total_metrics"] == 1
    assert result["services"]["api/latency"]["mean"] == 120
    assert "stdev" not in result["services"]["api/latency"]


def test_compute_summary_multiple():
    metrics = [
        {"service": "api", "name": "latency", "value": 100},
        {"service": "api", "name": "latency", "value": 200},
        {"service": "api", "name": "latency", "value": 150},
    ]
    result = compute_summary(metrics)
    assert result["total_metrics"] == 3
    s = result["services"]["api/latency"]
    assert s["mean"] == 150.0
    assert s["min"] == 100
    assert s["max"] == 200
    assert "stdev" in s
