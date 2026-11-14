import http from 'k6/http';
import { DEFAULT_HEADERS } from './config.js';

/**
 * Generates an RFC 4122 v4 UUID string for Idempotency-Key headers.
 * Uses Math.random() so it works in k6 without external dependencies.
 */
export function uuidv4() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

/**
 * Ensures a test user exists (register is idempotent across reruns) and logs in
 * to retrieve a valid JWT access token.
 */
export function ensureTestUser(baseUrl, email, password, name = 'K6 Load Tester') {
  const registerPayload = JSON.stringify({
    name,
    email,
    password,
  });

  // Ignore register errors if user already exists (400/409/500 depending on DB constraint)
  http.post(`${baseUrl}/api/v1/users/register`, registerPayload, {
    headers: DEFAULT_HEADERS,
    tags: { name: 'setup_register' },
  });

  const loginPayload = JSON.stringify({
    email,
    password,
  });

  const loginRes = http.post(`${baseUrl}/api/v1/users/login`, loginPayload, {
    headers: DEFAULT_HEADERS,
    tags: { name: 'setup_login' },
  });

  if (loginRes.status !== 200) {
    throw new Error(`Setup login failed for ${email}: status=${loginRes.status}, body=${loginRes.body}`);
  }

  const token = loginRes.json('data.access_token');
  if (!token) {
    throw new Error(`Setup login did not return data.access_token: body=${loginRes.body}`);
  }

  return token;
}

/**
 * Builds request params with JSON Content-Type and Bearer Authorization header.
 */
export function authHeaders(token, extra = {}) {
  return {
    headers: Object.assign(
      {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      extra
    ),
  };
}

/**
 * Returns a random integer between min and max (inclusive).
 */
export function randomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

