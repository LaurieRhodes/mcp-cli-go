# Iterative Development Demo Guide

## 🎯 Quick Start

Set your API key and run:
```bash
export DEEPSEEK_API_KEY='your-key-here'

./mcp-cli --workflow iterative_developer \
  --input-data "Write a function to calculate fibonacci numbers"
```

## 📁 What's In This Directory

```
config/workflows/iterative_dev/
├── README.md              ← Full documentation
├── DEMO_GUIDE.md         ← This file (quick start)
│
├── planner.yaml          ← Analyzes request → requirements
├── test_designer.yaml    ← Creates test criteria
├── code_writer.yaml      ← Writes/improves code
├── code_reviewer.yaml    ← Reviews code → PASS/FAIL
│
├── dev_cycle.yaml        ← One iteration (write + review)
├── iterative_developer.yaml ← Main orchestrator ⭐
│
└── simple_test.yaml      ← Verification test
```

## 🚀 Demo Scenarios

### 1. Complete Iterative Development (Full Demo)
```bash
./mcp-cli --workflow iterative_developer \
  --input-data "Write a function to calculate fibonacci numbers"
```

**What happens:**
1. Creates requirements ✓
2. Designs test criteria ✓
3. **Loop starts** (max 5 iterations):
   - Iteration 1: Write code → Review → likely FAIL
   - Iteration 2: Improve → Review → maybe FAIL  
   - Iteration 3: Improve → Review → **PASS** → EXIT!
4. Reports summary ✓

**Expected output:**
```
[INFO] Executing step: requirements
Requirements: Calculate fibonacci sequence...

[INFO] Executing step: test_criteria
Tests: Should handle n=0, n=1, n=5...

[INFO] Starting loop: develop_until_pass (max 5)
[INFO] Loop iteration 1/5
Review: FAIL: Missing base case

[INFO] Loop iteration 2/5
Review: FAIL: Incorrect recursive logic

[INFO] Loop iteration 3/5
Review: PASS: All tests met

[INFO] Loop exit condition met after 3 iterations ✓

[INFO] Executing step: report
Built fibonacci function in 3 iterations
```

### 2. Individual Components (Testing)

**Planning only:**
```bash
./mcp-cli --workflow planner \
  --input-data "Write a function to validate emails"
```

**Test design only:**
```bash
./mcp-cli --workflow test_designer \
  --input-data "Function should validate email addresses with proper format"
```

**Single dev cycle:**
```bash
./mcp-cli --workflow dev_cycle \
  --input-data "requirements: email validator"
```

### 3. Different Coding Tasks

**Easy (1-2 iterations):**
```bash
./mcp-cli --workflow iterative_developer \
  --input-data "Write a function to reverse a string"
```

**Medium (2-3 iterations):**
```bash
./mcp-cli --workflow iterative_developer \
  --input-data "Create a function to validate email addresses"
```

**Complex (3-5 iterations):**
```bash
./mcp-cli --workflow iterative_developer \
  --input-data "Write a function to parse and evaluate simple math expressions"
```

## 🎓 Key Concepts Demonstrated

### 1. Workflow Composition
- Workflows call other workflows
- Clean separation of concerns
- Reusable components

### 2. Iterative Loops
- Automatic iteration until quality criteria met
- LLM evaluates exit conditions semantically
- Safe with max_iterations limit

### 3. Context Isolation
- Each workflow runs independently
- Previous results passed as parameters
- No context pollution

### 4. Agentic Development
- LLM decides when code is good enough
- Iterative improvement based on feedback
- Measurable progress (PASS/FAIL)

## 📊 Expected Performance

**Timing (with deepseek):**
- Planning: ~3 seconds
- Test design: ~3 seconds
- Each iteration: ~5-6 seconds (write + review)
- Total for 3 iterations: ~20-25 seconds

**Iteration patterns:**
- Simple tasks: 1-2 iterations (usually FAIL → PASS)
- Medium tasks: 2-3 iterations
- Complex tasks: 3-4 iterations
- Rarely needs all 5 iterations

## 🔧 Troubleshooting

**Problem:** "API key required"
```bash
export DEEPSEEK_API_KEY='your-key-here'
```

**Problem:** Loop runs all 5 iterations without exiting
- Check that code_reviewer outputs start with "PASS:" or "FAIL:"
- Condition evaluation needs clear markers

**Problem:** Workflow not found
```bash
# List all workflows
./mcp-cli --list-workflows | grep iterative
```

**Problem:** Want to use different provider
- Edit `execution:` section in each workflow file
- Change `provider: deepseek` to your provider

## 🎯 Success Criteria

You'll know it's working when:
- ✅ Requirements are generated clearly
- ✅ Test criteria are specific and measurable
- ✅ Loop exits early (before max_iterations)
- ✅ Exit reason is "condition_met" not "max_iterations"
- ✅ Final code addresses the original request
- ✅ Summary shows iteration count

## 🚀 What This Demonstrates

This is a **complete proof of concept** for agentic development:

1. **Planning** - LLM breaks down requirements
2. **Test Design** - LLM creates validation criteria
3. **Iterative Development** - LLM writes, reviews, improves
4. **Semantic Control** - LLM decides when done
5. **Safety** - Max iterations prevents runaway
6. **Reporting** - Summary of process and results

**Key insight:** LLMs can iteratively improve their own work until objective criteria are met - true agentic behavior! 🎉

## 📈 Next Steps

**Extend this demo:**
- Add actual test execution (not just review)
- Include code formatting/linting
- Generate multiple files
- Add security review step
- Create specialized reviewers

**Production use:**
- Reduce max_iterations for faster prototyping
- Add cost tracking (API calls)
- Cache previous good solutions
- Add human approval step

---

**Ready to see it in action?**
```bash
export DEEPSEEK_API_KEY='your-key'
./mcp-cli --workflow iterative_developer --input-data "your coding task here"
```
