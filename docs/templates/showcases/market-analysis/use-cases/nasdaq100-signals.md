# Nasdaq-100 Trade Signals

> **Template:** [nasdaq100_trade_signal.yaml](../templates/nasdaq100_trade_signal.yaml)  
> **Workflow:** Macro → Options → Earnings → Relative → Technical → Insider → Signal  
> **Best For:** Systematic analysis of Nasdaq-100 tech stocks with proper risk management

---

## Problem Description

### The Multi-Source Analysis Challenge

**Trading Nasdaq-100 stocks requires analyzing 10+ data sources:**

```
Manual analysis for single stock (e.g., AAPL):

1. Check Fed policy and macro (30 min)
   - FRED: Fed Funds rate, CPI, GDP
   - Treasury: Yield curve
   - VIX: Fear gauge

2. Search for unusual options activity (45 min)
   - Check multiple options platforms
   - Look for dark pool prints
   - Analyze put/call ratios

3. Research earnings estimates (30 min)
   - Find consensus estimates
   - Track recent revisions
   - Read guidance commentary

4. Calculate relative strength (20 min)
   - Compare to QQQ performance
   - Check sector rotation
   - Analyze peer stocks

5. Review technical levels (20 min)
   - Chart price action
   - Check volume
   - Identify support/resistance

6. Check insider activity (15 min)
   - SEC Form 4 filings
   - Recent transactions
   - Cluster analysis

Total time: 2.5-4 hours per stock
Result: Can only analyze 1-2 stocks per day
Problem: Miss opportunities in other 98 Nasdaq stocks
```

**Real trading example - missed opportunity:**

```
Day 1: Spent 4 hours analyzing AAPL (bullish, bought)
Day 1: NVDA had massive unusual call buying (missed this)
Day 2-10: AAPL up 2%, NVDA up 15%
Cost of missed opportunity: 13% gains

Reason: Can't manually track 100 stocks
```

---

## Template Solution

### What It Does

This template implements **multi-factor Nasdaq stock analysis with expert weighting**:

1. **Macro Regime** (35%) - Fed policy, VIX, yields → RISK_ON/RISK_OFF
2. **Options Flow** (25%) - Unusual activity, dark pools → Bullish/Bearish
3. **Earnings Momentum** (20%) - Estimate revisions → Positive/Negative
4. **Relative Strength** (10%) - vs QQQ/sector → Strong/Weak
5. **Technical Confirmation** (5%) - Volume, levels → Confirmed/Not
6. **Insider Activity** (5%) - Form 4 filings → Buying/Selling
7. **Aggregate** - Weighted score → BUY/SELL + position size

### Real-World Tool Usage

