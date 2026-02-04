#!/bin/bash
# Explore the GraphRAG knowledge graph

COMMAND="$1"

if [ -z "$COMMAND" ]; then
  echo "Usage: ./explore_graphrag.sh <command> [args]"
  echo ""
  echo "Commands:"
  echo "  stats          - Show graph statistics"
  echo "  top [N]        - Show N most connected entities (default: 20)"
  echo "  list <TYPE>    - List entities of type (CONCEPT, ACTOR, CONTROL, etc.)"
  echo ""
  echo "Examples:"
  echo "  ./explore_graphrag.sh stats"
  echo "  ./explore_graphrag.sh top 50"
  echo "  ./explore_graphrag.sh list CONCEPT"
  exit 1
fi

cd /media/laurie/Data/Github/mcp-cli-go

GRAPH_FILE="/tmp/mcp-outputs/rlm_poc/knowledge_graph.json"
OUTPUT_FILE="/tmp/mcp-outputs/rlm_poc/explore_results.json"

if [ "$COMMAND" = "stats" ]; then
  ./config/skills/python-context-builder/scripts/explore_graph.py stats "$GRAPH_FILE" "$OUTPUT_FILE"
  cat "$OUTPUT_FILE" | jq '.'
  
elif [ "$COMMAND" = "top" ]; then
  N="${2:-20}"
  ./config/skills/python-context-builder/scripts/explore_graph.py top "$GRAPH_FILE" "$OUTPUT_FILE" "$N"
  echo "Top $N Most Connected Entities:"
  cat "$OUTPUT_FILE" | jq -r '.entities[] | "\(.connections)\t\(.type)\t\(.text)"' | column -t -s $'\t'
  
elif [ "$COMMAND" = "list" ]; then
  if [ -z "$2" ]; then
    echo "Error: Please specify entity type"
    echo "Example: ./explore_graphrag.sh list CONCEPT"
    exit 1
  fi
  TYPE="$2"
  LIMIT="${3:-50}"
  ./config/skills/python-context-builder/scripts/explore_graph.py list "$GRAPH_FILE" "$OUTPUT_FILE" "$TYPE" "$LIMIT"
  echo "Entities of type $TYPE (showing first $LIMIT):"
  cat "$OUTPUT_FILE" | jq -r '.entities[] | "\(.id)\t\(.text)"' | column -t -s $'\t'
  
else
  echo "Unknown command: $COMMAND"
  exit 1
fi
