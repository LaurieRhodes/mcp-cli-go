# Iterative Development Workflows

This directory contains a complete demonstration of iterative, agentic development using LLM loops.

## Workflow Architecture

```
User Request
    ↓
[planner] → Requirements & Milestones
    ↓
[test_designer] → Test Criteria
    ↓
[LOOP: develop_until_pass] ← Iterates until tests pass
    ├─ [dev_cycle]
    │   ├─ write code
    │   └─ review code
    ↓
[report] → Final Summary
```

## Individual Workflows

### 1. planner.yaml
**Purpose:** Analyzes request and creates clear requirements
- Input: User's coding request
- Output: Requirements and success criteria

### 2. test_designer.yaml
**Purpose:** Creates test criteria for validation
- Input: Requirements from planner
- Output: 3-5 specific test cases

### 3. code_writer.yaml
**Purpose:** Writes or improves code
- Input: Requirements, previous code, feedback
- Output: Python code

### 4. code_reviewer.yaml
**Purpose:** Reviews code against tests
- Input: Code and test criteria
- Output: "PASS" or "FAIL: [reason]"

### 5. dev_cycle.yaml
**Purpose:** Single iteration of write + review
- Input: Requirements, tests, previous attempt
- Output: Code + review result

### 6. iterative_developer.yaml ⭐
**Purpose:** Main orchestrator - runs complete iterative development
- Calls planner → test_designer → loop(dev_cycle) → report
- Exits early when code passes all tests
- Max 5 iterations for safety

## How It Works

1. **Planning Phase:**
   - User provides request (e.g., "Write a fibonacci function")
   - Planner breaks it into requirements
   - Test designer creates validation criteria

2. **Development Loop:**
   - Iteration 1: Write code → Review → likely FAIL
   - Iteration 2: Improve based on feedback → Review → maybe FAIL
   - Iteration 3: Improve again → Review → hopefully PASS
   - **Exits when:** Review says "PASS" OR 5 iterations reached

3. **Reporting:**
   - Summarizes iterations and final code
   - Shows development history

## Usage

### Run Complete Iterative Development
```bash
./mcp-cli --workflow iterative_developer \
  --input-data "Write a function to calculate fibonacci numbers"
```

### Run Individual Components (for testing)
```bash
# Just planning
./mcp-cli --workflow planner \
  --input-data "Write a fibonacci function"

# Just test design
./mcp-cli --workflow test_designer \
  --input-data "Function should calculate fibonacci numbers"

# Single dev cycle
./mcp-cli --workflow dev_cycle \
  --input-data "requirements: fibonacci function"
```

## Example Requests

**Simple:**
- "Write a function to calculate fibonacci numbers"
- "Create a function to validate email addresses"
- "Write a function to reverse a string"

**Moderate:**
- "Create a class to manage a todo list with add/remove/list methods"
- "Write a function to find the longest common substring"

**Complex:**
- "Create a simple calculator class with basic operations"
- "Write a function to parse and evaluate mathematical expressions"

## Expected Behavior

**Successful run:**
```
[INFO] Executing step: requirements
Requirements created ✓

[INFO] Executing step: test_criteria  
Tests designed ✓

[INFO] Starting loop: develop_until_pass (max 5 iterations)
[INFO] Loop iteration 1/5
Review: FAIL: Missing edge case handling

[INFO] Loop iteration 2/5
Review: FAIL: Incorrect base case

[INFO] Loop iteration 3/5
Review: PASS: All tests met ✓

[INFO] Loop exit condition met after 3 iterations

[INFO] Executing step: report
Summary: Built fibonacci function in 3 iterations ✓
```

## Configuration

All workflows use:
- **Provider:** deepseek (has credit)
- **Model:** deepseek-chat
- **Temperature:** 0.3-0.7 (varies by task)
- **Logging:** verbose (for main orchestrator)

To use a different provider, edit the `execution:` section in each workflow.

## Loop Configuration

```yaml
loops:
  - name: develop_until_pass
    workflow: dev_cycle
    max_iterations: 5           # Safety limit
    until: "The review says PASS"  # LLM evaluates
    on_failure: continue        # Keep trying
    accumulate: development_history  # Store all attempts
```

**Key settings:**
- `max_iterations: 5` - Won't run forever
- `until: "The review says PASS"` - Semantic exit condition
- `on_failure: continue` - Don't give up on errors

## Troubleshooting

**Problem:** Loop runs all 5 iterations without exiting early
**Solution:** Check that code_reviewer uses EXACT format: "PASS: ..." or "FAIL: ..."

**Problem:** API key errors
**Solution:** Ensure DEEPSEEK_API_KEY is set: `export DEEPSEEK_API_KEY='your-key'`

**Problem:** Condition evaluation failing
**Solution:** Keep condition simple: "The review says PASS" (don't use {{loop.output}})

## Design Principles

1. **Simple YAML** - Easy for non-programmers to modify
2. **Semantic Control** - LLM decides when code is good enough
3. **Context Isolation** - Each workflow runs independently
4. **Reusable Components** - Workflows can be used standalone
5. **Safe Iteration** - max_iterations prevents runaway loops

## Next Steps

**Extend this system:**
- Add more sophisticated test generation
- Include actual test execution (not just review)
- Add code formatting/linting steps
- Create specialized reviewers (security, performance, style)
- Build multi-file project development

**Key insight:** The loop system enables LLMs to iteratively improve their work until quality criteria are met - true agentic development! 🚀