```yaml
name: nasdaq100_trade_signal

# Required MCP servers and APIs
tools:
  # Macro analysis (FREE)
  - fred-api              # Federal Reserve data
  - treasury-api          # Yield curve
  - vix-api               # Volatility index
  
  # Options flow (PAID - $50/month)
  - unusual-whales        # Unusual options activity
  - cboe-api              # Put/call ratios
  
  # Earnings (PAID - $50/month OR FREE limited)
  - estimize-api          # Crowdsourced estimates
  - yahoo-finance         # Basic consensus (free)
  
  # Market data (FREE)
  - yahoo-finance         # Prices, volume
  - alpha-vantage         # Technical indicators (free tier)
  
  # SEC filings (FREE)
  - sec-edgar             # Form 4, 13F

steps:
  # Stage 1: Macro environment analysis (35% weight)
  - name: analyze_macro
    template: macro_regime_analysis
    tools: [fred-api, treasury-api, vix-api]
    prompt: |
      Analyze macro environment for tech stocks:
      
      **Fed Policy:**
      - Current Fed Funds rate: {{fred.get_series('FEDFUNDS')}}
      - Fed Funds futures (6 months): {{cme.fedwatch.probability}}
      - Trend: Hiking, Pausing, or Cutting?
      
      **Inflation:**
      - Latest CPI: {{fred.get_series('CPIAUCSL')}}
      - Core PCE: {{fred.get_series('PCEPILFE')}}
      - Trend: Rising or Cooling?
      
      **Risk Environment:**
      - VIX level: {{vix.current}}
        VIX <15: Low fear (RISK_ON)
        VIX 15-25: Normal
        VIX >25: Elevated fear (RISK_OFF)
      
      **Treasury Yields:**
      - 10-year yield: {{treasury.get_yield('10Y')}}
      - 2-year yield: {{treasury.get_yield('2Y')}}
      - Curve: {{treasury.curve_shape}}
        Inverted: Recession risk
        Steepening: Growth expectations
      
      **Credit Spreads:**
      - Investment grade: {{fred.get_series('BAMLC0A0CM')}}
      - High yield: {{fred.get_series('BAMLH0A0HYM')}}
        Widening: Risk-off
        Tightening: Risk-on
      
      Return macro signal:
      ```json
      {
        "signal": "RISK_ON|NEUTRAL|RISK_OFF",
        "confidence": 0-100,
        "rationale": "Fed pausing, VIX 16, yields stable = favorable for tech",
        "regime_change": false,
        "weight": 0.35
      }
      ```
    output: macro_signal
  
  # Stage 2: Options flow analysis (25% weight)
  - name: analyze_options_flow
    template: options_flow_analysis
    tools: [unusual-whales, cboe-api]
    prompt: |
      Analyze options activity for {{input_data.ticker}}:
      
      **Unusual Options Activity (via Unusual Whales API):**
      - Query: Last 5 trading days
      - Filters: Premium > $100K, Sentiment: Bullish or Bearish
      
      {{unusual_whales.get_flow(ticker, days=5, min_premium=100000)}}
      
      Look for:
      - Large call sweeps (bullish - buying at ask)
      - Large put sweeps (bearish - buying at ask)
      - Dark pool prints (institutional activity)
      - Unusual volume vs open interest
      
      **Put/Call Analysis (via CBOE):**
      - Stock put/call ratio: {{cboe.get_pc_ratio(ticker)}}
      - Market-wide put/call: {{cboe.get_market_pc()}}
      
      Interpretation:
      - P/C > 1.5: Very bearish (contrarian bullish?)
      - P/C 0.7-1.5: Normal
      - P/C < 0.7: Very bullish (contrarian bearish?)
      
      **Gamma Exposure:**
      - Dealer gamma position: {{options.dealer_gamma}}
      - Positive gamma: Dealers stabilize price
      - Negative gamma: Dealers amplify moves
      
      Return options signal:
      ```json
      {
        "signal": "BULLISH|NEUTRAL|BEARISH",
        "confidence": 0-100,
        "evidence": [
          "3 large call sweeps $500K+ in past 2 days",
          "Dark pool print 500K shares at $195 (2% above current)"
        ],
        "put_call_ratio": 0.65,
        "smart_money_direction": "Accumulating calls",
        "weight": 0.25
      }
      ```
    output: options_signal
  
  # Stage 3: Earnings momentum analysis (20% weight)
  - name: analyze_earnings
    template: earnings_revision_analysis
    tools: [estimize-api, yahoo-finance]
    prompt: |
      Analyze earnings estimates for {{input_data.ticker}}:
      
      **Current Consensus (via Estimize):**
      - Current quarter EPS estimate: {{estimize.get_consensus(ticker, 'current_q')}}
      - Next quarter EPS estimate: {{estimize.get_consensus(ticker, 'next_q')}}
      - Full year EPS estimate: {{estimize.get_consensus(ticker, 'FY1')}}
      
      **Revision Trends (past 30/60/90 days):**
      {{estimize.get_revisions(ticker, days=[30, 60, 90])}}
      
      Positive momentum:
      - Upgrades > Downgrades
      - Estimates trending higher
      - Upside surprise expected
      
      Negative momentum:
      - Downgrades > Upgrades
      - Estimates trending lower
      - Downside surprise risk
      
      **Beat/Miss History:**
      - Last 4 quarters: {{yahoo.get_earnings_history(ticker, 4)}}
      - Pattern: Stock that beat 3+ quarters likely beats again
      
      **Guidance Expectations:**
      - Based on management commentary and industry trends
      - Will they raise, maintain, or lower guidance?
      
      Return earnings signal:
      ```json
      {
        "signal": "POSITIVE|NEUTRAL|NEGATIVE",
        "confidence": 0-100,
        "revision_trend": "15% upward revisions past 30 days",
        "surprise_probability": "High probability of beat",
        "guidance_expectation": "Likely to raise guidance",
        "weight": 0.20
      }
      ```
    output: earnings_signal
  
  # Stage 4: Relative strength analysis (10% weight)
  - name: analyze_relative_strength
    template: relative_strength_analysis
    tools: [yahoo-finance]
    prompt: |
      Calculate relative performance for {{input_data.ticker}}:
      
      **vs Nasdaq-100 (QQQ):**
      - Stock price: {{yahoo.get_price(ticker)}}
      - QQQ price: {{yahoo.get_price('QQQ')}}
      - Relative strength ratio: Stock/QQQ
      - 20-day RS trend: {{rs_ratio_20d_change}}
      
      Interpretation:
      - RS rising: Outperforming (bullish)
      - RS falling: Underperforming (bearish)
      
      **vs Sector (XLK for tech):**
      - Sector performance: {{yahoo.get_price('XLK')}}
      - Relative to sector: {{ticker_vs_sector}}
      
      **vs Peers:**
      - Similar companies performance
      - Leading or lagging peer group?
      
      **Sector Rotation:**
      - Is money flowing into or out of tech?
      - XLK vs SPY: {{sector_rotation}}
      
      Return relative strength signal:
      ```json
      {
        "signal": "OUTPERFORMING|NEUTRAL|UNDERPERFORMING",
        "confidence": 0-100,
        "vs_qqq": "Outperforming by 3% past 20 days",
        "vs_sector": "In line with XLK",
        "sector_flow": "Money flowing into tech",
        "weight": 0.10
      }
      ```
    output: relative_signal
  
  # Stage 5: Technical confirmation (5% weight)
  - name: analyze_technicals
    template: technical_confirmation
    tools: [alpha-vantage, yahoo-finance]
    prompt: |
      Technical analysis for {{input_data.ticker}}:
      
      **Moving Averages (trend):**
      - Current price: {{yahoo.get_price(ticker)}}
      - 50-day MA: {{alpha_vantage.get_sma(ticker, 50)}}
      - 200-day MA: {{alpha_vantage.get_sma(ticker, 200)}}
      
      Trend:
      - Price > 50MA > 200MA: Strong uptrend
      - Price < 50MA < 200MA: Strong downtrend
      
      **Volume Confirmation:**
      - Today's volume: {{yahoo.get_volume(ticker)}}
      - 20-day avg volume: {{yahoo.get_avg_volume(ticker, 20)}}
      
      Healthy:
      - Rising on up days (buyers active)
      - Falling on down days (sellers passive)
      
      **Key Levels:**
      - Recent high: {{recent_high}}
      - Recent low: {{recent_low}}
      - Support levels: {{support_levels}}
      - Resistance levels: {{resistance_levels}}
      
      Return technical signal:
      ```json
      {
        "signal": "CONFIRMED|NEUTRAL|NOT_CONFIRMED",
        "confidence": 0-100,
        "trend": "Uptrend - price above both MAs",
        "volume": "Confirming - high on up days",
        "key_levels": "Breaking above $200 resistance",
        "weight": 0.05
      }
      ```
    output: technical_signal
  
  # Stage 6: Insider activity (5% weight)
  - name: check_insider_activity
    template: insider_activity_analysis
    tools: [sec-edgar]
    prompt: |
      Check insider transactions for {{input_data.ticker}}:
      
      **Recent Form 4 Filings (via SEC EDGAR):**
      {{sec_edgar.get_form4(ticker, days=90)}}
      
      **Bullish Signals:**
      - CEO/CFO buying (high conviction)
      - Multiple insiders buying (cluster)
      - Large dollar amounts
      - Market purchases (not just option exercises)
      
      **Bearish Signals:**
      - Unusual selling by executives
      - Large blocks sold
      - Multiple insiders selling
      
      **Neutral:**
      - 10b5-1 plans (pre-scheduled)
      - Small amounts
      - Option exercises for tax planning
      
      Return insider signal:
      ```json
      {
        "signal": "BULLISH|NEUTRAL|BEARISH",
        "confidence": 0-100,
        "activity": "CEO bought $5M shares past 30 days",
        "cluster_buying": true,
        "conviction": "High - large dollar amount",
        "weight": 0.05
      }
      ```
    output: insider_signal
  
  # Stage 7: Calculate position size
  - name: calculate_position
    tools: [alpha-vantage]
    prompt: |
      Calculate volatility-adjusted position size:
      
      **Stock Data:**
      - Ticker: {{input_data.ticker}}
      - Current price: {{current_price}}
      - ATR (14-day): {{alpha_vantage.get_atr(ticker, 14)}}
      
      **Portfolio Parameters:**
      - Portfolio value: {{input_data.portfolio_value}}
      - Risk per trade: {{input_data.risk_per_trade | default: 0.01}} (1%)
      - Max position: {{input_data.max_position | default: 0.10}} (10%)
      
      **Position Sizing:**
      
      Method 1: ATR-based stop
      - Stop distance: 3 × ATR = {{atr * 3}}
      - Risk amount: Portfolio × Risk% = {{portfolio * risk_pct}}
      - Shares: Risk / Stop = {{risk_amount / stop_distance}}
      
      Method 2: Fixed percentage stop (-12%)
      - Stop price: Current × 0.88 = {{current * 0.88}}
      - Risk per share: Current - Stop = {{current - stop}}
      - Shares: Risk / Risk_per_share = {{risk_amount / risk_per_share}}
      
      Use: Smaller of the two methods
      
      **Volatility Adjustment:**
      - ATR%: {{atr / current_price}}
      - If ATR% > 3%: High vol, reduce position 50%
      - If ATR% < 1%: Low vol, can increase 25%
      
      Return position sizing:
      ```json
      {
        "shares": 42,
        "dollar_amount": 8190,
        "percent_of_portfolio": 8.2,
        "entry_price": 195.00,
        "stop_loss": 171.60,
        "take_profit": 200.85,
        "risk_amount": 983,
        "risk_percent": 0.98
      }
      ```
    output: position_sizing
  
  # Stage 8: Aggregate all signals
  - name: generate_trade_recommendation
    prompt: |
      Aggregate weighted signals and generate trade recommendation:
      
      **All Signals:**
      
      1. Macro (35%): {{macro_signal}}
      2. Options Flow (25%): {{options_signal}}
      3. Earnings (20%): {{earnings_signal}}
      4. Relative Strength (10%): {{relative_signal}}
      5. Technicals (5%): {{technical_signal}}
      6. Insiders (5%): {{insider_signal}}
      
      **Weighted Score Calculation:**
      
      Score = (Macro × 0.35) + (Options × 0.25) + (Earnings × 0.20) + 
              (Relative × 0.10) + (Technical × 0.05) + (Insider × 0.05)
      
      Where each signal contributes:
      - STRONG BULLISH: +10
      - BULLISH: +5
      - NEUTRAL: 0
      - BEARISH: -5
      - STRONG BEARISH: -10
      
      **Final Score:**
      - Score > 6: STRONG BUY
      - Score 3-6: BUY
      - Score -3 to 3: HOLD
      - Score -6 to -3: SELL
      - Score < -6: STRONG SELL
      
      **Confidence Level:**
      - All signals agree: High confidence (90%+)
      - Most signals agree: Medium confidence (70-90%)
      - Mixed signals: Low confidence (<70%)
      
      **Risk Factors:**
      - Macro risk: {{macro_signal.risks}}
      - Earnings risk: {{earnings_signal.risks}}
      - Technical risk: {{technical_signal.risks}}
      
      Generate final recommendation:
      ```json
      {
        "ticker": "AAPL",
        "recommendation": "BUY",
        "confidence": 78,
        "weighted_score": 5.2,
        
        "entry": {
          "price": 195.00,
          "timing": "At market or limit $193-195"
        },
        
        "targets": {
          "take_profit": 200.85,
          "profit_pct": 3.0,
          "stop_loss": 171.60,
          "loss_pct": -12.0
        },
        
        "position": {
          "shares": 42,
          "dollar_amount": 8190,
          "percent_portfolio": 8.2
        },
        
        "rationale": {
          "pros": [
            "Fed pausing supports tech valuations (Macro: RISK_ON)",
            "Unusual call buying $500K+ detected (Options: Bullish)",
            "Earnings estimates revised up 8% past 30 days (Earnings: Positive)",
            "Outperforming QQQ by 3% (Relative: Strong)"
          ],
          "cons": [
            "Approaching $200 resistance (Technical risk)",
            "VIX at 16 (could spike on negative news)"
          ]
        },
        
        "risk_management": {
          "max_loss": 983,
          "risk_reward_ratio": "1:0.25 (not great, but macro favorable)",
          "hold_period": "2-4 weeks or until stop/target hit",
          "re_evaluate_if": "Macro regime shifts to RISK_OFF"
        },
        
        "next_events": {
          "earnings": "Jan 30 (35 days)",
          "fomc_meeting": "Jan 31",
          "ex_dividend": "Jan 15"
        }
      }
      ```
    output: trade_recommendation
```

