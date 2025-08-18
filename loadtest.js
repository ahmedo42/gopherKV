import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  scenarios: {
    readHeavy: {
      executor: 'constant-vus',
      vus: 1000,
      duration: '5m',
      exec: 'readHeavyScenario',
    },
    writeHeavy: {
      executor: 'constant-vus',
      vus: 1000,
      duration: '5m',
      exec: 'writeHeavyScenario',
    },
    mixed: {
      executor: 'constant-vus',
      vus: 1000,
      duration: '5m',
      exec: 'mixedScenario',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],   // <1% errors
    http_req_duration: ['p(95)<100'], // 95% <100ms
  },
  insecureSkipTLSVerify: true,
  noConnectionReuse: false,
};

// --- Read-heavy workload ---
export function readHeavyScenario() {
  let key = `key${Math.floor(Math.random() * 1000)}`;
  let res = http.get(`http://localhost:8080/get/${key}`);
  check(res, { 'status is 200 or 404': (r) => r.status === 200 || r.status === 404 });
  sleep(0.1);
}

// --- Write-heavy workload ---
export function writeHeavyScenario() {
  let key = `key${Math.floor(Math.random() * 1000)}`;
  let payload = JSON.stringify({ value: Math.random().toString(36).substring(7) });
  let res = http.put(`http://localhost:8080/put/${key}`, payload, {
    headers: { 'Content-Type': 'application/json' },
  });
  check(res, { 'status is 200': (r) => r.status === 200 });
  sleep(0.1);
}

// --- Mixed workload (80% reads, 20% writes/deletes) ---
export function mixedScenario() {
  let rand = Math.random();
  let key = `key${Math.floor(Math.random() * 1000)}`;

  if (rand < 0.7) {
    http.get(`http://localhost:8080/get/${key}`);
  } else if (rand < 0.9) {
    let payload = JSON.stringify({ value: Math.random().toString(36).substring(7) });
    http.put(`http://localhost:8080/put/${key}`, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
  } else {
    http.del(`http://localhost:8080/delete/${key}`);
  }
  sleep(0.1);
}
