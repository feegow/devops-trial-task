import { createServer } from "http";
import { readFile } from "fs/promises";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const publicDir = path.join(__dirname, "public");

const MIME_TYPES = {
  ".html": "text/html; charset=utf-8",
  ".js": "text/javascript; charset=utf-8",
  ".css": "text/css; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".svg": "image/svg+xml; charset=utf-8",
  ".png": "image/png",
  ".ico": "image/x-icon",
};

const PORT = Number(process.env.PORT || 3000);

// Simple Prometheus metrics store
const metrics = {
  requests: new Map(), // route -> status -> count
  histograms: new Map(), // route -> {buckets, sum, count}
  buckets: [0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10],
};

function observeRequest(route, status, durationSeconds) {
  // Update counter
  if (!metrics.requests.has(route)) {
    metrics.requests.set(route, new Map());
  }
  const routeMap = metrics.requests.get(route);
  routeMap.set(status, (routeMap.get(status) || 0) + 1);

  // Update histogram
  if (!metrics.histograms.has(route)) {
    metrics.histograms.set(route, {
      buckets: new Array(metrics.buckets.length).fill(0),
      infBucket: 0,
      sum: 0,
      count: 0,
    });
  }
  const hist = metrics.histograms.get(route);
  hist.sum += durationSeconds;
  hist.count++;
  
  let foundBucket = false;
  for (let i = 0; i < metrics.buckets.length; i++) {
    if (durationSeconds <= metrics.buckets[i]) {
      hist.buckets[i]++;
      foundBucket = true;
      break;
    }
  }
  if (!foundBucket) {
    hist.infBucket++;
  }
}

function generateMetrics() {
  let output = '';
  
  // Request counter
  output += '# HELP http_requests_total Total HTTP requests\n';
  output += '# TYPE http_requests_total counter\n';
  for (const [route, statuses] of metrics.requests.entries()) {
    for (const [status, count] of statuses.entries()) {
      output += `http_requests_total{route="${route}",status="${status}"} ${count}\n`;
    }
  }
  
  // Request duration histogram
  output += '# HELP http_request_duration_seconds Request latency in seconds\n';
  output += '# TYPE http_request_duration_seconds histogram\n';
  for (const [route, hist] of metrics.histograms.entries()) {
    let cumulative = 0;
    for (let i = 0; i < metrics.buckets.length; i++) {
      cumulative += hist.buckets[i];
      output += `http_request_duration_seconds_bucket{route="${route}",le="${metrics.buckets[i]}"} ${cumulative}\n`;
    }
    cumulative += hist.infBucket;
    output += `http_request_duration_seconds_bucket{route="${route}",le="+Inf"} ${cumulative}\n`;
    output += `http_request_duration_seconds_sum{route="${route}"} ${hist.sum.toFixed(6)}\n`;
    output += `http_request_duration_seconds_count{route="${route}"} ${hist.count}\n`;
  }
  
  return output;
}

async function serveStatic(urlPath) {
  let normalized = decodeURI(urlPath.split("?")[0]);
  if (normalized === "/") {
    normalized = "/index.html";
  }

  // Prevent path traversal
  normalized = normalized.replace(/\.\./g, "");

  const filePath = path.join(publicDir, normalized);
  const ext = path.extname(filePath).toLowerCase();
  const mime = MIME_TYPES[ext] || "application/octet-stream";
  const buffer = await readFile(filePath);
  return { buffer, mime };
}

const server = createServer(async (req, res) => {
  const startTime = process.hrtime.bigint();
  let status = 200;
  let route = "/";
  
  try {
    if (req.url === "/healthz") {
      route = "/healthz";
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ status: "ok" }));
      return;
    }

    if (req.url === "/metrics") {
      res.writeHead(200, { "Content-Type": "text/plain; version=0.0.4" });
      res.end(generateMetrics());
      return;
    }

    const { buffer, mime } = await serveStatic(req.url);
    res.writeHead(200, {
      "Content-Type": mime,
      "Cache-Control": mime.startsWith("text/") ? "no-cache" : "public, max-age=604800",
    });
    res.end(buffer);
  } catch (error) {
    if (error.code === "ENOENT") {
      status = 404;
      res.writeHead(404, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ error: "Not found" }));
    } else {
      status = 500;
      res.writeHead(500, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ error: "Internal server error" }));
    }
  } finally {
    const endTime = process.hrtime.bigint();
    const durationNs = Number(endTime - startTime);
    const durationSeconds = durationNs / 1_000_000_000;
    
    // Don't track metrics endpoint itself
    if (req.url !== "/metrics") {
      observeRequest(route, status, durationSeconds);
    }
  }
});

server.listen(PORT, () => {
  console.log(`available-schedules-web listening on http://0.0.0.0:${PORT}`);
});

