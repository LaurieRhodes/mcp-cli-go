# Market Analysis & Trading Signals

> **For:** Traders, Portfolio Managers, Quantitative Analysts, Financial Advisors  
> **Purpose:** Generate data-driven trade signals through multi-source analysis and template composition

---

## What This Showcase Contains

This section demonstrates how templates automate sophisticated market analysis by composing signals from multiple data sources. Uses **expert-informed weighting** based on what actually moves stock prices, not academic theory.

### Available Use Cases

**Market Analysis:**

1. **[Nasdaq-100 Trade Signals](use-cases/nasdaq100-signals.md)** - Multi-factor analysis for tech stock trading
2. **Macro Regime Analysis** - Fed policy and risk environment assessment
3. **Options Flow Intelligence** - Smart money tracking via unusual options activity
4. **Earnings Momentum Analysis** - Estimate revisions and guidance trends

---

## Why Templates Matter for Market Analysis

### 1. Multi-Source Signal Aggregation

**The Challenge:** Making trading decisions requires synthesizing data from 10+ disparate sources. No single indicator tells the whole story.

**Template Solution:** Compose signals from multiple specialized workflows

```yaml
# Master template aggregates sub-analyses
name: nasdaq100_trade_signal

steps:
  # Stage 1: Macro environment (35% weight)
  - name: macro_regime
    template: macro_regime_analysis
    tools: [fred-api, treasury-api, vix-api]
    output: macro_signal
    # Fed policy, yields, VIX, credit spreads

  # Stage 2: Options flow (25% weight)
  - name: options_flow
    template: options_flow_analysis
    tools: [unusual-whales, flowalgo, cboe-api]
    output: options_signal
    # Smart money positioning, dark pools, put/call

  # Stage 3: Earnings dynamics (20% weight)
  - name: earnings_momentum
    template: earnings_revision_analysis
    tools: [factset-api, estimize-api]
    output: earnings_signal
    # Estimate revisions, guidance trends

  # Stage 4: Relative strength (10% weight)
  - name: relative_performance
    template: relative_strength_analysis
    tools: [yahoo-finance, sector-etf-data]
    output: relative_signal
    # vs QQQ, sector rotation

  # Stage 5: Technical confirmation (5% weight)
  - name: technical_levels
    template: technical_confirmation
    tools: [alpha-vantage, tradingview]
    output: technical_signal
    # Volume, key levels, MA

  # Stage 6: Insider activity (5% weight)
  - name: insider_tracking
    template: insider_activity_analysis
    tools: [sec-edgar, openinsider]
    output: insider_signal
    # Form 4 filings, cluster buying
```

**Impact:**

- Manual analysis: 4 hours per stock
- Automated: 3 minutes per stock
- **Coverage: 1 stock/day → 20 stocks/day**

---

### 2. Expert-Informed Signal Weighting

**What actually moves Nasdaq-100 stocks:**

```
Macro/Fed Policy: 35%     ← Interest rates drive discount rate
Options Flow: 25%          ← Smart money knows first  
Earnings Momentum: 20%     ← Estimate revisions > historical
Relative Strength: 10%     ← Context vs QQQ/sector
Technicals: 5%             ← Timing only
Insiders: 5%               ← Delayed, limited value
```

**Why macro gets 35%:**

2022 example:

- Fed raised rates from 0% → 4.5%
- Nasdaq dropped 33% (discount rate killed growth stocks)
- Individual stock fundamentals didn't matter
- **Lesson: When macro bad, even great stocks fall**

2023 example:

- Fed paused rate hikes
- Nasdaq rallied 54%
- Rising tide lifted all boats
- **Lesson: Macro drives 35-40% of returns**

---

### 3. Real-World Tool Integration

**Critical data sources with specific tools:**

**FREE Tier (Start Here):**

```yaml
servers:
  # Macro data
  fred-api:
    command: "fred-mcp-server"
    env:
      API_KEY: "${FRED_KEY}"  # Free from stlouisfed.org
    provides:
      - Fed Funds rate
      - CPI/PCE inflation
      - GDP growth
      - Unemployment
      - Fed balance sheet

  # Market data
  yahoo-finance:
    command: "yahoo-finance-mcp-server"
    provides:
      - Stock prices (real-time delayed 15min)
      - Historical OHLCV
      - Basic fundamentals
      - Volume data

  # SEC filings
  sec-edgar:
    command: "sec-edgar-mcp-server"
    provides:
      - Form 4 (insider trades)
      - Form 13F (institutional holdings)
      - 10-Q/10-K (financials)
      - 8-K (material events)
```

**Recommended Tier ($100/month - Worth It):**

```yaml
servers:
  # Options flow - THE EDGE
  unusual-whales:
    command: "unusual-whales-mcp-server"
    env:
      API_KEY: "${UW_KEY}"  # $50/month
    provides:
      - Unusual options activity
      - Dark pool prints
      - Put/Call ratios
      - Options flow alerts

  # Earnings estimates
  estimize:
    command: "estimize-mcp-server"
    env:
      API_KEY: "${ESTIMIZE_KEY}"  # $50/month
    provides:
      - Crowdsourced estimates
      - Estimate revisions
      - Earnings surprise predictions
      - Guidance expectations
```

**Professional Tier ($$$$):**

