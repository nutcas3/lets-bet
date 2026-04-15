# Betting Platform - Production-Ready Architecture

A **Tier-1 betting platform** built for **Kenya, Nigeria, and Ghana** with support for:
- **Sports Betting** (Single, Multi, System bets)
- **Crash Games** (Aviator-style with Provably Fair algorithm)
- **M-Pesa & Airtel Money** integration
- **Real-time odds** via WebSocket
- **BCLB Compliance** (KYC, Responsible Gaming, Tax calculations)

---

## Architecture Overview

### Multi-Tenant Microservices
```
┌─────────────────────────────────────────────────────────────┐
│                     API GATEWAY (Port 8080)                  │
│  HTTP/REST + WebSocket | Authentication | Rate Limiting      │
└─────────────────────────────────────────────────────────────┘
         │              │              │              │
    ┌────▼────┐    ┌───▼────┐    ┌───▼─────┐   ┌───▼─────┐
    │ WALLET  │    │ ENGINE │    │ GAMES   │   │SETTLEMENT│
    │ Service │    │Service │    │ Service │   │ Service  │
    └────┬────┘    └───┬────┘    └───┬─────┘   └───┬─────┘
         │              │              │              │
    ┌────▼──────────────▼──────────────▼──────────────▼────┐
    │         PostgreSQL (Multi-tenant with country_code)   │
    └───────────────────────────────────────────────────────┘
                          │
              ┌───────────┼───────────┐
              │           │           │
         ┌────▼────┐ ┌───▼────┐ ┌───▼─────┐
         │  Redis  │ │  NATS  │ │Cloudflare│
         │ (Cache) │ │ (Queue)│ │  (CDN)   │
         └─────────┘ └────────┘ └──────────┘
```

### Tech Stack
- **Backend:** Go 1.21+ (high concurrency, low latency)
- **Database:** PostgreSQL 15 (ACID transactions)
- **Cache:** Redis 7 (live odds, sessions, leaderboards)
- **Queue:** NATS (event-driven architecture)
- **WebSocket:** Gorilla WebSocket (real-time crash games)
- **Payments:** Safaricom Daraja API (M-Pesa), Airtel Money API

---

## Project Structure

```
betting-platform/
├── cmd/                    # Service entry points
│   ├── gateway/            # Public-facing API (HTTP + WebSocket)
│   ├── wallet/             # Balance, deposits, withdrawals
│   ├── engine/             # Betting logic and odds processing
│   ├── settlement/         # Winner payouts
│   └── games/              # Crash game engine
├── internal/
│   ├── core/               # GLOBAL CORE (country-agnostic)
│   │   ├── domain/         # Entities (User, Bet, Transaction, Game)
│   │   └── usecase/        # Business logic (PlaceBet, ProvablyFair)
│   ├── platform/           # Shared infrastructure
│   │   ├── db/             # Database connectors
│   │   ├── queue/          # NATS wrapper
│   │   └── auth/           # JWT authentication
│   ├── tenant/             # COUNTRY-SPECIFIC ADAPTERS
│   │   ├── ke/             # Kenya: M-Pesa, BCLB tax (15% GGR + 20% WHT)
│   │   ├── ng/             # Nigeria: Paystack, local gaming taxes
│   │   └── gh/             # Ghana: MTN MoMo, local compliance
│   └── games/              # Crash game engine + WebSocket hub
├── scripts/
│   ├── schema.sql          # Database schema
│   └── migrate.go          # Migration runner
├── deployments/            # Docker + Kubernetes
│   ├── ke-prod/            # Kenya production config
│   └── ng-prod/            # Nigeria production config
├── docker-compose.yml      # Local development environment
├── Makefile
└── README.md
```

---

## Quick Start

### 1. Prerequisites
- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 15 (via Docker)
- Redis 7 (via Docker)

### 2. Local Development Setup
```bash
# Clone the repository
git clone <your-repo-url>
cd betting-platform

# Start infrastructure (Postgres, Redis, NATS)
docker-compose up -d

# Run database migrations
make migrate

# Build all services
make build

# Start all services
make run-all
```

### 3. Test the API
```bash
# Health check
curl http://localhost:8080/health

# Register a user (mock endpoint)
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"phone":"0712345678","password":"SecurePass123"}'

# Check wallet balance
curl http://localhost:8080/api/v1/users/me/wallet
```

### 4. Connect to Crash Game WebSocket
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/games/crash-game-id');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Multiplier:', data.current_multiplier);
};

// Place a bet
ws.send(JSON.stringify({
  action: 'place_bet',
  amount: 100
}));

// Cashout
ws.send(JSON.stringify({
  action: 'cashout'
}));
```

---

## M-Pesa Integration (Kenya)

### Deposit Flow (STK Push)
```go
// User requests deposit of KES 500
mpesaClient.InitiateDeposit(ctx, "0712345678", 500, "DEP-123")

