export const BASE_URL = __ENV.BASE_URL || 'http://localhost:3001';

export const DEFAULT_HEADERS = {
  'Content-Type': 'application/json',
};

export const SLA_THRESHOLDS = {
  cacheHit: {
    http_req_duration: ['p(95)<30', 'p(99)<100'],
    http_req_failed: ['rate<0.01'],
  },
  authBcrypt: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.01'],
  },
  transaction: {
    http_req_duration: ['p(95)<400', 'p(99)<800'],
    http_req_failed: ['rate<0.05'],
  },
  mixedLoad: {
    http_req_duration: ['p(95)<300', 'p(99)<750'],
    http_req_failed: ['rate<0.02'],
  },
};

