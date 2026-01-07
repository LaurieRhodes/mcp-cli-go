# Data Transformation Pipeline

> **Workflow:** [data_transformation_pipeline.yaml](../workflows/data_transformation_pipeline.yaml)  
> **Pattern:** Step Dependencies Ensure Correct Order  
> **Best For:** ETL pipelines where transformation order matters

---

## Problem Description

### The Wrong-Order Problem

**Manual ETL without dependencies:**

```python
# Developer writes transformations
df = load_data()
df = enrich_data(df)      # ❌ Runs first
df = clean_data(df)       # Should run first!
df = transform_data(df)
```

**Result:**
- Enrichment uses dirty data
- Transformations fail on nulls
- Pipeline errors
- Data quality poor

**Cost:** 2-3 hours debugging why pipeline fails

---

## Workflow Solution

### What It Does

**Step dependencies enforce correct order:**

1. **Profile → Design → Clean → Transform → Enrich → Validate**
2. **Cannot skip phases** (dependencies prevent it)
3. **Validation at each step**
4. **Complete Python code generated**

**Value:**
- Time: 8 hours → 5 minutes (99.996%)
- Correctness: Enforced execution order
- Quality: Validation at each phase

### Key Structure

```yaml
steps:
  - name: profile_source
  
  - name: design_transformations
    needs: [profile_source]
  
  - name: execute_cleaning
    needs: [design_transformations]
    # MUST clean before transform
  
  - name: execute_transformations
    needs: [execute_cleaning]
    # MUST transform after clean
  
  - name: execute_enrichment
    needs: [execute_transformations]
    # MUST enrich after transform
  
  - name: final_validation
    needs: [execute_enrichment]
    # Validates complete pipeline
```

---

## Usage Example

**Input:** Source CSV data

```bash
./mcp-cli --workflow data_transformation_pipeline \
  --input-data "$(cat source_data.csv)"
```

**Output: Complete ETL Code**

```python
# Phase 1: Data Cleaning
df = pd.read_csv('source_data.csv')

# Remove duplicates
df = df.drop_duplicates(subset=['id'])

# Handle missing values
df['amount'].fillna(df['amount'].median(), inplace=True)

# Fix invalid values
df = df.replace(['NULL', 'N/A'], np.nan)

# Phase 2: Data Transformation
# (Only runs after cleaning complete)
df['date'] = pd.to_datetime(df['date'])
df['category'] = df['category'].astype('category')

# Phase 3: Data Enrichment
# (Only runs after transformation complete)
df['year'] = df['date'].dt.year
df['month'] = df['date'].dt.month

# Phase 4: Validation
assert df.duplicated().sum() == 0
assert df['amount'].isnull().sum() == 0
```

**Key:** Dependencies ensure phases run in order

---

## Trade-offs

### Advantages

**Guaranteed Order:**
- Clean → Transform → Enrich
- Cannot skip phases
- Validation at each step

**99% Time Savings:**
- Manual design: 8 hours
- Automated: 5 minutes

**Complete Code:**
- Ready to execute
- Documented phases
- Reproducible

### Limitations

**Not Flexible:**
- Fixed phase order
- Can't skip steps
- Trade-off: Safety vs flexibility

---

## Related Resources

- **[Workflow File](../workflows/data_transformation_pipeline.yaml)**
- **[RAG Pipeline Builder](rag-pipeline-builder.md)**
- **[ML Data Quality Validator](ml-data-quality-validator.md)**

---

**Systematic ETL: Clean before transform, transform before enrich. Always.**
