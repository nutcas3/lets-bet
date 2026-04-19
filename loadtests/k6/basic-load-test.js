import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
let errorRate = new Rate('errors');

// Test configuration
export const options = {
  stages: [
    { duration: '2m', target: 1000 }, // Ramp up to 1000 users
    { duration: '5m', target: 1000 }, // Stay at 1000 users
    { duration: '2m', target: 5000 }, // Ramp up to 5000 users
    { duration: '5m', target: 5000 }, // Stay at 5000 users
    { duration: '2m', target: 10000 }, // Ramp up to 10000 users
    { duration: '5m', target: 10000 }, // Stay at 10000 users
    { duration: '2m', target: 50000 }, // Ramp up to 50000 users
    { duration: '5m', target: 50000 }, // Stay at 50000 users
    { duration: '2m', target: 100000 }, // Ramp up to 100000 users
    { duration: '10m', target: 100000 }, // Stay at 100000 users
    { duration: '2m', target: 0 }, // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
    http_req_failed: ['rate<0.1'], // Error rate should be below 10%
    errors: ['rate<0.1'], // Custom error rate below 10%
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_BASE = `${BASE_URL}/api/v1`;

// Test data
const users = [
  { phone: '+254712345678', email: 'user1@example.com', password: 'password123' },
  { phone: '+254712345679', email: 'user2@example.com', password: 'password123' },
  { phone: '+254712345680', email: 'user3@example.com', password: 'password123' },
  { phone: '+254712345681', email: 'user4@example.com', password: 'password123' },
  { phone: '+254712345682', email: 'user5@example.com', password: 'password123' },
];

let authToken = '';
let userId = '';
let walletId = '';

export function setup() {
  // Setup phase - create test user and get auth token
  const user = users[0];
  
  // Register user
  const registerPayload = JSON.stringify({
    phone: user.phone,
    email: user.email,
    password: user.password,
    full_name: 'Load Test User',
    country_code: 'KE',
    date_of_birth: '1990-01-01'
  });

  const registerResponse = http.post(`${API_BASE}/auth/register`, registerPayload, {
    headers: { 'Content-Type': 'application/json' },
  });

  if (registerResponse.status !== 201) {
    console.log('Registration failed, trying to login...');
  }

  // Login user
  const loginPayload = JSON.stringify({
    phone: user.phone,
    password: user.password,
  });

  const loginResponse = http.post(`${API_BASE}/auth/login`, loginPayload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(loginResponse, {
    'login successful': (r) => r.status === 200,
    'token received': (r) => r.json('data.token') !== undefined,
  });

  authToken = loginResponse.json('data.token');
  userId = loginResponse.json('data.user.id');
  
  console.log(`Setup complete. User ID: ${userId}`);
  
  return { authToken, userId };
}

export default function() {
  const authHeaders = {
    'Authorization': `Bearer ${authToken}`,
    'Content-Type': 'application/json',
  };

  // Test 1: Health Check
  const healthResponse = http.get(`${BASE_URL}/health`, { headers: authHeaders });
  check(healthResponse, {
    'health check status is 200': (r) => r.status === 200,
    'health check response time < 100ms': (r) => r.timings.duration < 100,
  }) || errorRate.add(1);

  // Test 2: Get Wallet Balance
  const walletResponse = http.get(`${API_BASE}/wallet/balance`, { headers: authHeaders });
  check(walletResponse, {
    'wallet balance status is 200': (r) => r.status === 200,
    'wallet balance response time < 200ms': (r) => r.timings.duration < 200,
    'wallet balance has balance field': (r) => r.json('data.balance') !== undefined,
  }) || errorRate.add(1);

  // Test 3: Get Available Games
  const gamesResponse = http.get(`${API_BASE}/games/available`, { headers: authHeaders });
  check(gamesResponse, {
    'games list status is 200': (r) => r.status === 200,
    'games list response time < 300ms': (r) => r.timings.duration < 300,
    'games list has games array': (r) => Array.isArray(r.json('data.games')),
  }) || errorRate.add(1);

  // Test 4: Create Game (if available)
  const createGamePayload = JSON.stringify({
    game_type: 'crash',
    min_bet: 10.00,
    max_bet: 1000.00,
    max_multiplier: 100.00,
    country_code: 'KE',
  });

  const createGameResponse = http.post(`${API_BASE}/games`, createGamePayload, { headers: authHeaders });
  const gameId = createGameResponse.status === 201 ? createGameResponse.json('data.id') : null;

  // Test 5: Place Bet
  if (gameId) {
    const betPayload = JSON.stringify({
      game_id: gameId,
      amount: 50.00,
      auto_cashout_multiplier: 2.5,
    });

    const betResponse = http.post(`${API_BASE}/games/${gameId}/bets`, betPayload, { headers: authHeaders });
    check(betResponse, {
      'bet placement status is 201': (r) => r.status === 201,
      'bet placement response time < 500ms': (r) => r.timings.duration < 500,
      'bet has id field': (r) => r.json('data.id') !== undefined,
    }) || errorRate.add(1);

    // Test 6: Cashout Bet
    if (betResponse.status === 201) {
      const betId = betResponse.json('data.id');
      const cashoutPayload = JSON.stringify({
        multiplier: 2.5,
      });

      const cashoutResponse = http.post(`${API_BASE}/games/${gameId}/bets/${betId}/cashout`, cashoutPayload, { headers: authHeaders });
      check(cashoutResponse, {
        'cashout status is 200': (r) => r.status === 200,
        'cashout response time < 300ms': (r) => r.timings.duration < 300,
      }) || errorRate.add(1);
    }
  }

  // Test 7: Get Transaction History
  const transactionsResponse = http.get(`${API_BASE}/wallet/transactions?limit=10&offset=0`, { headers: authHeaders });
  check(transactionsResponse, {
    'transactions status is 200': (r) => r.status === 200,
    'transactions response time < 400ms': (r) => r.timings.duration < 400,
    'transactions has transactions array': (r) => Array.isArray(r.json('data.transactions')),
  }) || errorRate.add(1);

  // Test 8: Get User Profile
  const profileResponse = http.get(`${API_BASE}/user/profile`, { headers: authHeaders });
  check(profileResponse, {
    'profile status is 200': (r) => r.status === 200,
    'profile response time < 200ms': (r) => r.timings.duration < 200,
    'profile has user data': (r) => r.json('data.id') === userId,
  }) || errorRate.add(1);

  // Test 9: Get Metrics
  const metricsResponse = http.get(`${BASE_URL}/metrics`, { headers: authHeaders });
  check(metricsResponse, {
    'metrics status is 200': (r) => r.status === 200,
    'metrics response time < 100ms': (r) => r.timings.duration < 100,
  }) || errorRate.add(1);

  // Small delay between iterations
  sleep(0.1);
}

export function teardown() {
  // Cleanup phase
  console.log('Load test completed');
}