---

## Usage Example

### Analyzing AAPL

**Input:**
```json
{
  "ticker": "AAPL",
  "portfolio_value": 100000,
  "risk_per_trade": 0.01,
  "max_position": 0.10
}
```

**Execution:**
```bash
mcp-cli --template nasdaq100_trade_signal --input-data @aapl.json
```

**What Happens:**

```
[09:00:00] Starting nasdaq100_trade_signal for AAPL
[09:00:00] Step: analyze_macro
[09:00:15] ✓ Macro: RISK_ON (Confidence: 82%)
  - Fed paused rate hikes
  - VIX: 16 (low fear)
  - 10Y yield: 3.8% (stable)
  - Credit spreads: Tightening

[09:00:15] Step: analyze_options_flow
[09:00:30] ✓ Options: BULLISH (Confidence: 75%)
  - 3 unusual call sweeps >$500K past 2 days
  - Dark pool: 500K shares at $195 (+2%)
  - Put/Call: 0.62 (bullish)

[09:00:30] Step: analyze_earnings
[09:00:45] ✓ Earnings: POSITIVE (Confidence: 70%)
  - Estimates up 8% past 30 days
  - 12 upgrades vs 2 downgrades
  - Beat last 4 quarters

[09:00:45] Step: analyze_relative_strength
[09:00:55] ✓ Relative: OUTPERFORMING (Confidence: 65%)
  - +3% vs QQQ past 20 days
  - In line with XLK sector
  - Money flowing into tech

[09:00:55] Step: analyze_technicals
[09:01:05] ✓ Technical: CONFIRMED (Confidence: 60%)
  - Price above 50/200 MA
  - Volume confirming
  - Breaking $200 resistance

[09:01:05] Step: check_insider_activity
[09:01:15] ✓ Insider: BULLISH (Confidence: 85%)
  - CEO bought $5M shares
  - 3 directors bought past 30 days
  - No unusual selling

[09:01:15] Step: calculate_position
[09:01:20] ✓ Position: 42 shares ($8,190)
  - ATR-based sizing
  - 8.2% of portfolio
  - Risk: $983 (0.98%)

[09:01:20] Step: generate_trade_recommendation
[09:01:30] ✓ Recommendation generated

[09:01:30] ✓ Template completed (90 seconds)
```