// M-Pesa sends prompt to user's phone
// User enters PIN
// Callback received at /api/mpesa/callback
// Wallet credited automatically
```

### Withdrawal Flow (B2C)
```go
// User requests withdrawal of KES 1000
mpesaClient.InitiateWithdrawal(ctx, "0712345678", 1000, "WTD-456")

// Wallet debited immediately
// M-Pesa processes payout
// Money sent to user's M-Pesa in ~30 seconds
```

### Configuration
```bash
# Environment variables for M-Pesa
MPESA_CONSUMER_KEY=your_consumer_key
MPESA_CONSUMER_SECRET=your_consumer_secret
MPESA_SHORTCODE=174379  # Your Paybill/Till number
MPESA_PASSKEY=your_passkey
MPESA_ENVIRONMENT=sandbox  # or production
```

---

## Crash Game (Provably Fair)

### How It Works
1. **Server generates seeds** before each round (SHA-256 hashed)
2. **Players place bets** during 5-second betting phase
3. **Multiplier increases** from 1.00x → crash point
4. **Players cashout** anytime before crash
5. **Game crashes** at predetermined point
6. **Seeds revealed** - players can verify fairness

### Example Round
```
Round 42:
- Server Seed Hash: 7a3f...e92c (public)
- Client Seed: player_combined_...
- Crash Point: 3.52x (hidden until crash)

Timeline:
00:00 - Betting phase starts
00:05 - Flight begins (1.00x → 1.01x → 1.02x...)
00:17 - User cashes out at 2.45x
00:22 - CRASH at 3.52x
00:23 - Server seed revealed for verification
```

---

## BCLB Compliance (Kenya)

### Required Features
- KYC Verification: National ID + KRA PIN validation  
- Self-Exclusion: Users can ban themselves for 1-12 months  
- Deposit Limits: Daily/weekly/monthly caps  
- Tax Deduction: 20% withholding tax on winnings  
- Audit Log: Every transaction stored for 7 years  
- Local Mirror: Read-only database replica for BCLB inspectors

### Tax Calculation
```go
// Example: User wins KES 10,000 from KES 500 stake
profit := 10000 - 500  // KES 9,500
tax := profit * 0.20   // KES 1,900 (20% WHT)
payout := 10000 - tax  // KES 8,100 sent to wallet
```

---

## Database Schema Highlights

### Optimistic Locking for Wallets
```sql
-- Prevents double-spending
UPDATE wallets 
SET balance = balance - 100, version = version + 1
WHERE user_id = '...' AND version = 5;
-- Fails if version changed (concurrent update)
```

### Multi-Tenant Design
```sql
-- Every table has country_code for isolation
SELECT * FROM bets WHERE country_code = 'KE';
SELECT * FROM transactions WHERE country_code = 'NG';
```

---

## Deployment

### Production (AWS)
```bash
# Build Docker images
make docker-build

# Deploy to Kenya region (af-south-1)
cd deployments/ke-prod
terraform apply

# Deploy to Nigeria region (eu-west-2)
cd deployments/ng-prod
terraform apply
```

### Environment Variables
```bash
# Database
DATABASE_URL=postgres://user:pass@host:5432/betting_db

# Redis
REDIS_URL=redis://host:6379

# M-Pesa (Kenya)
MPESA_CONSUMER_KEY=...
MPESA_CONSUMER_SECRET=...

# Country
COUNTRY_CODE=KE
CURRENCY=KES
```

---

## Performance Benchmarks

| Metric | Target | Status |
|--------|--------|--------|
| Bet Placement | < 200ms | |
| Wallet Update | < 50ms | |
| WebSocket Latency | < 100ms | |
| Concurrent Users | 100,000+ | |
| M-Pesa Payout | < 60s | |

---

## Next Steps

### Phase 1: Core Infrastructure (Completed)
- [x] Multi-tenant architecture
- [x] Database schema
- [x] M-Pesa integration
- [x] Crash game engine

### Phase 2: Production Hardening
- [ ] PostgreSQL repository implementations
- [ ] Redis caching layer
- [ ] JWT authentication
- [ ] Rate limiting
- [ ] Load testing (100k concurrent users)

### Phase 3: Advanced Features
- [ ] Live sports betting (Sportradar API)
- [ ] Edit-a-Bet feature
- [ ] Jackpots
- [ ] Virtual sports
- [ ] Admin dashboard

### Phase 4: Regulatory
- [ ] BCLB technical vetting
- [ ] Security audit
- [ ] Penetration testing
- [ ] GDPR compliance

---


## Support

For technical questions or deployment assistance, open an issue or contact the development team.

---

## License

Proprietary - All rights reserved
