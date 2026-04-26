import logging
import os
from statistics import mean, stdev

import httpx
from flask import Flask, jsonify, request

app = Flask(__name__)

logging.basicConfig(
    level=logging.INFO,
    format="[analytics] %(asctime)s %(levelname)s %(message)s",
)
logger = logging.getLogger(__name__)

COLLECTOR_URL = os.getenv("COLLECTOR_URL", "http://localhost:8081")


def fetch_metrics():
    try:
        resp = httpx.get(f"{COLLECTOR_URL}/metrics", timeout=5.0)
        resp.raise_for_status()
        return resp.json()
    except httpx.HTTPError as e:
        logger.error("Failed to fetch metrics from collector: %s", e)
        return None


def compute_summary(metrics):
    if not metrics:
        return {"services": {}, "total_metrics": 0}

    grouped = {}
    for m in metrics:
        svc = m.get("service", "unknown")
        name = m.get("name", "unknown")
        value = m.get("value", 0)
        key = f"{svc}/{name}"
        grouped.setdefault(key, []).append(value)

    summary = {}
    for key, values in grouped.items():
        entry = {"count": len(values), "mean": round(mean(values), 4)}
        if len(values) >= 2:
            entry["stdev"] = round(stdev(values), 4)
        entry["min"] = min(values)
        entry["max"] = max(values)
        summary[key] = entry

    return {"services": summary, "total_metrics": len(metrics)}


@app.route("/health")
def health():
    return jsonify({"status": "ok", "service": "analytics"})


@app.route("/summary")
def summary():
    logger.info("GET /summary")
    metrics = fetch_metrics()
    if metrics is None:
        logger.warning("Collector unreachable, returning empty summary")
        return jsonify({"error": "collector unreachable"}), 502
    result = compute_summary(metrics)
    logger.info("Summary computed: %d total metrics", result["total_metrics"])
    return jsonify(result)


@app.route("/analyze", methods=["POST"])
def analyze():
    logger.info("POST /analyze")
    data = request.get_json(silent=True)
    if not data or "metrics" not in data:
        logger.warning("Invalid payload for /analyze")
        return jsonify({"error": "metrics array is required"}), 400
    result = compute_summary(data["metrics"])
    return jsonify(result)


def create_app():
    return app


if __name__ == "__main__":
    port = int(os.getenv("ANALYTICS_PORT", "8082"))
    logger.info("Starting analytics service on :%d", port)
    app.run(host="0.0.0.0", port=port)