**Output:**

```markdown
# Trade Signal: AAPL

**Recommendation:** BUY  
**Confidence:** 78%  
**Generated:** 2024-12-28 09:01:30

---

## Trade Details

**Entry:** $195.00  
**Take Profit:** $200.85 (+3%)  
**Stop Loss:** $171.60 (-12%)

**Position:**
- Shares: 42
- Dollar amount: $8,190
- Portfolio %: 8.2%
- Max risk: $983 (0.98% of portfolio)

---

## Signal Breakdown

### Macro Environment: RISK_ON (35% weight)
**Confidence:** 82%

Federal Reserve paused rate hikes, creating favorable environment for tech stocks.

- Fed Funds: 5.50% (paused)
- Inflation cooling: CPI 3.1% (down from 3.7%)
- VIX: 16 (low fear, RISK_ON)
- 10Y yield: 3.8% (stable)
- Credit spreads: Tightening (risk-on)

**Impact:** Positive for tech valuations

### Options Flow: BULLISH (25% weight)
**Confidence:** 75%

Smart money accumulating calls, dark pool activity above current price.

- 3 unusual call sweeps >$500K in past 2 days
- Dark pool: 500K shares at $195 (+2% premium)
- Put/Call ratio: 0.62 (bullish)
- Dealer gamma: Positive (price stabilization)

**Impact:** Institutional accumulation detected

### Earnings Momentum: POSITIVE (20% weight)
**Confidence:** 70%

Analysts raising estimates, beat pattern continues.

- Estimates up 8% past 30 days
- Revisions: 12 upgrades, 2 downgrades
- Beat history: 4/4 last quarters
- Guidance expectation: Likely raise

**Impact:** Positive earnings surprise likely

### Relative Strength: OUTPERFORMING (10% weight)
**Confidence:** 65%

Outperforming Nasdaq and in line with tech sector.

- vs QQQ: +3% past 20 days
- vs XLK: In line
- Sector rotation: Money into tech

**Impact:** Leading the market

### Technical: CONFIRMED (5% weight)
**Confidence:** 60%

Uptrend confirmed, breaking resistance on volume.

- Trend: Above 50MA ($187) and 200MA ($178)
- Volume: Above average on up days
- Key level: Breaking $200 resistance

**Impact:** Technical breakout

### Insider Activity: BULLISH (5% weight)
**Confidence:** 85%

CEO and directors buying shares.

- CEO: Bought $5M shares (high conviction)
- Directors: 3 bought past 30 days
- Pattern: Cluster buying (bullish)

**Impact:** Management confidence

---

## Rationale

**Why BUY:**
1. Macro environment favorable (Fed pause = tech rally)
2. Smart money accumulating (options flow bullish)
3. Earnings momentum positive (estimates rising)
4. Technical breakout above $200

**Risk Factors:**
1. Approaching resistance (could pause at $200)
2. Earnings in 35 days (volatility risk)
3. FOMC meeting Jan 31 (potential surprise)

**Risk/Reward:** 1:0.25 (not ideal, but macro favorable)

---

## Trade Management

**Entry Strategy:**
- Limit order: $193-195
- Or market if under $196

**Exit Strategy:**
- Take profit: $200.85 (+3%)
- Stop loss: $171.60 (-12%)
- Trail stop if momentum continues

**Re-evaluate If:**
- Macro shifts to RISK_OFF
- Unusual selling appears in options
- Earnings estimates reversed

**Hold Period:** 2-4 weeks

---

## Upcoming Events

- **Ex-dividend:** Jan 15 ($0.24)
- **FOMC Meeting:** Jan 31
- **Earnings:** Jan 30

---

**Analysis Time:** 90 seconds  
**Data Sources:** FRED, Unusual Whales, Estimize, Yahoo Finance, SEC EDGAR  
**Template:** nasdaq100_trade_signal v1.0
```

