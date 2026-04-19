# Load Testing with k6

This directory contains k6 load testing scripts for the betting platform. The tests are designed to simulate real-world usage patterns and validate the platform's performance under various load conditions.

## Prerequisites

- [k6](https://k6.io/docs/getting-started/installation/) installed
- Node.js (for running setup scripts)
- Access to the betting platform API

## Test Scenarios

### 1. Basic Load Test (`basic-load-test.js`)
Tests core platform functionality including:
- Health checks
- Wallet operations
- Game creation and betting
- Transaction history
- User profile management

**Target:** Up to 100,000 concurrent users
**Duration:** ~30 minutes with gradual ramp-up

### 2. Sports Betting Load Test (`sports-betting-load-test.js`)
Focuses on sports betting specific functionality:
- Sports events and markets
- Live betting scenarios
- Bet placement and settlement
- Odds updates and live scores

**Target:** Up to 100,000 concurrent users
**Duration:** ~40 minutes with gradual ramp-up

### 3. Payment Load Test (`payment-load-test.js`)
Tests payment processing capabilities:
- M-Pesa deposits and withdrawals
- Flutterwave NG/GH transactions
- Transaction history and verification
- Payment limits and statistics

**Target:** Up to 100,000 concurrent users
**Duration:** ~45 minutes with gradual ramp-up

## Configuration

The `config.json` file contains configuration for different environments:
- **Development:** Local testing with limited users
- **Staging:** Pre-production testing
- **Production:** Full-scale production testing

## Running Tests

### Basic Usage

```bash
# Set environment variables
export BASE_URL="http://localhost:8080"

# Run basic load test
k6 run loadtests/k6/basic-load-test.js

# Run sports betting test
k6 run loadtests/k6/sports-betting-load-test.js

# Run payment test
k6 run loadtests/k6/payment-load-test.js
```

### With Configuration File

```bash
# Use config for specific environment
k6 run --config loadtests/k6/config.json loadtests/k6/basic-load-test.js

# Override base URL
k6 run -e BASE_URL=https://staging.betting-platform.com loadtests/k6/basic-load-test.js
```

### Advanced Options

```bash
# Run with specific VUs and duration
k6 run --vus 1000 --duration 10m loadtests/k6/basic-load-test.js

# Run with custom options
k6 run --options loadtests/k6/options.json loadtests/k6/basic-load-test.js

# Generate HTML report
k6 run --out html=report.html loadtests/k6/basic-load-test.js
```

## Performance Thresholds

The tests include the following performance thresholds:

- **Response Time:** 95% of requests under 500ms
- **Error Rate:** Less than 10% failed requests
- **Success Rate:** Payment operations >80% success rate
- **Throughput:** Minimum transactions per second based on user load

## Metrics and Monitoring

### Custom Metrics

Each test tracks custom metrics:
- `errors`: Error rate
- `bets_placed`: Number of bets placed
- `bets_won`: Number of winning bets
- `bets_lost`: Number of losing bets
- `deposits_initiated`: Number of deposit attempts
- `withdrawals_initiated`: Number of withdrawal attempts
- `payment_success`: Payment success rate
- `payment_response_time`: Payment operation response times

### Built-in k6 Metrics

k6 automatically tracks:
- HTTP request duration
- HTTP request count
- Virtual users (VUs)
- Data sent/received
- Iteration duration

## Test Data

### Users

The tests generate synthetic user data:
- Phone numbers: +2547xxxxxxx format
- Emails: userX@loadtest.com
- Passwords: LoadTest123!
- Names: Load Test User X

### Payment Data

Test payment data includes:
- M-Pesa: Various amount ranges (100-5000 KES)
- Flutterwave NG: Amount ranges (1000-50000 NGN)
- Flutterwave GH: Amount ranges (50-2000 GHS)

### Sports Data

Mock sports data for betting tests:
- Football: Premier League matches
- Basketball: NBA games
- Tennis: ATP matches

## Best Practices

### Before Running Tests

1. **Verify Environment**: Ensure the target environment is running and accessible
2. **Check Dependencies**: Verify all external services (payment providers, etc.) are available
3. **Monitor Resources**: Monitor server resources during tests
4. **Clear Data**: Clear test data if needed

### During Tests

1. **Monitor Metrics**: Watch real-time metrics in k6 output
2. **Check Logs**: Monitor application logs for errors
3. **Resource Usage**: Monitor CPU, memory, and database connections
4. **External Services**: Monitor payment provider APIs

### After Tests

1. **Analyze Results**: Review k6 reports and metrics
2. **Identify Bottlenecks**: Look for performance issues
3. **Optimize**: Address identified issues
4. **Re-test**: Run tests again to verify improvements

## Troubleshooting

### Common Issues

1. **Connection Refused**: Check if the server is running
2. **High Error Rates**: Verify API endpoints and authentication
3. **Slow Response Times**: Check database performance and external services
4. **Memory Issues**: Monitor k6 memory usage with `--max-memory-usage`

### Debug Mode

```bash
# Enable debug output
k6 run --verbose loadtests/k6/basic-load-test.js

# Run with HTTP debugging
k6 run --http-debug loadtests/k6/basic-load-test.js
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: Load Tests
on: [push, pull_request]

jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Setup k6
        run: |
          sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
          echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
          sudo apt-get update
          sudo apt-get install k6
      - name: Run Load Tests
        run: k6 run --vus 100 --duration 2m loadtests/k6/basic-load-test.js
```

## Security Considerations

- Use test credentials only
- Never use production credentials in load tests
- Ensure test data doesn't contain sensitive information
- Clean up test data after tests
- Monitor for any security alerts during high load

## Reporting

### HTML Reports

```bash
k6 run --out html=report.html loadtests/k6/basic-load-test.js
```

### JSON Reports

```bash
k6 run --out json=results.json loadtests/k6/basic-load-test.js
```

### Cloud Integration

```bash
# Upload results to k6 Cloud
k6 cloud --token YOUR_TOKEN loadtests/k6/basic-load-test.js
```

## Load Testing Strategy

### Phase 1: Baseline Testing
- Small scale tests (100-1000 users)
- Identify baseline performance
- Validate test scripts

### Phase 2: Scale Testing
- Medium scale tests (1000-10000 users)
- Identify performance bottlenecks
- Optimize based on results

### Phase 3: Stress Testing
- Large scale tests (10000-100000 users)
- Test system limits
- Validate scalability

### Phase 4: Endurance Testing
- Long-running tests (several hours)
- Test for memory leaks
- Validate stability

## Support

For issues with load testing:
1. Check k6 documentation: https://k6.io/docs/
2. Review application logs
3. Monitor system resources
4. Contact the development team

## License

This load testing framework is part of the betting platform project and follows the same license terms.
