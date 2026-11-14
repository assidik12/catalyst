import http from 'k6/http';
import { check } from 'k6';
import { Counter, Trend } from 'k6/metrics';
import { BASE_URL, SLA_THRESHOLDS } from './lib/config.js';
import { authHeaders, ensureTestUser, randomInt, uuidv4 } from './lib/helpers.js';

const txCreated201 = new Counter('tx_created_201');
const txOutOfStock400 = new Counter('tx_out_of_stock_400');
const txConflict409 = new Counter('tx_conflict_409');
const txLatency = new Trend('tx_create_duration', true);

const BUYER_EMAIL = __ENV.K6_BUYER_EMAIL || 'k6_buyer@catalyst.dev';
const BUYER_PASSWORD = __ENV.K6_BUYER_PASSWORD || 'Password123!';

export const options = {
  scenarios: {
    tx_arrival_rate: {
      executor: 'ramping-arrival-rate',
      startRate: 10,
      timeUnit: '1s',
      preAllocatedVUs: 30,
      maxVUs: 100,
      stages: [
        { duration: '15s', target: 10 },
        { duration: '25s', target: 30 },
        { duration: '20s', target: 50 },
      ],
    },
  },
  thresholds: {
    http_req_duration: SLA_THRESHOLDS.transaction.http_req_duration,
    http_req_failed: SLA_THRESHOLDS.transaction.http_req_failed,
  },
};

export function setup() {
  const token = ensureTestUser(BASE_URL, BUYER_EMAIL, BUYER_PASSWORD, 'K6 Buyer');
  return { token };
}

export default function (data) {
  // Use seeded high-stock Food & Beverages / Automotive products (IDs 41..45)
  const productId = randomInt(41, 45);
  const payload = JSON.stringify({
    products: [{ id: productId, qty: 1 }],
  });

  const params = authHeaders(data.token, {
    'Idempotency-Key': uuidv4(),
  });
  params.tags = { name: 'POST /api/v1/transactions' };

  const res = http.post(`${BASE_URL}/api/v1/transactions`, payload, params);
  txLatency.add(res.timings.duration);

  if (res.status === 201) {
    txCreated201.add(1);
  } else if (res.status === 400) {
    txOutOfStock400.add(1);
  } else if (res.status === 409) {
    txConflict409.add(1);
  }

  check(res, {
    'transaction created (201)': (r) => r.status === 201,
    'server returns transaction id': (r) => r.status !== 201 || Boolean(r.json('data.id')),
  });
}