---

## When to Use

### ✅ Appropriate Use Cases

**Nasdaq-100 stocks:**
- AAPL, MSFT, NVDA, GOOGL, AMZN, META, TSLA, etc.
- Tech-focused companies
- High liquidity

**Multi-day trades:**
- Hold period: 2-14 days
- Not day trading (need different approach)
- Not buy-and-hold (this is tactical)

**Systematic analysis:**
- Want data-driven decisions
- Eliminate emotion
- Consistent methodology

### ❌ Inappropriate Use Cases

**Penny stocks:**
- Low liquidity
- Options flow not meaningful
- Different dynamics

**Day trading:**
- Need tick-by-tick data
- Different timeframe
- This template too slow

**Long-term investing:**
- Macro shifts over years
- This is for tactical 2-4 week trades
- Use fundamental analysis instead

---

## Trade-offs

### Advantages

**Multi-source synthesis:**
- Macro + options + earnings + technical + insider
- No human can track all manually
- **Coverage: 1 stock/day → 20 stocks/day**

**Expert weighting:**
- Macro properly weighted (35%)
- Options flow included (what pros watch)
- Not just technical analysis

**Risk management:**
- Volatility-adjusted position sizing
- Clear stop loss
- Max 1% risk per trade

### Limitations

**Data costs:**
- Free tier: Limited (no options flow)
- Recommended: $100/month
- Missing options flow = missing 25% of signal

