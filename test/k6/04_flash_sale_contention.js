import http from 'k6/http';
import { check } from 'k6';
import { Counter } from 'k6/metrics';
import { BASE_URL } from './lib/config.js';
import { authHeaders, ensureTestUser, uuidv4 } from './lib/helpers.js';

const flashSaleCreated = new Counter('flash_sale_created_201');
const flashSaleOutOfStock = new Counter('flash_sale_out_of_stock_400');
const flashSaleError = new Counter('flash_sale_unexpected_error');

const idempCreated = new Counter('idempotency_created_201');
const idempConflict = new Counter('idempotency_conflict_409');

const BUYER_EMAIL = __ENV.K6_FLASH_EMAIL || 'k6_flash@catalyst.dev';
const BUYER_PASSWORD = __ENV.K6_FLASH_PASSWORD || 'Password123!';
const TARGET_PRODUCT_ID = parseInt(__ENV.TARGET_PRODUCT_ID || '1', 10);

export const options = {
  scenarios: {
    // 4A: 150 concurrent requests racing for limited stock on 1 product (e.g. stock=100)
    flash_sale_race: {
      executor: 'shared-iterations',
      vus: 50,
      iterations: 150,
      maxDuration: '45s',
      exec: 'flashSaleRace',
    },
    // 4B: 20 VUs sending the exact same Idempotency-Key simultaneously
    idempotency_race: {
      executor: 'per-vu-iterations',
      startTime: '15s',
      vus: 20,
      iterations: 1,
      maxDuration: '30s',
      exec: 'idempotencyRace',
    },
  },
  thresholds: {
    flash_sale_unexpected_error: ['count==0'],
    idempotency_created_201: ['count<=1'],
    idempotency_conflict_409: ['count>=19'],
  },
};

export function setup() {
  const token = ensureTestUser(BASE_URL, BUYER_EMAIL, BUYER_PASSWORD, 'K6 Flash Sale Buyer');
  return {
    token,
    sharedIdempotencyKey: `k6-idemp-race-${Date.now()}`,
  };
}

export function flashSaleRace(data) {
  const payload = JSON.stringify({
    products: [{ id: TARGET_PRODUCT_ID, qty: 1 }],
  });

  const params = authHeaders(data.token, {
    'Idempotency-Key': uuidv4(),
  });
  params.tags = { name: 'POST /api/v1/transactions (flash_sale)' };

  const res = http.post(`${BASE_URL}/api/v1/transactions`, payload, params);

  if (res.status === 201) {
    flashSaleCreated.add(1);
  } else if (res.status === 400) {
    flashSaleOutOfStock.add(1);
  } else {
    flashSaleError.add(1);
  }

  check(res, {
    'flash sale status is 201 or 400': (r) => r.status === 201 || r.status === 400,
  });
}

export function idempotencyRace(data) {
  const payload = JSON.stringify({
    products: [{ id: 45, qty: 1 }],
  });

  const params = authHeaders(data.token, {
    'Idempotency-Key': data.sharedIdempotencyKey,
  });
  params.tags = { name: 'POST /api/v1/transactions (idempotency_race)' };

  const res = http.post(`${BASE_URL}/api/v1/transactions`, payload, params);

  if (res.status === 201) {
    idempCreated.add(1);
  } else if (res.status === 409) {
    idempConflict.add(1);
  }

  check(res, {
    'idempotency race returns 201 or 409': (r) => r.status === 201 || r.status === 409,
  });
}

