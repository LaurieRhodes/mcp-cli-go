#!/bin/bash
# Direct GraphRAG query without workflow complexity

if [ -z "$1" ]; then
  echo "Usage: ./ask_graphrag_direct.sh <your question>"
  echo ""
  echo "Examples:"
  echo "  ./ask_graphrag_direct.sh 'cyber security'"
  echo "  ./ask_graphrag_direct.sh 'email requirements'"
  exit 1
fi

QUESTION="$*"

cd /media/laurie/Data/Github/mcp-cli-go

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║                    GraphRAG Direct Query                          ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""
echo "Question: $QUESTION"
echo ""

# Run query script directly
./config/skills/python-context-builder/scripts/query_graphrag.py \
  "$QUESTION" \
  "/tmp/mcp-outputs/rlm_poc/knowledge_graph.json" \
  "/tmp/mcp-outputs/rlm_poc/graph_chunks" \
  "/tmp/mcp-outputs/rlm_poc/query_results.json" \
  2

echo ""
echo "═══════════════════════════════════════════════════════════════════"
echo "                           RESULTS"
echo "═══════════════════════════════════════════════════════════════════"
echo ""

if [ -f /tmp/mcp-outputs/rlm_poc/query_results.json ]; then
  STATUS=$(cat /tmp/mcp-outputs/rlm_poc/query_results.json | jq -r '.status')
  
  if [ "$STATUS" = "success" ]; then
    echo "✅ Query successful!"
    echo ""
    
    # Show summary
    echo "Summary:"
    cat /tmp/mcp-outputs/rlm_poc/query_results.json | jq '{
      matching_entities: .matching_entities | length,
      total_related: .total_related_entities,
      chunks_found: .relevant_chunks
    }'
    
    echo ""
    echo "Matching Entities:"
    cat /tmp/mcp-outputs/rlm_poc/query_results.json | jq -r '.matching_entities[] | "  • \(.type): \(.text)"' | head -10
    
    echo ""
    echo "Relevant Chunks (first 5):"
    cat /tmp/mcp-outputs/rlm_poc/query_results.json | jq -r '.chunks[0:5][] | "  • \(.chunk_id): \(.entities | length) entities"'
    
    echo ""
    echo "───────────────────────────────────────────────────────────────────"
    echo "Full results: /tmp/mcp-outputs/rlm_poc/query_results.json"
    echo ""
    
    # Now let's get the actual chunk content
    echo "Retrieving chunk content..."
    CHUNK_IDS=$(cat /tmp/mcp-outputs/rlm_poc/query_results.json | jq -r '.chunks[0:3][].chunk_id')
    
    echo ""
    echo "Top 3 Relevant Chunks:"
    echo ""
    
    for CHUNK_ID in $CHUNK_IDS; do
      CONTENT=$(cat /tmp/mcp-outputs/rlm_poc/chunks.json | jq -r ".[] | select(.chunk_id == \"$CHUNK_ID\") | .content")
      echo "[$CHUNK_ID]"
      echo "$CONTENT" | fold -w 65 -s | head -3
      echo "..."
      echo ""
    done
    
  else
    echo "❌ No matches found"
    cat /tmp/mcp-outputs/rlm_poc/query_results.json | jq -r '.message'
  fi
else
  echo "❌ Query failed - no results file created"
fi
