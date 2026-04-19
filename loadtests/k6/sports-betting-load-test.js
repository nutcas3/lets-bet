import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Counter } from 'k6/metrics';

// Custom metrics
let errorRate = new Rate('errors');
let betCounter = new Counter('bets_placed');
let winCounter = new Counter('bets_won');
let lossCounter = new Counter('bets_lost');

// Test configuration for sports betting
export const options = {
  stages: [
    { duration: '1m', target: 100 }, // Warm up
    { duration: '3m', target: 1000 }, // Ramp to 1000
    { duration: '5m', target: 5000 }, // Ramp to 5000
    { duration: '10m', target: 10000 }, // Ramp to 10000
    { duration: '15m', target: 25000 }, // Ramp to 25000
    { duration: '20m', target: 50000 }, // Ramp to 50000
    { duration: '10m', target: 100000 }, // Peak at 100000
    { duration: '5m', target: 0 }, // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<800'], // 95% under 800ms for sports betting
    http_req_failed: ['rate<0.15'], // 15% error rate acceptable for complex operations
    errors: ['rate<0.15'],
    bets_placed: ['count>10000'], // At least 10k bets placed
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_BASE = `${BASE_URL}/api/v1`;

// Sports betting test data
const sportsData = {
  football: [
    { home: 'Manchester United', away: 'Liverpool', odds: { home: 2.5, draw: 3.2, away: 2.8 } },
    { home: 'Arsenal', away: 'Chelsea', odds: { home: 1.8, draw: 3.5, away: 4.2 } },
    { home: 'Barcelona', away: 'Real Madrid', odds: { home: 2.1, draw: 3.3, away: 3.4 } },
  ],
  basketball: [
    { home: 'Lakers', away: 'Celtics', odds: { home: 1.9, draw: null, away: 1.9 } },
    { home: 'Warriors', away: 'Heat', odds: { home: 2.2, draw: null, away: 1.7 } },
  ],
  tennis: [
    { home: 'Nadal', away: 'Federer', odds: { home: 1.7, draw: null, away: 2.1 } },
    { home: 'Djokovic', away: 'Murray', odds: { home: 1.5, draw: null, away: 2.5 } },
  ],
};

// Test users
const users = Array.from({ length: 1000 }, (_, i) => ({
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
  
  console.log(`Sports betting test setup complete. User ID: ${userId}`);
  
  return { authToken, userId };
}

function getRandomSport() {
  const sports = Object.keys(sportsData);
  return sports[Math.floor(Math.random() * sports.length)];
}

function getRandomMatch(sport) {
  const matches = sportsData[sport];
  return matches[Math.floor(Math.random() * matches.length)];
}

function getRandomOutcome(match) {
  const outcomes = ['home'];
  if (match.odds.draw) outcomes.push('draw');
  outcomes.push('away');
  return outcomes[Math.floor(Math.random() * outcomes.length)];
}

function getOddsForOutcome(match, outcome) {
  if (outcome === 'home') return match.odds.home;
  if (outcome === 'draw') return match.odds.draw;
  if (outcome === 'away') return match.odds.away;
  return 1.0;
}

export default function() {
  const authHeaders = {
    'Authorization': `Bearer ${authToken}`,
    'Content-Type': 'application/json',
  };

  // Test 1: Get Available Sports Events
  const eventsResponse = http.get(`${API_BASE}/sports/events`, { headers: authHeaders });
  check(eventsResponse, {
    'sports events status is 200': (r) => r.status === 200,
    'sports events response time < 500ms': (r) => r.timings.duration < 500,
    'sports events has events array': (r) => Array.isArray(r.json('data.events')),
  }) || errorRate.add(1);

  // Test 2: Get Live Events
  const liveEventsResponse = http.get(`${API_BASE}/sports/events/live`, { headers: authHeaders });
  check(liveEventsResponse, {
    'live events status is 200': (r) => r.status === 200,
    'live events response time < 400ms': (r) => r.timings.duration < 400,
  }) || errorRate.add(1);

  // Test 3: Get Upcoming Events
  const upcomingResponse = http.get(`${API_BASE}/sports/events/upcoming?hours=24`, { headers: authHeaders });
  check(upcomingResponse, {
    'upcoming events status is 200': (r) => r.status === 200,
    'upcoming events response time < 600ms': (r) => r.timings.duration < 600,
  }) || errorRate.add(1);

  // Test 4: Get Event Details
  const sport = getRandomSport();
  const match = getRandomMatch(sport);
  const mockEventId = `event-${sport}-${Math.random().toString(36).substr(2, 9)}`;
  
  const eventDetailsResponse = http.get(`${API_BASE}/sports/events/${mockEventId}`, { headers: authHeaders });
  check(eventDetailsResponse, {
    'event details status is 200 or 404': (r) => r.status === 200 || r.status === 404,
    'event details response time < 300ms': (r) => r.timings.duration < 300,
  }) || errorRate.add(1);

  // Test 5: Get Betting Markets for Event
  const marketsResponse = http.get(`${API_BASE}/sports/events/${mockEventId}/markets`, { headers: authHeaders });
  check(marketsResponse, {
    'markets status is 200 or 404': (r) => r.status === 200 || r.status === 404,
    'markets response time < 400ms': (r) => r.timings.duration < 400,
  }) || errorRate.add(1);

  // Test 6: Place Sports Bet
  const outcome = getRandomOutcome(match);
  const odds = getOddsForOutcome(match, outcome);
  const betAmount = Math.floor(Math.random() * 500) + 50; // 50-550

  const betPayload = JSON.stringify({
    event_id: mockEventId,
    market_type: 'MATCH_WINNER',
    outcome: outcome,
    odds: odds,
    amount: betAmount,
    currency: 'KES',
  });

  const betResponse = http.post(`${API_BASE}/sports/bets`, betPayload, { headers: authHeaders });
  const betSuccess = check(betResponse, {
    'sports bet status is 201 or 400': (r) => r.status === 201 || r.status === 400,
    'sports bet response time < 800ms': (r) => r.timings.duration < 800,
  });

  if (betResponse.status === 201) {
    betCounter.add(1);
  }

  betSuccess || errorRate.add(1);

  // Test 7: Get User's Sports Bets
  const userBetsResponse = http.get(`${API_BASE}/sports/bets?limit=20&offset=0`, { headers: authHeaders });
  check(userBetsResponse, {
    'user bets status is 200': (r) => r.status === 200,
    'user bets response time < 500ms': (r) => r.timings.duration < 500,
    'user bets has bets array': (r) => Array.isArray(r.json('data.bets')),
  }) || errorRate.add(1);

  // Test 8: Get Bet Details
  if (betResponse.status === 201) {
    const betId = betResponse.json('data.id');
    const betDetailsResponse = http.get(`${API_BASE}/sports/bets/${betId}`, { headers: authHeaders });
    check(betDetailsResponse, {
      'bet details status is 200': (r) => r.status === 200,
      'bet details response time < 300ms': (r) => r.timings.duration < 300,
    }) || errorRate.add(1);

    // Simulate bet settlement (random win/loss)
    if (Math.random() > 0.5) {
      winCounter.add(1);
    } else {
      lossCounter.add(1);
    }
  }

  // Test 9: Get Betting History
  const historyResponse = http.get(`${API_BASE}/sports/bets/history?limit=50&offset=0`, { headers: authHeaders });
  check(historyResponse, {
    'betting history status is 200': (r) => r.status === 200,
    'betting history response time < 600ms': (r) => r.timings.duration < 600,
  }) || errorRate.add(1);

  // Test 10: Get Live Scores
  const scoresResponse = http.get(`${API_BASE}/sports/scores/live`, { headers: authHeaders });
  check(scoresResponse, {
    'live scores status is 200': (r) => r.status === 200,
    'live scores response time < 400ms': (r) => r.timings.duration < 400,
  }) || errorRate.add(1);

  // Test 11: Get Standings/League Table
  const standingsResponse = http.get(`${API_BASE}/sports/standings/football/premier-league`, { headers: authHeaders });
  check(standingsResponse, {
    'standings status is 200 or 404': (r) => r.status === 200 || r.status === 404,
    'standings response time < 500ms': (r) => r.timings.duration < 500,
  }) || errorRate.add(1);

  // Test 12: Get User Statistics
  const statsResponse = http.get(`${API_BASE}/sports/stats/user`, { headers: authHeaders });
  check(statsResponse, {
    'user stats status is 200': (r) => r.status === 200,
    'user stats response time < 400ms': (r) => r.timings.duration < 400,
  }) || errorRate.add(1);

  // Small delay between iterations
  sleep(0.2);
}

export function teardown() {
  console.log(`Sports betting load test completed. Total bets: ${betCounter.count}, Wins: ${winCounter.count}, Losses: ${lossCounter.count}`);
}
