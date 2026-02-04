╔══════════════════════════════════════════════════════════════════╗
║              GRAPHRAG USAGE GUIDE - PRACTICAL EXAMPLES           ║
╚══════════════════════════════════════════════════════════════════╝

🎉 YOUR GRAPHRAG SYSTEM IS COMPLETE AND READY TO USE!

WHAT YOU HAVE:
==============
✅ 531 ISM policy document chunks
✅ 2,337 extracted entities (concepts, actors, controls)
✅ 2,313 relationships between entities
✅ Complete knowledge graph with intelligent traversal

FILES:
======
Knowledge Graph:    /tmp/mcp-outputs/rlm_poc/knowledge_graph.json
Entity Mappings:    /tmp/mcp-outputs/rlm_poc/graph_chunks/entities_*.json
Original Chunks:    /tmp/mcp-outputs/rlm_poc/chunks.json

═══════════════════════════════════════════════════════════════════
                    QUICK START - 3 WAYS TO USE
═══════════════════════════════════════════════════════════════════

METHOD 1: ASK QUESTIONS (EASIEST) 🎯
═════════════════════════════════════

Just ask a question in natural language:

  /tmp/ask_graphrag.sh "email security requirements"

The system will:
  1. Find entities matching your question
  2. Traverse the graph (2-hop by default)
  3. Find all relevant chunks
  4. Synthesize a comprehensive answer
  5. Cite sources (chunk IDs)

EXAMPLE QUERIES:

  /tmp/ask_graphrag.sh "TOP SECRET classification"
  /tmp/ask_graphrag.sh "encryption requirements"
  /tmp/ask_graphrag.sh "cyber security incidents"
  /tmp/ask_graphrag.sh "access control"
  /tmp/ask_graphrag.sh "personnel security clearances"

METHOD 2: EXPLORE THE GRAPH 🔍
═══════════════════════════════

Discover what's in the knowledge graph:

  # Show overall statistics
  /tmp/explore_graphrag.sh stats

  # Find most connected concepts (hubs)
  /tmp/explore_graphrag.sh top 20

  # List all entities of a type
  /tmp/explore_graphrag.sh list CONCEPT
  /tmp/explore_graphrag.sh list ACTOR
  /tmp/explore_graphrag.sh list CONTROL

METHOD 3: PROGRAMMATIC ACCESS 💻
═════════════════════════════════

Use the workflow directly for automation:

  echo '{"question":"email security"}' | \
    ./mcp-cli --workflow rlm_poc/workflows/ask_graphrag

═══════════════════════════════════════════════════════════════════
                        PRACTICAL USE CASES
═══════════════════════════════════════════════════════════════════

USE CASE 1: Policy Research & Compliance
═════════════════════════════════════════

Scenario: You need to understand email security requirements

Command:
  /tmp/ask_graphrag.sh "email security requirements"

What GraphRAG does:
  1. Finds: "email", "security", "protective_markings", "data_spills"
  2. Traverses to related: "encryption", "classification", "attachments"
  3. Returns chunks: CHUNK-089, CHUNK-134, CHUNK-201, etc.
  4. Synthesizes answer with full context

Result:
  Comprehensive answer covering:
  - Protective marking requirements
  - Encryption controls
  - Classification handling
  - Data spill prevention
  - All with ISM chunk references

USE CASE 2: Gap Analysis
════════════════════════

Scenario: Check if your system meets all encryption requirements

Step 1: Find all encryption-related entities
  /tmp/ask_graphrag.sh "encryption cryptographic"

Step 2: Review the related concepts
  (Graph shows: key_management, algorithms, storage, transmission)

Step 3: Compare against your implementation
  - Review each entity
  - Check each relationship
  - Identify gaps

USE CASE 3: Understanding Relationships
════════════════════════════════════════

Scenario: How does classification relate to access control?

Command:
  /tmp/ask_graphrag.sh "classification access control relationship"

GraphRAG traverses:
  classification → requires → clearance
  clearance → determines → access_level
  access_level → controls → data_access
  
Result: Full chain of relationships with source chunks

USE CASE 4: Finding Similar Concepts
═════════════════════════════════════

Scenario: Find all concepts similar to "cyber incidents"

Command:
  /tmp/ask_graphrag.sh "cyber incident"

Graph finds:
  - security_incidents
  - data_breaches
  - unauthorized_access
  - incident_response
  - reporting_requirements

All connected concepts with relationships!

USE CASE 5: Entity Type Exploration
════════════════════════════════════

Scenario: What organizational actors are mentioned?

Command:
  /tmp/explore_graphrag.sh list ACTOR

Returns:
  - ACSC (Australian Cyber Security Centre)
  - Chief_Security_Officer
  - System_Administrator
  - Security_Personnel
  - etc.

═══════════════════════════════════════════════════════════════════
                      ADVANCED FEATURES
═══════════════════════════════════════════════════════════════════

MULTI-HOP TRAVERSAL
═══════════════════

The graph traverses relationships in multiple hops:

1-hop: Direct connections only
  entity → related_entity

2-hop (default): Two degrees of separation
  entity → related → second_level

