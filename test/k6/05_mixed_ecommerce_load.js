import http from 'k6/http';
import { check, sleep } from 'k6';
import { BASE_URL, DEFAULT_HEADERS, SLA_THRESHOLDS } from './lib/config.js';
import { authHeaders, ensureTestUser, randomInt, uuidv4 } from './lib/helpers.js';

const MIXED_EMAIL = __ENV.K6_MIXED_EMAIL || 'k6_mixed@catalyst.dev';
const MIXED_PASSWORD = __ENV.K6_MIXED_PASSWORD || 'Password123!';

export const options = {
  scenarios: {
    // 75% weight: Catalog browsing
    browse_catalog: {
      executor: 'ramping-arrival-rate',
      startRate: 15,
      timeUnit: '1s',
      preAllocatedVUs: 30,
      maxVUs: 100,
      stages: [
        { duration: '20s', target: 45 },
        { duration: '60s', target: 60 },
        { duration: '20s', target: 15 },
      ],
      exec: 'browseCatalog',
    },
    // 15% weight: Checkout & order history
    checkout_and_history: {
      executor: 'ramping-arrival-rate',
      startRate: 3,
      timeUnit: '1s',
      preAllocatedVUs: 15,
      maxVUs: 50,
      stages: [
        { duration: '20s', target: 9 },
        { duration: '60s', target: 12 },
        { duration: '20s', target: 3 },
      ],
      exec: 'checkoutAndHistory',
    },
    // 10% weight: User login (CPU-bound bcrypt)
    user_login: {
      executor: 'ramping-arrival-rate',
      startRate: 2,
      timeUnit: '1s',
      preAllocatedVUs: 10,
      maxVUs: 30,
      stages: [
        { duration: '20s', target: 6 },
        { duration: '60s', target: 8 },
        { duration: '20s', target: 2 },
      ],
      exec: 'userLogin',
    },
  },
  thresholds: {
    http_req_duration: SLA_THRESHOLDS.mixedLoad.http_req_duration,
    http_req_failed: SLA_THRESHOLDS.mixedLoad.http_req_failed,
  },
};

export function setup() {
  const token = ensureTestUser(BASE_URL, MIXED_EMAIL, MIXED_PASSWORD, 'K6 Mixed Load User');
  return {
    token,
    email: MIXED_EMAIL,
    password: MIXED_PASSWORD,
  };
}

export function browseCatalog() {
  const page = randomInt(1, 4);
  const listRes = http.get(`${BASE_URL}/api/v1/products?page=${page}&pageSize=10`, {
    tags: { name: 'GET /api/v1/products' },
  });
  check(listRes, { 'browse list 200': (r) => r.status === 200 });

  const productId = randomInt(1, 20);
  const detailRes = http.get(`${BASE_URL}/api/v1/products/${productId}`, {
    tags: { name: 'GET /api/v1/products/:id' },
  });
  check(detailRes, { 'browse detail 200': (r) => r.status === 200 });

  sleep(0.05);
}

export function checkoutAndHistory(data) {
  const productId = randomInt(41, 45);
  const payload = JSON.stringify({
    products: [{ id: productId, qty: 1 }],
  });

  const params = authHeaders(data.token, {
    'Idempotency-Key': uuidv4(),
  });
  params.tags = { name: 'POST /api/v1/transactions' };

  const txRes = http.post(`${BASE_URL}/api/v1/transactions`, payload, params);
  check(txRes, { 'checkout status 201': (r) => r.status === 201 });

  const historyParams = authHeaders(data.token);
  historyParams.tags = { name: 'GET /api/v1/transactions' };
  const histRes = http.get(`${BASE_URL}/api/v1/transactions`, historyParams);
  check(histRes, { 'history status 200': (r) => r.status === 200 });
}

export function userLogin(data) {
  const payload = JSON.stringify({
    email: data.email,
    password: data.password,
  });

  const res = http.post(`${BASE_URL}/api/v1/users/login`, payload, {
    headers: DEFAULT_HEADERS,
    tags: { name: 'POST /api/v1/users/login' },
  });

  check(res, { 'login status 200': (r) => r.status === 200 });
}

export function teardown() {
  const res = http.get(`${BASE_URL}/metrics`, {
    tags: { name: 'GET /metrics (teardown)' },
  });

  if (res.status !== 200) {
    console.warn(`Unable to fetch /metrics during teardown: status=${res.status}`);
    return;
  }

  const extractMetric = (regex) => {
    const match = res.body.match(regex);
    return match ? match[1] : 'N/A';
  };

  const goroutines = extractMetric(/^go_goroutines\s+([0-9.eE+-]+)/m);
  const heapAlloc = extractMetric(/^go_memstats_alloc_bytes\s+([0-9.eE+-]+)/m);
  const gcCount = extractMetric(/^go_gc_duration_seconds_count\s+([0-9.eE+-]+)/m);

  console.log(
    `[Teardown Metrics Snapshot] go_goroutines=${goroutines}, go_memstats_alloc_bytes=${heapAlloc}, go_gc_duration_seconds_count=${gcCount}`
  );
}

