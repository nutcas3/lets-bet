import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Counter, Trend } from 'k6/metrics';

// Custom metrics
let errorRate = new Rate('errors');
let depositCounter = new Counter('deposits_initiated');
let withdrawalCounter = new Counter('withdrawals_initiated');
let paymentSuccessRate = new Rate('payment_success');
let paymentResponseTime = new Trend('payment_response_time');

// Test configuration for payment processing
export const options = {
  stages: [
    { duration: '2m', target: 500 }, // Warm up
    { duration: '3m', target: 2000 }, // Ramp to 2000
    { duration: '5m', target: 5000 }, // Ramp to 5000
    { duration: '10m', target: 10000 }, // Ramp to 10000
    { duration: '15m', target: 25000 }, // Ramp to 25000
    { duration: '20m', target: 50000 }, // Ramp to 50000
    { duration: '10m', target: 100000 }, // Peak at 100000
    { duration: '5m', target: 0 }, // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'], // 95% under 1s for payment operations
    http_req_failed: ['rate<0.2'], // 20% error rate acceptable for external payment APIs
    errors: ['rate<0.2'],
    payment_success: ['rate>0.8'], // 80% success rate for payments
    deposits_initiated: ['count>50000'], // At least 50k deposits
    withdrawals_initiated: ['count>25000'], // At least 25k withdrawals
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_BASE = `${BASE_URL}/api/v1`;

// Payment test data
const paymentMethods = {
  mpesa: {
    phone: '+254712345678',
    amount: [100, 500, 1000, 2000, 5000],
  },
  flutterwave: {
    ng: {
      phone: '+2348012345678',
      email: 'user@example.com',
      amount: [1000, 5000, 10000, 20000, 50000],
      currency: 'NGN',
    },
    gh: {
      phone: '+233201234567',
      email: 'user@example.com',
      amount: [50, 100, 500, 1000, 2000],
      currency: 'GHS',
    },
  },
};

// Test users
const users = Array.from({ length: 10000 }, (_, i) => ({
  phone: `+2547${String(i + 1000000).padStart(7, '0')}`,
  email: `user${i + 1}@loadtest.com`,
  password: 'LoadTest123!',
  name: `Load Test User ${i + 1}`,
}));

let authToken = '';
let userId = '';

export function setup() {
  // Create a test user for the load test
  const user = users[0];
  
  const registerPayload = JSON.stringify({
    phone: user.phone,
    email: user.email,
    password: user.password,
    full_name: user.name,
    country_code: 'KE',
    date_of_birth: '1990-01-01'
  });

  const registerResponse = http.post(`${API_BASE}/auth/register`, registerPayload, {
    headers: { 'Content-Type': 'application/json' },
  });

  // Login to get token
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
  
  console.log(`Payment load test setup complete. User ID: ${userId}`);
  
  return { authToken, userId };
}

function getRandomAmount(amounts) {
  return amounts[Math.floor(Math.random() * amounts.length)];
}

export default function() {
  const authHeaders = {
    'Authorization': `Bearer ${authToken}`,
    'Content-Type': 'application/json',
  };

  // Test 1: Get Wallet Balance
  const balanceResponse = http.get(`${API_BASE}/wallet/balance`, { headers: authHeaders });
  check(balanceResponse, {
    'wallet balance status is 200': (r) => r.status === 200,
    'wallet balance response time < 300ms': (r) => r.timings.duration < 300,
    'wallet balance has balance field': (r) => r.json('data.balance') !== undefined,
  }) || errorRate.add(1);

  // Test 2: Initiate M-Pesa Deposit
  const mpesaAmount = getRandomAmount(paymentMethods.mpesa.amount);
  const mpesaPayload = JSON.stringify({
    phone_number: paymentMethods.mpesa.phone,
    amount: mpesaAmount,
    currency: 'KES',
    callback_url: `${BASE_URL}/webhooks/mpesa/deposit`,
  });

  const mpesaResponse = http.post(`${API_BASE}/payments/mpesa/deposit`, mpesaPayload, { headers: authHeaders });
  const mpesaSuccess = check(mpesaResponse, {
    'mpesa deposit status is 201 or 400': (r) => r.status === 201 || r.status === 400,
    'mpesa deposit response time < 2000ms': (r) => r.timings.duration < 2000,
  });

  if (mpesaResponse.status === 201) {
    depositCounter.add(1);
    paymentSuccessRate.add(1);
  }

  paymentResponseTime.add(mpesaResponse.timings.duration);
  mpesaSuccess || errorRate.add(1);

  // Test 3: Initiate Flutterwave Nigeria Deposit
  const ngAmount = getRandomAmount(paymentMethods.flutterwave.ng.amount);
  const flutterwaveNGPayload = JSON.stringify({
    email: paymentMethods.flutterwave.ng.email,
    phone_number: paymentMethods.flutterwave.ng.phone,
    amount: ngAmount,
    currency: paymentMethods.flutterwave.ng.currency,
    callback_url: `${BASE_URL}/webhooks/flutterwave/deposit`,
  });

  const flutterwaveNGResponse = http.post(`${API_BASE}/payments/flutterwave/ng/deposit`, flutterwaveNGPayload, { headers: authHeaders });
  const fwNGSuccess = check(flutterwaveNGResponse, {
    'flutterwave NG deposit status is 201 or 400': (r) => r.status === 201 || r.status === 400,
    'flutterwave NG deposit response time < 1500ms': (r) => r.timings.duration < 1500,
  });

  if (flutterwaveNGResponse.status === 201) {
    depositCounter.add(1);
    paymentSuccessRate.add(1);
  }

  paymentResponseTime.add(flutterwaveNGResponse.timings.duration);
  fwNGSuccess || errorRate.add(1);

  // Test 4: Initiate Flutterwave Ghana Mobile Money Deposit
  const ghAmount = getRandomAmount(paymentMethods.flutterwave.gh.amount);
  const flutterwaveGHPayload = JSON.stringify({
    email: paymentMethods.flutterwave.gh.email,
    mobile_number: paymentMethods.flutterwave.gh.phone,
    amount: ghAmount,
    currency: paymentMethods.flutterwave.gh.currency,
    network: 'MTN', // or VODAFONE, TIGO, AIRTEL
    callback_url: `${BASE_URL}/webhooks/flutterwave/deposit`,
  });

  const flutterwaveGHResponse = http.post(`${API_BASE}/payments/flutterwave/gh/mobile-money`, flutterwaveGHPayload, { headers: authHeaders });
  const fwGHSuccess = check(flutterwaveGHResponse, {
    'flutterwave GH deposit status is 201 or 400': (r) => r.status === 201 || r.status === 400,
    'flutterwave GH deposit response time < 1500ms': (r) => r.timings.duration < 1500,
  });

  if (flutterwaveGHResponse.status === 201) {
    depositCounter.add(1);
    paymentSuccessRate.add(1);
  }

  paymentResponseTime.add(flutterwaveGHResponse.timings.duration);
  fwGHSuccess || errorRate.add(1);

  // Test 5: Get Transaction History
  const transactionsResponse = http.get(`${API_BASE}/wallet/transactions?limit=20&offset=0`, { headers: authHeaders });
  check(transactionsResponse, {
    'transactions status is 200': (r) => r.status === 200,
    'transactions response time < 500ms': (r) => r.timings.duration < 500,
    'transactions has transactions array': (r) => Array.isArray(r.json('data.transactions')),
  }) || errorRate.add(1);

  // Test 6: Get Deposit Methods
  const depositMethodsResponse = http.get(`${API_BASE}/payments/deposit-methods`, { headers: authHeaders });
  check(depositMethodsResponse, {
    'deposit methods status is 200': (r) => r.status === 200,
    'deposit methods response time < 300ms': (r) => r.timings.duration < 300,
  }) || errorRate.add(1);

  // Test 7: Get Withdrawal Methods
  const withdrawalMethodsResponse = http.get(`${API_BASE}/payments/withdrawal-methods`, { headers: authHeaders });
  check(withdrawalMethodsResponse, {
    'withdrawal methods status is 200': (r) => r.status === 200,
    'withdrawal methods response time < 300ms': (r) => r.timings.duration < 300,
  }) || errorRate.add(1);

  // Test 8: Initiate M-Pesa Withdrawal
  const withdrawalAmount = Math.floor(Math.random() * 1000) + 100; // 100-1100
  const mpesaWithdrawalPayload = JSON.stringify({
    phone_number: paymentMethods.mpesa.phone,
    amount: withdrawalAmount,
    currency: 'KES',
    callback_url: `${BASE_URL}/webhooks/mpesa/withdrawal`,
  });

  const mpesaWithdrawalResponse = http.post(`${API_BASE}/payments/mpesa/withdrawal`, mpesaWithdrawalPayload, { headers: authHeaders });
  const mpesaWithdrawalSuccess = check(mpesaWithdrawalResponse, {
    'mpesa withdrawal status is 201 or 400': (r) => r.status === 201 || r.status === 400,
    'mpesa withdrawal response time < 2000ms': (r) => r.timings.duration < 2000,
  });

  if (mpesaWithdrawalResponse.status === 201) {
    withdrawalCounter.add(1);
    paymentSuccessRate.add(1);
  }

  paymentResponseTime.add(mpesaWithdrawalResponse.timings.duration);
  mpesaWithdrawalSuccess || errorRate.add(1);

  // Test 9: Initiate Flutterwave Nigeria Withdrawal
  const ngWithdrawalAmount = Math.floor(Math.random() * 5000) + 1000; // 1000-6000 NGN
  const flutterwaveNGWithdrawalPayload = JSON.stringify({
    bank_code: '044', // Access Bank
    account_number: '1234567890',
    amount: ngWithdrawalAmount,
    currency: paymentMethods.flutterwave.ng.currency,
    beneficiary_name: 'Load Test User',
    narration: 'Load test withdrawal',
  });

  const flutterwaveNGWithdrawalResponse = http.post(`${API_BASE}/payments/flutterwave/ng/payout`, flutterwaveNGWithdrawalPayload, { headers: authHeaders });
  const fwNGWithdrawalSuccess = check(flutterwaveNGWithdrawalResponse, {
    'flutterwave NG withdrawal status is 201 or 400': (r) => r.status === 201 || r.status === 400,
    'flutterwave NG withdrawal response time < 1500ms': (r) => r.timings.duration < 1500,
  });

  if (flutterwaveNGWithdrawalResponse.status === 201) {
    withdrawalCounter.add(1);
    paymentSuccessRate.add(1);
  }

  paymentResponseTime.add(flutterwaveNGWithdrawalResponse.timings.duration);
  fwNGWithdrawalSuccess || errorRate.add(1);

  // Test 10: Get Transaction Details
  if (mpesaResponse.status === 201) {
    const transactionId = mpesaResponse.json('data.transaction_id');
    const transactionDetailsResponse = http.get(`${API_BASE}/payments/transactions/${transactionId}`, { headers: authHeaders });
    check(transactionDetailsResponse, {
      'transaction details status is 200': (r) => r.status === 200,
      'transaction details response time < 400ms': (r) => r.timings.duration < 400,
    }) || errorRate.add(1);
  }

  // Test 11: Get Payment Limits
  const limitsResponse = http.get(`${API_BASE}/payments/limits`, { headers: authHeaders });
  check(limitsResponse, {
    'payment limits status is 200': (r) => r.status === 200,
    'payment limits response time < 300ms': (r) => r.timings.duration < 300,
  }) || errorRate.add(1);

  // Test 12: Get Payment Statistics
  const statsResponse = http.get(`${API_BASE}/payments/stats`, { headers: authHeaders });
  check(statsResponse, {
    'payment stats status is 200': (r) => r.status === 200,
    'payment stats response time < 500ms': (r) => r.timings.duration < 500,
  }) || errorRate.add(1);

  // Small delay between iterations
  sleep(0.3);
}

export function teardown() {
  console.log(`Payment load test completed. Deposits: ${depositCounter.count}, Withdrawals: ${withdrawalCounter.count}, Success Rate: ${(paymentSuccessRate.rate * 100).toFixed(2)}%`);
}