```yaml
servers:
  # Institutional-grade data
  factset:
    command: "factset-mcp-server"
    env:
      API_KEY: "${FACTSET_KEY}"  # Enterprise pricing
    provides:
      - Consensus estimates
      - Detailed revisions
      - Institutional ownership
      - Comprehensive fundamentals

  bloomberg:
    command: "bloomberg-mcp-server"
    env:
      TERMINAL_ID: "${BLOOMBERG_ID}"  # $2,000/month
    provides:
      - Real-time data
      - News feeds
      - Analytics
      - Everything
```

**Recommendation for retail traders:**

- Start: FREE tier only
- After profitable: Add $100/month (Unusual Whales + Estimize)
- Skip Bloomberg unless managing millions

---

### 4. Volatility-Adjusted Position Sizing

**The problem with fixed stops:**

```
Stock A: $100, ATR $1 (1% volatility)
12% stop = $88 (12 ATR away - very wide)

Stock B: $100, ATR $5 (5% volatility)  
12% stop = $88 (2.4 ATR away - too tight, will get stopped out)
```

**Template solution:**

```yaml
- name: calculate_position_size
  tools: [alpha-vantage]  # Get ATR
  prompt: |
    Calculate ATR-based position size:

    Ticker: {{ticker}}
    Price: {{price}}
    ATR (14-day): {{atr}}
    Portfolio: {{portfolio_value}}
    Risk per trade: 1%

    ## Method 1: Fixed Dollar Risk
    Risk amount = Portfolio × 1% = {{portfolio_value * 0.01}}
    Stop distance = 3 × ATR = {{atr * 3}}
    Position size = Risk / Stop = {{(portfolio_value * 0.01) / (atr * 3)}} shares

    ## Method 2: Volatility Adjustment
    If ATR/Price > 0.03 (high vol):
      Reduce position by 50%
    If ATR/Price < 0.01 (low vol):
      Increase position by 25%

    Return: shares, dollar_amount, risk_pct
```

**Example:**

```
NVDA: $500, ATR $25 (5% - high vol)
→ 3×ATR stop = $75
→ Risk $1,000 / $75 = 13 shares
→ High vol adjustment: 13 × 0.5 = 6 shares
→ Position: $3,000 (3% of $100K portfolio)

KO: $60, ATR $0.60 (1% - low vol)
→ 3×ATR stop = $1.80
→ Risk $1,000 / $1.80 = 555 shares
→ Low vol adjustment: 555 × 1.25 = 694 shares  
→ Position: $41,640 (41% of portfolio - too big!)
→ Cap at 10% = 166 shares

Result: Smaller positions in volatile stocks, larger in stable stocks
```

---

## Quick Start

### 1. Set Up Free Data Sources

```bash
# Install MCP servers
npm install -g fred-mcp-server
npm install -g yahoo-finance-mcp-server
npm install -g sec-edgar-mcp-server

# Get free API keys
# FRED: https://fred.stlouisfed.org/docs/api/api_key.html
# SEC: No key needed
# Yahoo: No key needed

# Configure
export FRED_API_KEY="your_key_here"
```

### 2. Run Your First Analysis

```bash
# Analyze AAPL
mcp-cli --template nasdaq100_trade_signal --input-data "{
  \"ticker\": \"AAPL\",
  \"portfolio_value\": 100000
}"
```

### 3. Upgrade to Paid Data (Optional)

```bash
# Add Unusual Whales ($50/month)
npm install -g unusual-whales-mcp-server
export UW_API_KEY="your_key_here"

# Add Estimize ($50/month)
npm install -g estimize-mcp-server
export ESTIMIZE_API_KEY="your_key_here"
```

---

## Integration Patterns

### Pattern 1: Daily Pre-Market Screener

```yaml
name: daily_nasdaq100_screener

schedule:
  frequency: daily
  day: weekday
  time: "08:00"

steps:
  # Check macro first
  - name: macro_check
    template: macro_regime_analysis
    tools: [fred-api, vix-api]
    output: macro

  # Screen all 100 stocks
  - name: screen_stocks
    condition: "{{macro.signal}} != 'RISK_OFF'"
    for_each: "{{nasdaq100_tickers}}"
    template: nasdaq100_trade_signal
    parallel:
      max_concurrent: 10
    output: signals

  # Filter top opportunities
  - name: filter
    prompt: |
      Top 5 stocks where:
      - Signal: BUY or STRONG_BUY
      - Confidence: >70
      - Options flow: Bullish
      - Macro: RISK_ON
    output: opportunities

  # Send to Slack
  - name: notify
    tools: [slack-api]
    prompt: "Send watchlist to #trading channel"
```

---

## Template Library

**Master Workflows:**

- `nasdaq100_trade_signal.yaml` - Complete analysis
- `daily_stock_screener.yaml` - Scan 100 stocks

**Sub-Templates:**

- `macro_regime_analysis.yaml` - Fed, VIX, yields
- `options_flow_analysis.yaml` - Smart money
- `earnings_revision_analysis.yaml` - Estimates
- `relative_strength_analysis.yaml` - vs QQQ

---

## Honest Limitations

**❌ Can't predict:**

- Black swans (COVID, flash crashes)
- Fed surprises
- Earnings shocks

**❌ Won't beat:**

- High-frequency traders
- Insider traders  
- Market makers

**✅ Does well at:**

- Systematic analysis
- Multi-source synthesis
- Early signal detection
- Risk management

---

**Market analysis with AI: Synthesize macro, options flow, earnings, and technicals into systematic trade signals.**

Remember: No system wins 100%. Always use stops and size positions properly.
