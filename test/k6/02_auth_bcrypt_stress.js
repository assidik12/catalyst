import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend } from 'k6/metrics';
import { BASE_URL, DEFAULT_HEADERS, SLA_THRESHOLDS } from './lib/config.js';
import { ensureTestUser } from './lib/helpers.js';

const loginLatency = new Trend('bcrypt_login_duration', true);

const AUTH_EMAIL = __ENV.K6_AUTH_EMAIL || 'k6_auth@catalyst.dev';
const AUTH_PASSWORD = __ENV.K6_AUTH_PASSWORD || 'Password123!';

export const options = {
  scenarios: {
    auth_bcrypt_stress: {
      executor: 'ramping-vus',
      startVUs: 5,
      stages: [
        { duration: '15s', target: 10 },
        { duration: '20s', target: 30 },
        { duration: '15s', target: 50 },
        { duration: '10s', target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_duration: SLA_THRESHOLDS.authBcrypt.http_req_duration,
    http_req_failed: SLA_THRESHOLDS.authBcrypt.http_req_failed,
  },
};

export function setup() {
  ensureTestUser(BASE_URL, AUTH_EMAIL, AUTH_PASSWORD, 'K6 Auth Stress User');
  return { email: AUTH_EMAIL, password: AUTH_PASSWORD };
}

export default function (data) {
  const payload = JSON.stringify({
    email: data.email,
    password: data.password,
  });

  const res = http.post(`${BASE_URL}/api/v1/users/login`, payload, {
    headers: DEFAULT_HEADERS,
    tags: { name: 'POST /api/v1/users/login' },
  });

  loginLatency.add(res.timings.duration);

  check(res, {
    'login status is 200': (r) => r.status === 200,
    'access_token present': (r) => Boolean(r.json('data.access_token')),
  });

  sleep(0.2);
}

