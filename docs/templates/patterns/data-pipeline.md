# Data Pipeline Pattern

Build robust ETL (Extract-Transform-Load) workflows with AI.

---

## Overview

The **Data Pipeline Pattern** structures data processing workflows using ETL (Extract-Transform-Load).

**What is ETL?** (For beginners)
- **Extract:** Get data from somewhere (database, file, API, logs)
- **Transform:** Clean it up, fix errors, change format, add calculations
- **Load:** Put it somewhere else (database, file, dashboard, report)

**Why use this pattern:**
- Reproducible: Same process every time
- Auditable: Track what changed and when
- Reliable: Built-in validation and error handling
- Maintainable: Clear stages, easy to debug

**Use when:**
- Moving data between systems (CRM → warehouse)
- Cleaning messy data (fix formats, remove duplicates)
- Processing logs or events (parse → classify → alert)
- Scheduled data jobs (nightly ETL, hourly updates)
- Data quality is critical (healthcare, finance, compliance)

**Real-world examples:**
- Pull customer data from Salesforce → clean phone numbers → load to BigQuery
- Parse application logs → find errors → send alerts
- Extract sales data → calculate metrics → generate dashboard

---

## Pattern Structure

```
Extract → Validate → Transform → Enrich → Load
```

**What happens in each stage:**
1. **Extract:** Get raw data from source
2. **Validate:** Check if data is good (required fields, valid formats)
3. **Transform:** Clean and modify data (fix dates, uppercase names, calculate totals)
4. **Enrich:** Add extra information (lookup customer details, calculate metrics)
5. **Load:** Save final data to destination

### Basic ETL Pipeline

**What it does:** Complete extract-transform-load workflow with validation at each stage.

**Use when:** You need to move data from one place to another with quality checks.

```yaml
name: basic_etl
description: Extract, transform, and load data with validation
version: 1.0.0

config:
  defaults:
    provider: anthropic
    model: claude-sonnet-4

steps:
  # EXTRACT: Get data from source
  - name: extract_data
    servers: [database, filesystem]
    prompt: |
      Extract data from source:
      {{input_data.source}}
      
      Return as structured JSON.
    output: raw_data
    error_handling:
      on_failure: stop
      default_output: "EXTRACTION_FAILED"
  
  # VALIDATE: Check if data is good
  - name: validate_data
    prompt: |
      Validate extracted data:
      {{raw_data}}
      
      Check:
      - Required fields present
      - Data types correct
      - Values in valid ranges
      - No duplicates
      
      Return validation report with status (VALID/INVALID) and issues.
    output: validation_result
  
  # TRANSFORM: Clean and modify (only if valid)
  - name: transform_data
    condition: "{{validation_result}} contains 'VALID'"
    prompt: |
      Transform data:
      {{raw_data}}
      
      Apply transformations:
      {{input_data.transformation_rules}}
      
      Return transformed data in {{input_data.output_format}} format.
    output: transformed_data
  
  # ENRICH: Add extra information
  - name: enrich_data
    servers: [api-service]
    prompt: |
      Enrich transformed data:
      {{transformed_data}}
      
      Add:
      - Calculated fields
      - Lookup data from external sources
      - Derived metrics
    output: enriched_data
  
  # LOAD: Save to destination
  - name: load_data
    servers: [database]
    prompt: |
      Load data to destination:
      
      Data: {{enriched_data}}
      Destination: {{input_data.destination}}
      
      Return load status and record count.
    output: load_result
    error_handling:
      on_failure: retry
      max_retries: 3
      retry_backoff: exponential
```

**Usage:**
```bash
# Example 1: Customer data pipeline
mcp-cli --template basic_etl --input-data '{
  "source": "customers_raw.csv",
  "transformation_rules": "uppercase names, format phone numbers as +1-XXX-XXX-XXXX",
  "output_format": "JSON",
  "destination": "customers_clean.json"
}'

# Example 2: Sales data ETL
mcp-cli --template basic_etl --input-data '{
  "source": "SELECT * FROM sales WHERE date > 2024-01-01",
  "transformation_rules": "calculate totals, convert currency to USD",
  "output_format": "CSV",
  "destination": "data_warehouse.sales_cleaned"
}'
```