3-hop: Extended network
  entity → related → second_level → third_level

Adjust in workflow or script for broader/narrower search.

ENTITY TYPE FILTERING
══════════════════════

Entities are typed for precise queries:

  CONCEPT:      Policy concepts, definitions
  ACTOR:        Organizational roles, personnel
  CONTROL:      ISM controls, requirements
  PROCESS:      Procedures, workflows
  SYSTEM:       Technical systems, infrastructure
  DOCUMENT:     Forms, reports, documentation

SUBGRAPH EXTRACTION
═══════════════════

Query results include a focused subgraph:
  - Only relevant nodes
  - Only relevant edges
  - Ready for visualization

CHUNK-ENTITY MAPPING
═════════════════════

Every entity links back to source chunks:
  entity → [CHUNK-001, CHUNK-045, CHUNK-123]

Full traceability to original policy text!

═══════════════════════════════════════════════════════════════════
                     INTEGRATION PATTERNS
═══════════════════════════════════════════════════════════════════

PATTERN 1: Enhanced RAG
═══════════════════════

Traditional RAG: Query → Find chunks → Answer
GraphRAG: Query → Find entities → Traverse graph → Find chunks → Answer

Benefits:
  ✅ More comprehensive (finds related concepts)
  ✅ Better context (includes relationships)
  ✅ Explainable (shows reasoning path)

PATTERN 2: Compliance Dashboard
═══════════════════════════════

Build a dashboard that:
  1. Lists all security controls (by type)
  2. Shows relationships between controls
  3. Links to implementation evidence
  4. Identifies gaps

Query examples:
  - /tmp/explore_graphrag.sh list CONTROL
  - /tmp/ask_graphrag.sh "encryption controls"
  - /tmp/ask_graphrag.sh "access control requirements"

PATTERN 3: Policy Navigator
════════════════════════════

Help users navigate complex policy:

  User: "email security"
  Graph: Shows related: encryption, classification, data_spills
  User clicks: "encryption"
  Graph: Shows related: algorithms, key_management, storage
  User drills down further...

PATTERN 4: Question-Answer System
══════════════════════════════════

Build a Q&A interface:
  - User asks question
  - GraphRAG finds answer
  - Returns answer with sources
  - User can explore related concepts

═══════════════════════════════════════════════════════════════════
                     EXAMPLE WORKFLOWS
═══════════════════════════════════════════════════════════════════

WORKFLOW 1: Security Assessment
════════════════════════════════

Goal: Assess email security posture

Commands:
  1. /tmp/ask_graphrag.sh "email security requirements"
  2. /tmp/ask_graphrag.sh "email encryption"
  3. /tmp/ask_graphrag.sh "email classification"
  4. /tmp/ask_graphrag.sh "email data spills"

Compile results into assessment report.

WORKFLOW 2: Implementation Planning
════════════════════════════════════

Goal: Plan new system implementation

Commands:
  1. /tmp/explore_graphrag.sh list CONTROL
  2. /tmp/ask_graphrag.sh "[your system type] requirements"
  3. /tmp/ask_graphrag.sh "[technology] controls"
  4. Build implementation checklist

WORKFLOW 3: Incident Response
══════════════════════════════

Goal: Understand incident reporting requirements

Commands:
  1. /tmp/ask_graphrag.sh "security incidents"
  2. /tmp/ask_graphrag.sh "incident reporting"
  3. /tmp/ask_graphrag.sh "data breaches"
  4. Create incident response plan

═══════════════════════════════════════════════════════════════════
                      TIPS & BEST PRACTICES
═══════════════════════════════════════════════════════════════════

✅ Start Broad, Then Narrow
  Begin with general terms, drill down based on results

✅ Use Multiple Queries
  Different phrasings can reveal different connections

✅ Explore Entity Types
  Use explore_graphrag.sh list to discover available entities

✅ Follow the Graph
  Let relationships guide you to related concepts

✅ Check Source Chunks
  Always verify answers against original policy text

✅ Combine Methods
  Use exploration to discover, queries to answer

═══════════════════════════════════════════════════════════════════
                     TROUBLESHOOTING
═══════════════════════════════════════════════════════════════════

Q: No results found
A: Try broader terms, check spelling, explore entity types first

Q: Too many results
A: Use more specific terms, combine multiple concepts

Q: Missing connections
A: Some entities may not be connected (isolated nodes)

Q: Unexpected results
A: GraphRAG finds semantic relationships - explore to understand

═══════════════════════════════════════════════════════════════════
                      NEXT STEPS
═══════════════════════════════════════════════════════════════════

1. TRY IT NOW:
   /tmp/ask_graphrag.sh "cyber security"

2. EXPLORE THE GRAPH:
   /tmp/explore_graphrag.sh stats
   /tmp/explore_graphrag.sh top 50

3. BUILD YOUR USE CASE:
   - Compliance checking
   - Policy navigation
   - Risk assessment
   - Training materials

═══════════════════════════════════════════════════════════════════

🎉 YOUR GRAPHRAG SYSTEM IS READY! START QUERYING! 🎉

═══════════════════════════════════════════════════════════════════