**Lagging indicators:**
- Insider trades: Delayed by days
- 13F filings: Delayed by quarters
- Not all signals are real-time

**Can't predict:**
- Black swans
- Earnings shocks
- Fed surprises

---

## Best Practices

**Before Trading:**

**✅ Do:**
- Check macro regime first (don't fight the Fed)
- Wait for options flow confirmation
- Size positions by volatility
- Set stop losses BEFORE entering
- Re-evaluate if macro shifts

**❌ Don't:**
- Trade against macro (35% weight for reason)
- Ignore options flow (smart money knows first)
- Use fixed position sizes (volatility matters)
- Skip stop losses ("I'll watch it")
- Hold through earnings without plan

**After Entry:**

**✅ Do:**
- Honor stop losses (cut losses)
- Take profits at target (don't get greedy)
- Re-analyze if new data emerges
- Track win/loss ratio

**❌ Don't:**
- Move stop losses lower (that's how you blow up)
- Hold losing positions hoping for recovery
- Let winners turn into losers
- Revenge trade after loss

---

## Related Resources

- **[Template File](../templates/nasdaq100_trade_signal.yaml)** - Download complete template
- **[Macro Regime Analysis](../templates/macro_regime_analysis.yaml)** - Fed policy sub-template
- **[Options Flow Analysis](../templates/options_flow_analysis.yaml)** - Smart money tracking

---

**Nasdaq-100 trade signals: Systematic multi-factor analysis with proper risk management.**

Remember: Even with 78% confidence, 22% chance of loss. Always use stops.
