import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend } from 'k6/metrics';
import { BASE_URL, SLA_THRESHOLDS } from './lib/config.js';
import { randomInt } from './lib/helpers.js';

const stampedeLatency = new Trend('stampede_req_duration', true);
const warmCacheLatency = new Trend('warm_cache_req_duration', true);

export const options = {
  scenarios: {
    // 1A: 100 VUs hit /api/v1/products/1 simultaneously on cold cache to test singleflight
    stampede_spike: {
      executor: 'per-vu-iterations',
      vus: 100,
      iterations: 1,
      maxDuration: '30s',
      exec: 'stampedeSpike',
    },
    // 1B: Ramping VUs hitting warm Redis cache (list + detail)
    warm_cache_load: {
      executor: 'ramping-vus',
      startTime: '5s',
      startVUs: 20,
      stages: [
        { duration: '15s', target: 100 },
        { duration: '30s', target: 200 },
        { duration: '15s', target: 0 },
      ],
      exec: 'warmCacheLoad',
    },
  },
  thresholds: {
    http_req_duration: SLA_THRESHOLDS.cacheHit.http_req_duration,
    http_req_failed: SLA_THRESHOLDS.cacheHit.http_req_failed,
    warm_cache_req_duration: ['p(95)<25', 'p(99)<60'],
  },
};

export function stampedeSpike() {
  const res = http.get(`${BASE_URL}/api/v1/products/1`, {
    tags: { name: 'GET /api/v1/products/:id (stampede)' },
  });

  stampedeLatency.add(res.timings.duration);

  check(res, {
    'stampede status is 200': (r) => r.status === 200,
    'stampede returns product id 1': (r) => r.json('data.id') === 1,
  });
}

export function warmCacheLoad() {
  const productId = randomInt(1, 20);
  const resDetail = http.get(`${BASE_URL}/api/v1/products/${productId}`, {
    tags: { name: 'GET /api/v1/products/:id' },
  });
  warmCacheLatency.add(resDetail.timings.duration);

  check(resDetail, {
    'detail status is 200': (r) => r.status === 200,
  });

  const page = randomInt(1, 4);
  const resList = http.get(`${BASE_URL}/api/v1/products?page=${page}&pageSize=10`, {
    tags: { name: 'GET /api/v1/products' },
  });
  warmCacheLatency.add(resList.timings.duration);

  check(resList, {
    'list status is 200': (r) => r.status === 200,
  });

  sleep(0.1);
}

