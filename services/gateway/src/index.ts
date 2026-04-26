import express, { Request, Response, NextFunction } from "express";
import http from "http";

const app = express();
app.use(express.json());

const COLLECTOR_URL = process.env.COLLECTOR_URL || "http://localhost:8081";
const ANALYTICS_URL = process.env.ANALYTICS_URL || "http://localhost:8082";
const GATEWAY_PORT = parseInt(process.env.GATEWAY_PORT || "8080", 10);

function log(level: string, message: string): void {
  const ts = new Date().toISOString();
  console.log(`[gateway] ${ts} ${level} ${message}`);
}

function proxyRequest(
  targetUrl: string,
  method: string,
  body?: string
): Promise<{ status: number; data: string }> {
  return new Promise((resolve, reject) => {
    const url = new URL(targetUrl);
    const options: http.RequestOptions = {
      hostname: url.hostname,
      port: url.port,
      path: url.pathname + url.search,
      method,
      headers: { "Content-Type": "application/json" },
    };

    const req = http.request(options, (res) => {
      let data = "";
      res.on("data", (chunk) => (data += chunk));
      res.on("end", () =>
        resolve({ status: res.statusCode || 500, data })
      );
    });

    req.on("error", (err) => reject(err));
    req.setTimeout(5000, () => {
      req.destroy();
      reject(new Error("Request timeout"));
    });

    if (body) req.write(body);
    req.end();
  });
}

app.get("/health", (_req: Request, res: Response) => {
  res.json({ status: "ok", service: "gateway" });
});

app.get("/api/metrics", async (_req: Request, res: Response) => {
  log("INFO", "GET /api/metrics -> collector");
  try {
    const result = await proxyRequest(`${COLLECTOR_URL}/metrics`, "GET");
    res.status(result.status).set("Content-Type", "application/json").send(result.data);
  } catch (err) {
    log("ERROR", `Collector unreachable: ${err}`);
    res.status(502).json({ error: "collector service unavailable" });
  }
});

app.post("/api/metrics", async (req: Request, res: Response) => {
  log("INFO", "POST /api/metrics -> collector");
  try {
    const result = await proxyRequest(
      `${COLLECTOR_URL}/metrics`,
      "POST",
      JSON.stringify(req.body)
    );
    res.status(result.status).set("Content-Type", "application/json").send(result.data);
  } catch (err) {
    log("ERROR", `Collector unreachable: ${err}`);
    res.status(502).json({ error: "collector service unavailable" });
  }
});

app.get("/api/summary", async (_req: Request, res: Response) => {
  log("INFO", "GET /api/summary -> analytics");
  try {
    const result = await proxyRequest(`${ANALYTICS_URL}/summary`, "GET");
    res.status(result.status).set("Content-Type", "application/json").send(result.data);
  } catch (err) {
    log("ERROR", `Analytics unreachable: ${err}`);
    res.status(502).json({ error: "analytics service unavailable" });
  }
});

app.post("/api/analyze", async (req: Request, res: Response) => {
  log("INFO", "POST /api/analyze -> analytics");
  try {
    const result = await proxyRequest(
      `${ANALYTICS_URL}/analyze`,
      "POST",
      JSON.stringify(req.body)
    );
    res.status(result.status).set("Content-Type", "application/json").send(result.data);
  } catch (err) {
    log("ERROR", `Analytics unreachable: ${err}`);
    res.status(502).json({ error: "analytics service unavailable" });
  }
});

app.use((_req: Request, res: Response) => {
  res.status(404).json({ error: "not found" });
});

app.use((err: Error, _req: Request, res: Response, _next: NextFunction) => {
  log("ERROR", `Unhandled error: ${err.message}`);
  res.status(500).json({ error: "internal server error" });
});

export { app };

if (require.main === module) {
  app.listen(GATEWAY_PORT, () => {
    log("INFO", `Starting gateway service on :${GATEWAY_PORT}`);
  });
}