**What happens:**
1. Extract: Reads from `customers_raw.csv` → stores as `{{raw_data}}`
2. Validate: Checks data quality → if bad, reports issues and stops
3. Transform: Uppercases names, formats phones → stores as `{{transformed_data}}`
4. Enrich: Adds calculated fields → stores as `{{enriched_data}}`
5. Load: Saves to `customers_clean.json` → returns success/failure

**Performance:**
- Processes ~1,000 records per minute
- Cost: ~$0.05 per 1,000 records (AI transformation calls)

**When validation fails:**
Pipeline stops at step 2, returns validation report with issues. Fix data and re-run.

---

## Pattern: Parallel Data Processing

**What it does:** Extracts from multiple sources simultaneously, transforms each, then merges results.

**Use when:**
- Have multiple independent data sources (database + API + files)
- Sources don't depend on each other
- Speed matters (want 3x faster processing)

**Don't use when:**
- Later sources need data from earlier ones
- Sources are slow (won't gain much from parallelization)
- Single source only

```yaml
name: parallel_etl
description: Extract from multiple sources in parallel
version: 1.0.0

steps:
  # Extract from multiple sources simultaneously
  - name: extract_all_sources
    parallel:
      - name: extract_database
        servers: [database]
        prompt: "Extract from DB: {{input_data.db_query}}"
        output: db_data
      
      - name: extract_api
        servers: [api-service]
        prompt: "Extract from API: {{input_data.api_endpoint}}"
        output: api_data
      
      - name: extract_files
        servers: [filesystem]
        prompt: "Extract from files: {{input_data.file_pattern}}"
        output: file_data
    max_concurrent: 3
    aggregate: merge
    output: all_raw_data
  
  # Transform each source independently (also parallel)
  - name: transform_all
    parallel:
      - name: transform_db
        prompt: "Transform DB data: {{all_raw_data.db_data}}"
        output: transformed_db
      
      - name: transform_api
        prompt: "Transform API data: {{all_raw_data.api_data}}"
        output: transformed_api
      
      - name: transform_files
        prompt: "Transform file data: {{all_raw_data.file_data}}"
        output: transformed_files
    max_concurrent: 3
    aggregate: merge
    output: all_transformed
  
  # Merge transformed data from all sources
  - name: merge_data
    prompt: |
      Merge data from all sources:
      {{all_transformed}}
      
      Handle:
      - Duplicate records (keep most recent)
      - Conflicting values (use priority: DB > API > files)
      - Missing data (mark with NULL)
      
      Return unified dataset.
    output: merged_data
  
  # Load merged data to destination
  - name: load
    servers: [database]
    prompt: |
      Load merged data:
      {{merged_data}}
      
      To: {{input_data.destination}}
```

**Usage:**
```bash
mcp-cli --template parallel_etl --input-data '{
  "db_query": "SELECT * FROM orders WHERE date >= 2024-01-01",
  "api_endpoint": "https://api.example.com/customers",
  "file_pattern": "/data/logs/*.csv",
  "destination": "data_warehouse.merged_orders"
}'
```

**What happens:**
1. All 3 extractions start at the same time:
   - Database query runs
   - API call happens
   - File reads occur
2. Each completes independently, stores result
3. All 3 transformations run in parallel
4. Merge step combines all transformed data
5. Load saves merged data to warehouse

**Performance comparison:**
- **Sequential:** 30s (DB) + 20s (API) + 10s (files) = **60 seconds**
- **Parallel:** max(30s, 20s, 10s) = **30 seconds**
- **Speedup: 2x faster**

**Cost:** Same API calls either way (3 extracts + 3 transforms + 1 merge + 1 load)

**Why this works:**
- Sources are independent (can run simultaneously)
- Network/IO time parallelized
- CPU efficiently used

**When it doesn't help:**
- If one source takes 90% of time (bottleneck)
- If sources need data from each other (dependencies)

---

## Pattern: Data Quality Pipeline

**What it does:** Profiles data to find issues, creates cleaning rules, then cleans the data.

**Use when:**
- Data quality is critical (healthcare, finance, compliance)
- Messy incoming data (inconsistent formats, missing values)
- Need auditable cleaning process
- Want automated quality reports

```yaml
name: data_quality_pipeline
description: Profile, detect issues, clean, validate
version: 1.0.0

steps:
  # Step 1: Profile the data to understand it
  - name: profile_data
    prompt: |
      Profile this dataset:
      {{input_data.raw_data}}
      
      Analyze:
      - Field types and formats (dates, numbers, text)
      - Value distributions (min/max, common values)
      - Missing values (how many? which fields?)
      - Outliers (values outside normal range)
      - Overall data quality score (0-100)
      
      Return JSON profile report.
    output: data_profile
  
  # Step 2: Detect specific issues
  - name: detect_issues
    prompt: |
      Based on this profile:
      {{data_profile}}
      
      Identify data quality issues:
      - Invalid values (negative ages, future dates)
      - Inconsistent formats (phone: +1-555-1234 vs 5551234)
      - Missing required fields (NULL in email field)
      - Outliers (age: 150, seems wrong)
      - Duplicates (same record appears 3 times)
      
      Categorize each issue:
      - critical: Blocks processing
      - high: Affects accuracy
      - medium: Affects completeness
      - low: Minor formatting
      
      Return JSON issue report.
    output: quality_issues
  
  # Step 3: Generate cleaning rules automatically
  - name: create_cleaning_rules
    prompt: |
      Generate cleaning rules for these issues:
      {{quality_issues}}
      
      For each issue, create rule:
      - Rule description (e.g., "Format all phone numbers as +1-XXX-XXX-XXXX")
      - Transformation logic (e.g., "Extract digits, add country code")
      - Validation criteria (e.g., "Must be 11 digits")
      
      Return JSON rules.
    output: cleaning_rules
  
  # Step 4: Apply cleaning rules
  - name: clean_data
    prompt: |
      Apply these cleaning rules:
      {{cleaning_rules}}
      
      To this data:
      {{input_data.raw_data}}
      
      Return:
      - Cleaned dataset
      - Cleaning report (what changed, how many records affected)
    output: cleaned_data
  
  # Step 5: Validate cleaned data
  - name: validate_cleaned
    prompt: |
      Validate cleaned data meets quality standards:
      {{cleaned_data}}
      
      Check:
      - All critical issues resolved
      - Data integrity maintained (no data loss)
      - Quality score improved (compare to original)
      
      Return: PASS or FAIL with remaining issues.
    output: validation_report
```

**Usage:**
```bash
mcp-cli --template data_quality_pipeline --input-data '{
  "raw_data": [
    {"name": "john doe", "phone": "5551234", "age": "25"},
    {"name": "JANE SMITH", "phone": "+1-555-5678", "age": null},
    {"name": "Bob Lee", "phone": "555-9012", "age": "150"}
  ]
}'
```

**What happens:**
1. Profile: Analyzes data → finds inconsistent phone formats, missing age, outlier age (150)
2. Detect: Categorizes issues → phone format (medium), missing age (high), outlier (high)
3. Rules: Creates rules → "Format phones as +1-XXX-XXX-XXXX", "Flag ages > 120"
4. Clean: Applies rules → all phones formatted, age 150 flagged for review
5. Validate: Checks quality → quality score improved from 60 → 85

**Output example:**
```json
{
  "cleaned_data": [
    {"name": "John Doe", "phone": "+1-555-1234", "age": "25"},
    {"name": "Jane Smith", "phone": "+1-555-5678", "age": "NULL_FLAGGED"},
    {"name": "Bob Lee", "phone": "+1-555-9012", "age": "150_REVIEW_NEEDED"}
  ],
  "report": {
    "records_processed": 3,
    "issues_found": 5,
    "issues_fixed": 3,
    "manual_review_needed": 2,
    "quality_score_before": 60,
    "quality_score_after": 85
  }
}
```

**Why use this pattern:**
- Automated quality checking saves time
- Auditable process (know what changed and why)
- Consistent rules applied every time
- Generates documentation automatically

---

## Pattern: Incremental Data Pipeline

Process only new/changed data.

```yaml
name: incremental_pipeline
steps:
  # Get last processed timestamp
  - name: get_last_run
    servers: [database]
    prompt: "Get last successful run timestamp"
    output: last_run_time
  
  # Extract only new data
  - name: extract_incremental
    servers: [database]
    prompt: |
      Extract records modified after:
      {{last_run_time}}
      
      From: {{source_table}}
    output: new_records
  
  # Check if any new data
  - name: check_new_data
    prompt: |
      Count records in: {{new_records}}
      Return: "HAS_DATA" or "NO_DATA"
    output: data_status
  
  # Process only if new data exists
  - name: process_data
    condition: "{{data_status}} contains 'HAS_DATA'"
    prompt: "Transform: {{new_records}}"
    output: processed_data
  
  # Upsert to destination
  - name: upsert_data
    condition: "{{data_status}} contains 'HAS_DATA'"
    servers: [database]
    prompt: |
      Upsert to destination:
      {{processed_data}}
      
      Strategy: update existing, insert new
    output: upsert_result
  
  # Update last run timestamp
  - name: update_timestamp
    condition: "{{upsert_result}} contains 'SUCCESS'"
    servers: [database]
    prompt: "Update last run timestamp to: {{current_time}}"
```

---

## Pattern: Data Aggregation Pipeline

Aggregate and summarize large datasets.

```yaml
name: aggregation_pipeline
steps:
  # Step 1: Group data
  - name: group_data
    transform:
      - operation: group
        key: "{{group_by_field}}"
    input: raw_data
    output: grouped_data
  
  # Step 2: Aggregate each group
  - name: aggregate_groups
    for_each: "{{grouped_data}}"
    item_name: group
    prompt: |
      Aggregate group {{group.key}}:
      
      Data: {{group.items}}
      
      Calculate:
      - Count
      - Sum
      - Average
      - Min/Max
      - Custom metrics: {{metrics}}
    output: aggregations
  
  # Step 3: Create summary
  - name: create_summary
    prompt: |
      Create summary report from aggregations:
      {{aggregations}}
      
      Include:
      - Top 10 by {{ranking_metric}}
      - Trends
      - Anomalies
      - Insights
```

---

## Real-World Examples

### Example 1: Log Processing Pipeline

**What it does:** Reads log files, parses them, filters errors, classifies them, generates report.

**Use case:** Process application logs to find and categorize errors for alerting.

```yaml
name: log_processing_pipeline
version: 1.0.0

config:
  defaults:
    provider: anthropic
    model: claude-sonnet-4

steps:
  # Step 1: Extract logs from files
  - name: extract_logs
    servers: [filesystem]
    prompt: |
      Read log files matching pattern:
      {{input_data.log_pattern}}
      
      Return array of log entries.
    output: raw_logs
  
  # Step 2: Parse each log entry
  - name: parse_logs
    for_each: "{{raw_logs}}"
    item_name: log_entry
    prompt: |
      Parse this log entry:
      {{log_entry}}
      
      Extract:
      - Timestamp (ISO format)
      - Level (INFO, WARN, ERROR, DEBUG)
      - Message
      - Metadata (any additional fields)
      
      Return as JSON.
    output: parsed_logs
  
  # Step 3: Filter to errors only
  - name: filter_errors
    transform:
      - operation: filter
        condition: "level == 'ERROR'"
    input: parsed_logs
    output: error_logs
  
  # Step 4: Classify each error
  - name: classify_errors
    for_each: "{{error_logs}}"
    item_name: error
    prompt: |
      Classify this error:
      {{error}}
      
      Determine:
      - Type: network, database, application, authentication
      - Severity: critical, high, medium, low
      - Root cause: Brief explanation
      - Action needed: What should be done
      
      Return as JSON.
    output: classified_errors
  
  # Step 5: Generate error report
  - name: error_report
    prompt: |
      Create error analysis report from:
      {{classified_errors}}
      
      Include:
      - Executive summary
      - Error summary by type (how many of each)
      - Top 5 errors by frequency
      - Critical issues (need immediate attention)
      - Recommendations for fixes
      
      Format as markdown.
```

**Usage:**
```bash
# Example 1: Process today's logs
mcp-cli --template log_processing_pipeline --input-data '{
  "log_pattern": "/var/log/app/*.log"
}'

# Example 2: Process specific date range
mcp-cli --template log_processing_pipeline --input-data '{
  "log_pattern": "/var/log/app/2024-12-*.log"
}'
```

**What happens:**
1. Extract: Reads all .log files from `/var/log/app/` → 5,000 log entries
2. Parse: Extracts timestamp, level, message from each → 5,000 parsed entries
3. Filter: Keeps only ERROR level → 127 error entries
4. Classify: Categorizes each error → 127 classified errors
5. Report: Creates summary → "87 database errors (critical), 40 network errors (medium)"

**Example output:**
```markdown
# Error Analysis Report
Generated: 2024-12-26

## Executive Summary
- Total errors: 127
- Critical: 87 (68%)
- Action needed: Database connection pool exhausted

## Errors by Type
- Database: 87 (68%)
- Network: 40 (31%)
- Application: 0 (0%)

## Top Errors
1. "Connection pool exhausted" - 85 occurrences - CRITICAL
2. "Timeout connecting to API" - 35 occurrences - MEDIUM
3. "Invalid token" - 7 occurrences - LOW

## Recommendations
1. Increase database connection pool size
2. Add retry logic for API calls
3. Review token expiration policy
```

**Performance:**
- 5,000 logs processed in ~3 minutes
- Cost: ~$0.15 (classification is most expensive step)

### Example 2: Sales Data Pipeline

**What it does:** Extracts sales data, cleans it, enriches with customer info, calculates metrics, creates dashboard.

**Use case:** Daily sales data ETL for business intelligence dashboard.

```yaml
name: sales_data_pipeline
version: 1.0.0

steps:
  # Step 1: Extract sales data from database
  - name: extract_sales
    servers: [database]
    prompt: |
      Extract sales for date range:
      Start: {{input_data.start_date}}
      End: {{input_data.end_date}}
      
      Return as JSON array.
    output: raw_sales
  
  # Step 2: Clean and validate sales data
  - name: clean_sales
    prompt: |
      Clean sales data:
      {{raw_sales}}
      
      Fix these issues:
      - Missing customer IDs (lookup by name/email)
      - Invalid amounts (negative, zero, or too large)
      - Duplicate transactions (keep most recent)
      - Incorrect dates (future dates, obviously wrong)
      
      Return cleaned sales data and cleaning report.
    output: cleaned_sales
  
  # Step 3: Enrich with customer information
  - name: enrich_with_customers
    servers: [database]
    prompt: |
      Join customer data to sales:
      Sales: {{cleaned_sales}}
      
      For each sale, add:
      - Customer name
      - Customer segment (Enterprise, SMB, Individual)
      - Region (North, South, East, West)
      - Account manager
      
      Return enriched sales data.
    output: enriched_sales
  
  # Step 4: Calculate business metrics
  - name: calculate_metrics
    prompt: |
      Calculate metrics for each sale:
      {{enriched_sales}}
      
      Add these calculated fields:
      - Profit margin: (revenue - cost) / revenue * 100
      - Discount percentage: discount / list_price * 100
      - Customer LTV contribution: sale_amount / customer_total_spend * 100
      
      Return sales with metrics.
    output: sales_with_metrics
  
  # Step 5: Aggregate by customer segment
  - name: aggregate_by_segment
    transform:
      - operation: group
        key: "customer_segment"
    input: sales_with_metrics
    output: segment_sales
  
  # Step 6: Create dashboard data
  - name: create_dashboard
    prompt: |
      Create dashboard data from:
      {{segment_sales}}
      
      Generate:
      1. Sales by segment (for bar chart):
         - Segment name
         - Total revenue
         - Number of transactions
      
      2. Top 10 products (for table):
         - Product name
         - Units sold
         - Revenue
      
      3. Daily trends (for line chart):
         - Date
         - Revenue
         - Transaction count
      
      4. KPIs (for metrics cards):
         - Total revenue
         - Average order value
         - Profit margin %
         - Top customer segment
      
      Return as JSON formatted for dashboard.
```

**Usage:**
```bash
# Daily ETL job
mcp-cli --template sales_data_pipeline --input-data '{
  "start_date": "2024-12-25",
  "end_date": "2024-12-26"
}'

# Weekly report
mcp-cli --template sales_data_pipeline --input-data '{
  "start_date": "2024-12-19",
  "end_date": "2024-12-26"
}'
```

**What happens:**
1. Extract: Pulls 1,247 sales records from database
2. Clean: Fixes 23 missing IDs, removes 5 duplicates → 1,219 valid sales
3. Enrich: Adds customer data to each sale
4. Calculate: Adds profit margin, discount %, LTV contribution
5. Aggregate: Groups by segment (Enterprise: 347, SMB: 621, Individual: 251)
6. Dashboard: Creates chart/table data for visualization

**Example dashboard output:**
```json
{
  "sales_by_segment": [
    {"segment": "Enterprise", "revenue": 2847392, "transactions": 347},
    {"segment": "SMB", "revenue": 1254871, "transactions": 621},
    {"segment": "Individual", "revenue": 123847, "transactions": 251}
  ],
  "kpis": {
    "total_revenue": 4226110,
    "avg_order_value": 3466,
    "profit_margin": 42.3,
    "top_segment": "Enterprise"
  }
}
```

**Performance:**
- Processes 1,000-2,000 sales records in ~5 minutes
- Cost: ~$0.30 per run (enrichment and calculations are expensive)
- Runs daily at 2 AM via cron job

### Example 3: API Data Integration

**What it does:** Fetches data from multiple APIs in parallel, normalizes formats, deduplicates, stores.

**Use case:** Integrate data from multiple SaaS tools (Salesforce + HubSpot) into central database.

```yaml
name: api_integration_pipeline
version: 1.0.0

steps:
  # Step 1: Fetch from multiple APIs in parallel
  - name: fetch_all_apis
    parallel:
      - name: fetch_api1
        servers: [api-service]
        prompt: |
          Fetch from API1:
          {{input_data.api1_endpoint}}
          
          Return as JSON array.
        output: api1_data
        error_handling:
          on_failure: continue
          default_output: []
      
      - name: fetch_api2
        servers: [api-service]
        prompt: |
          Fetch from API2:
          {{input_data.api2_endpoint}}
          
          Return as JSON array.
        output: api2_data
        error_handling:
          on_failure: continue
          default_output: []
    max_concurrent: 2
    aggregate: merge
    output: all_api_data
  
  # Step 2: Normalize data to common schema
  - name: normalize_data
    prompt: |
      Normalize to common schema:
      
      API1 data (Salesforce): {{all_api_data.api1_data}}
      API2 data (HubSpot): {{all_api_data.api2_data}}
      
      Target schema:
      {{input_data.target_schema}}
      
      Map fields:
      - API1.Id → id
      - API1.FullName → name
      - API2.contact_id → id
      - API2.full_name → name
      
      Return normalized data array.
    output: normalized_data
  
  # Step 3: Deduplicate records
  - name: deduplicate
    prompt: |
      Remove duplicates from:
      {{normalized_data}}
      
      Match criteria:
      {{input_data.dedup_fields}} (e.g., email, phone)
      
      Strategy: Keep most recent record (based on updated_at)
      
      Return:
      - Deduplicated data
      - Duplicate report (how many removed, which fields matched)
    output: deduplicated_data
  
  # Step 4: Store results to database
  - name: store_data
    servers: [database]
    prompt: |
      Store to database:
      {{deduplicated_data}}
      
      Table: {{input_data.destination_table}}
      Mode: upsert (update if exists, insert if new)
      
      Return:
      - Records inserted
      - Records updated
      - Any errors
```

**Usage:**
```bash
mcp-cli --template api_integration_pipeline --input-data '{
  "api1_endpoint": "https://salesforce.com/api/contacts",
  "api2_endpoint": "https://api.hubspot.com/contacts/v1/lists/all/contacts",
  "target_schema": {
    "id": "string",
    "name": "string",
    "email": "string",
    "phone": "string",
    "source": "string"
  },
  "dedup_fields": ["email", "phone"],
  "destination_table": "crm.contacts"
}'
```

**What happens:**
1. Fetch: Pulls from Salesforce (342 contacts) and HubSpot (289 contacts) simultaneously
2. Normalize: Converts both to common schema → 631 total records
3. Deduplicate: Finds 47 duplicates (same email) → keeps most recent → 584 unique
4. Store: Upserts to database → 284 inserted, 300 updated

**Example dedup report:**
```json
{
  "total_input": 631,
  "duplicates_found": 47,
  "duplicates_removed": 47,
  "output_records": 584,
  "match_breakdown": {
    "email_match": 42,
    "phone_match": 3,
    "email_and_phone_match": 2
  }
}
```

**Performance:**
- Parallel API fetching saves ~10 seconds vs sequential
- Processes ~600 records in ~2 minutes
- Cost: ~$0.08 per run

**Why parallel fetching:**
- API1 takes 15s, API2 takes 12s
- Sequential: 15s + 12s = 27s
- Parallel: max(15s, 12s) = 15s
- **Saves 44% time**

**Error handling:**
- If API1 fails: continues with API2 data only
- If API2 fails: continues with API1 data only
- If both fail: pipeline stops with error

---

## Best Practices

### 1. Always Validate

```yaml
# Good: Validate at each stage
- name: extract
  output: raw_data
- name: validate_extraction
  prompt: "Validate: {{raw_data}}"
- name: transform
  condition: "valid"

# Bad: No validation
- name: extract
- name: transform  # Might process bad data
```

### 2. Handle Errors Gracefully

```yaml
# Good: Robust error handling
- name: extract
  error_handling:
    on_failure: retry
    max_retries: 3
    default_output: []

# Bad: No error handling
- name: extract  # Fails on any error
```

### 3. Make Idempotent

```yaml
# Good: Idempotent pipeline
- name: check_if_processed
  prompt: "Check if {{record_id}} already processed"
- name: process
  condition: "not processed"

# Bad: Reprocesses everything
- name: process  # Duplicate data on re-runs
```

### 4. Log Everything

```yaml
# Good: Comprehensive logging
- name: extract
  output: data
- name: log_extract
  prompt: |
    Log extraction:
    Records: {{data | length}}
    Timestamp: {{now}}
    Status: SUCCESS
```

### 5. Use Incremental Processing

```yaml
# Good: Process only new data
- name: get_last_processed
- name: extract_since
  prompt: "Extract since: {{last_processed}}"

# Bad: Full reload every time
- name: extract_all  # Slow and wasteful
```

---

## Performance Optimization

### Parallel Processing

```yaml
# Fast: Process in parallel
parallel:
  - name: transform_batch_1
  - name: transform_batch_2
  - name: transform_batch_3
max_concurrent: 3
```

### Batching

```yaml
# Process in batches
- for_each: "{{data}}"
  batch_size: 100  # Process 100 at a time
  prompt: "Transform batch: {{batch}}"
```

### Caching

```yaml
# Cache expensive lookups
- name: load_reference_data
  output: reference_cache  # Reuse across steps

- name: enrich_1
  prompt: "Lookup in: {{reference_cache}}"
- name: enrich_2
  prompt: "Lookup in: {{reference_cache}}"
```

---

## Monitoring and Observability

```yaml
steps:
  # Extract with metrics
  - name: extract
    prompt: "Extract from: {{source}}"
    output: raw_data
  
  # Log metrics
  - name: log_extraction_metrics
    prompt: |
      Log metrics:
      - Records extracted: {{raw_data | length}}
      - Duration: {{duration}}
      - Source: {{source}}
      - Timestamp: {{now}}
  
  # Transform with metrics
  - name: transform
    prompt: "Transform: {{raw_data}}"
    output: transformed
  
  # Log transform metrics
  - name: log_transform_metrics
    prompt: |
      Log metrics:
      - Records transformed: {{transformed | length}}
      - Records dropped: {{raw_data | length - transformed | length}}
      - Duration: {{duration}}
```

---

## Complete Example

```yaml
name: production_etl_pipeline
version: 1.0.0

config:
  defaults:
    provider: anthropic
    model: claude-sonnet-4

steps:
  # 1. Extract
  - name: extract
    servers: [database]
    prompt: "Extract: {{source_query}}"
    output: raw_data
    error_handling:
      on_failure: retry
      max_retries: 3
  
  # 2. Validate
  - name: validate
    prompt: "Validate: {{raw_data}}"
    output: validation
  
  # 3. Transform (if valid)
  - name: transform
    condition: "{{validation}} contains 'VALID'"
    parallel:
      - name: clean
        prompt: "Clean: {{raw_data}}"
      - name: enrich
        servers: [api]
        prompt: "Enrich: {{raw_data}}"
    aggregate: merge
    output: transformed
  
  # 4. Quality check
  - name: quality_check
    prompt: "Check quality: {{transformed}}"
    output: quality_report
  
  # 5. Load (if quality OK)
  - name: load
    condition: "{{quality_report}} contains 'PASS'"
    servers: [database]
    prompt: "Load: {{transformed}}"
    error_handling:
      on_failure: retry
      max_retries: 3
  
  # 6. Log results
  - name: log_pipeline
    prompt: |
      Log pipeline run:
      - Extracted: {{raw_data | length}}
      - Transformed: {{transformed | length}}
      - Loaded: {{load.count}}
      - Status: {{load.status}}
```

---

## Quick Reference

```yaml
# Basic ETL
extract → validate → transform → load

# Parallel ETL
parallel_extract → merge → transform → load

# Quality-focused
profile → detect_issues → clean → validate → load

# Incremental
get_last_run → extract_new → process → upsert

# Aggregation
group → aggregate → summarize
```

---

## Next Steps

- **[Document Analysis](document-analysis.md)** - Process documents
- **[Validation Pattern](validation.md)** - Multi-provider checks
- **[Examples](../examples/)** - Working pipelines

---

**Build robust data pipelines!** 🔄
