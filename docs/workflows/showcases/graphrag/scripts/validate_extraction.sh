#!/bin/bash
# Validate GraphRAG entity extraction against original ISM text

if [ -z "$1" ]; then
  echo "Usage: ./validate_extraction.sh <chunk_id>"
  echo ""
  echo "Example: ./validate_extraction.sh CHUNK-305"
  echo ""
  echo "This validates extraction by showing:"
  echo "  1. Original ISM policy text (source of truth)"
  echo "  2. Extracted entities and their descriptions"
  echo "  3. Extracted relationships"
  echo "  4. Validation checklist"
  exit 1
fi

CHUNK_ID="$1"

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║          EXTRACTION VALIDATION: $CHUNK_ID                     "
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

# Check files exist
CHUNK_FILE="/tmp/mcp-outputs/rlm_poc/graph_chunks/entities_${CHUNK_ID}.json"
if [ ! -f "$CHUNK_FILE" ]; then
  echo "❌ Entity file not found: $CHUNK_FILE"
  exit 1
fi

# Show original text
echo "═══════════════════════════════════════════════════════════════════"
echo "📄 ORIGINAL ISM POLICY TEXT (Source of Truth):"
echo "═══════════════════════════════════════════════════════════════════"
echo ""
cat /tmp/mcp-outputs/rlm_poc/chunks.json | jq -r ".[] | select(.chunk_id == \"$CHUNK_ID\") | .content" | fold -w 70 -s
echo ""

# Show extracted entities
echo "═══════════════════════════════════════════════════════════════════"
echo "🔍 EXTRACTED ENTITIES:"
echo "═══════════════════════════════════════════════════════════════════"
echo ""
cat "$CHUNK_FILE" | jq -r '.entities[] | "  [\(.type)] \(.id)\n    ➜ \(.text)\n"'

# Show relationships
echo "═══════════════════════════════════════════════════════════════════"
echo "🔗 EXTRACTED RELATIONSHIPS:"
echo "═══════════════════════════════════════════════════════════════════"
echo ""
REL_COUNT=$(cat "$CHUNK_FILE" | jq '.relationships | length')
if [ "$REL_COUNT" -gt 0 ]; then
  cat "$CHUNK_FILE" | jq -r '.relationships[] | "  \(.from)\n    → [\(.type)]\n    → \(.to)\n"'
else
  echo "  (No relationships extracted from this chunk)"
fi
echo ""

# Validation checklist
echo "═══════════════════════════════════════════════════════════════════"
echo "✅ VALIDATION CHECKLIST:"
echo "═══════════════════════════════════════════════════════════════════"
echo ""
echo "Compare entities to original text above:"
echo ""
echo "  ☐ Are all key concepts captured as entities?"
echo "  ☐ Are entity types appropriate (CONCEPT, CONTROL, ACTOR, etc.)?"
echo "  ☐ Do entity descriptions match the source text meaning?"
echo "  ☐ Are relationships between entities accurate?"
echo "  ☐ Is anything important missing?"
echo "  ☐ Are there any hallucinated entities (not in source)?"
echo "  ☐ Do ISM control IDs match exactly?"
echo ""

# Summary stats
echo "═══════════════════════════════════════════════════════════════════"
echo "📊 EXTRACTION STATISTICS:"
echo "═══════════════════════════════════════════════════════════════════"
echo ""
ENTITY_COUNT=$(cat "$CHUNK_FILE" | jq '.entities | length')
WORD_COUNT=$(cat /tmp/mcp-outputs/rlm_poc/chunks.json | jq -r ".[] | select(.chunk_id == \"$CHUNK_ID\") | .content" | wc -w)
echo "  Original text: $WORD_COUNT words"
echo "  Entities extracted: $ENTITY_COUNT"
echo "  Relationships: $REL_COUNT"
echo "  Density: $(echo "scale=2; $ENTITY_COUNT * 100 / $WORD_COUNT" | bc)% (entities per 100 words)"
echo ""
echo "═══════════════════════════════════════════════════════════════════"
echo ""
echo "Files:"
echo "  📄 Source: /tmp/mcp-outputs/rlm_poc/chunks.json"
echo "  🔍 Entities: $CHUNK_FILE"
echo "  🌐 Graph: /tmp/mcp-outputs/rlm_poc/knowledge_graph.json"
