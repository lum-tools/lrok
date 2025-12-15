import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    { duration: "1m", target: 10 }, // Ramp up to 10 users
    { duration: "3m", target: 10 }, // Stay at 10 users
    { duration: "1m", target: 20 }, // Ramp up to 20 users
    { duration: "3m", target: 20 }, // Stay at 20 users
    { duration: "1m", target: 0 }, // Ramp down
  ],
  thresholds: {
    http_req_duration: ["p(95)<1000"], // 95% of requests under 1s
    http_req_failed: ["rate<0.05"], // Less than 5% errors
  },
};

const API_URL = __ENV.LROK_API_URL || "http://localhost:4243";
const API_KEY = __ENV.LUM_API_KEY || "lum_test_key";

export default function () {
  const headers = {
    "Content-Type": "application/json",
    "Authorization": `Bearer ${API_KEY}`,
  };

  // Create tunnel
  const createPayload = JSON.stringify({
    type: "http",
    localPort: 8000 + Math.floor(Math.random() * 100),
  });

  const createResponse = http.post(`${API_URL}/api/v1/tunnels`, createPayload, { headers });

  check(createResponse, {
    "tunnel created": (r) => r.status === 201,
    "has tunnel ID": (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.id !== undefined;
      } catch {
        return false;
      }
    },
  });

  if (createResponse.status === 201) {
    const tunnel = JSON.parse(createResponse.body);

    // Get tunnel details
    const getResponse = http.get(`${API_URL}/api/v1/tunnels/${tunnel.id}`, { headers });
    check(getResponse, {
      "tunnel retrieved": (r) => r.status === 200,
    });

    sleep(1);

    // Get stats
    const statsResponse = http.get(`${API_URL}/api/v1/tunnels/${tunnel.id}/stats`, { headers });
    check(statsResponse, {
      "stats retrieved": (r) => r.status === 200,
    });

    // Delete tunnel
    const deleteResponse = http.del(`${API_URL}/api/v1/tunnels/${tunnel.id}`, null, { headers });
    check(deleteResponse, {
      "tunnel deleted": (r) => r.status === 204,
    });
  }

  sleep(1);
}

export function handleSummary(data) {
  return {
    "stdout": textSummary(data, { indent: " ", enableColors: true }),
    "test/results/load-test-summary.json": JSON.stringify(data),
  };
}

function textSummary(data, { indent, enableColors }) {
  const colors = {
    reset: enableColors ? "\x1b[0m" : "",
    green: enableColors ? "\x1b[32m" : "",
    red: enableColors ? "\x1b[31m" : "",
    yellow: enableColors ? "\x1b[33m" : "",
  };

  let summary = `\n${indent}📊 Load Test Summary\n`;
  summary += `${indent}${"=".repeat(50)}\n`;

  const metrics = data.metrics;

  // Requests
  if (metrics.http_reqs) {
    summary += `${indent}Total Requests: ${metrics.http_reqs.values.count}\n`;
  }

  // Duration
  if (metrics.http_req_duration) {
    const p95 = metrics.http_req_duration.values["p(95)"];
    const avg = metrics.http_req_duration.values.avg;
    const color = p95 < 1000 ? colors.green : colors.red;
    summary += `${indent}Request Duration (avg): ${avg.toFixed(2)}ms\n`;
    summary += `${indent}Request Duration (p95): ${color}${p95.toFixed(2)}ms${colors.reset}\n`;
  }

  // Error rate
  if (metrics.http_req_failed) {
    const rate = metrics.http_req_failed.values.rate * 100;
    const color = rate < 5 ? colors.green : colors.red;
    summary += `${indent}Error Rate: ${color}${rate.toFixed(2)}%${colors.reset}\n`;
  }

  summary += `${indent}${"=".repeat(50)}\n`;

  return summary;
}
