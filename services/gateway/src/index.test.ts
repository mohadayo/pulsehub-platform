import request from "supertest";
import { app } from "./index";

describe("Gateway", () => {
  describe("GET /health", () => {
    it("returns ok status", async () => {
      const res = await request(app).get("/health");
      expect(res.status).toBe(200);
      expect(res.body.status).toBe("ok");
      expect(res.body.service).toBe("gateway");
    });
  });

  describe("GET /api/metrics", () => {
    it("returns 502 when collector is down", async () => {
      const res = await request(app).get("/api/metrics");
      expect(res.status).toBe(502);
      expect(res.body.error).toBe("collector service unavailable");
    });
  });

  describe("POST /api/metrics", () => {
    it("returns 502 when collector is down", async () => {
      const res = await request(app)
        .post("/api/metrics")
        .send({ service: "test", name: "cpu", value: 50 });
      expect(res.status).toBe(502);
    });
  });

  describe("GET /api/summary", () => {
    it("returns 502 when analytics is down", async () => {
      const res = await request(app).get("/api/summary");
      expect(res.status).toBe(502);
      expect(res.body.error).toBe("analytics service unavailable");
    });
  });

  describe("POST /api/analyze", () => {
    it("returns 502 when analytics is down", async () => {
      const res = await request(app)
        .post("/api/analyze")
        .send({ metrics: [] });
      expect(res.status).toBe(502);
    });
  });

  describe("Unknown route", () => {
    it("returns 404", async () => {
      const res = await request(app).get("/unknown");
      expect(res.status).toBe(404);
      expect(res.body.error).toBe("not found");
    });
  });
});
