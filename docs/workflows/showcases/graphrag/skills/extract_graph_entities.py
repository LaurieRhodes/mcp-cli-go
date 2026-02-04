#!/usr/bin/env python3
"""
Extract entities and relationships from chunk for knowledge graph.

GraphRAG Phase 1: Entity extraction

Usage:
    python3 extract_graph_entities.py <chunks_file> <chunk_id> <output_file>
"""
import sys
import json
import re
from pathlib import Path
from typing import Dict, List, Any

sys.path.insert(0, str(Path(__file__).parent.parent / 'lib'))
from validation import (
    ValidationError, load_json_file, save_json_file,
    handle_validation_error, print_success,
)

def load_chunk(chunks: List[Dict], chunk_id: str) -> Dict:
    for chunk in chunks:
        if chunk.get('chunk_id') == chunk_id:
            return chunk
    raise ValidationError(f"Chunk {chunk_id} not found", "CHUNK_NOT_FOUND")

def extract_entities_simple(content: str, chunk_id: str) -> Dict:
    """Simple pattern-based extraction (LLM will do better in workflow)"""
    entities = []
    relationships = []
    
    # Extract control IDs
    controls = re.findall(r'\b(ISM-\d+)\b', content)
    for control in set(controls):
        entities.append({"id": control, "type": "CONTROL", "text": f"Control {control}"})
    
    # Extract common roles/concepts
    patterns = {
        r'\bCISO\b': ('CISO', 'ROLE'),
        r'\bMFA\b': ('MFA', 'CONCEPT'),
        r'\bauthentication\b': ('authentication', 'CONCEPT'),
        r'\bencryption\b': ('encryption', 'CONCEPT'),
    }
    
    for pattern, (entity_id, entity_type) in patterns.items():
        if re.search(pattern, content, re.IGNORECASE):
            entities.append({"id": entity_id, "type": entity_type, "text": entity_id})
    
    # Simple relationships
    for i in range(min(len(entities) - 1, 5)):
        relationships.append({
            "from": entities[i]['id'],
            "to": entities[i+1]['id'],
            "type": "RELATES_TO"
        })
    
    return {"chunk_id": chunk_id, "entities": entities, "relationships": relationships}

def main():
    if len(sys.argv) != 4:
        print("Usage: extract_graph_entities.py <chunks_file> <chunk_id> <output_file>", file=sys.stderr)
        sys.exit(1)
    
    try:
        chunks_data = load_json_file(sys.argv[1])
        chunk = load_chunk(chunks_data, sys.argv[2])
        result = extract_entities_simple(chunk.get('content', ''), sys.argv[2])
        save_json_file(result, sys.argv[3])
        print_success(f"Extracted {len(result['entities'])} entities, {len(result['relationships'])} relationships")
    except ValidationError as e:
        handle_validation_error(e, "extract_graph_entities")
        sys.exit(1)

if __name__ == '__main__':
    main()
