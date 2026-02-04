#!/usr/bin/env python3
"""
GraphRAG Query Engine - Find answers using the knowledge graph
"""

import sys
import json
import os
from pathlib import Path

def load_graph(graph_file):
    """Load the knowledge graph."""
    with open(graph_file, 'r') as f:
        return json.load(f)

def search_entities(graph, query_terms):
    """Find entities matching query terms."""
    query_lower = [term.lower() for term in query_terms]
    matches = []
    
    for node in graph['nodes']:
        node_text = (node.get('id', '') + ' ' + node.get('text', '')).lower()
        for term in query_lower:
            if term in node_text:
                matches.append(node)
                break
    
    return matches

def get_related_entities(graph, entity_ids, depth=2):
    """Get entities related to the given entities (multi-hop)."""
    related = set(entity_ids)
    
    for _ in range(depth):
        new_entities = set()
        for edge in graph['edges']:
            if edge['from'] in related:
                new_entities.add(edge['to'])
            if edge['to'] in related:
                new_entities.add(edge['from'])
        related.update(new_entities)
    
    return list(related)

def find_chunks_with_entities(entity_ids, graph_chunks_dir):
    """Find chunks containing the given entities."""
    chunks = []
    
    for entity_file in Path(graph_chunks_dir).glob('entities_CHUNK-*.json'):
        with open(entity_file, 'r') as f:
            data = json.load(f)
        
        chunk_entity_ids = [e['id'] for e in data.get('entities', [])]
        if any(eid in chunk_entity_ids for eid in entity_ids):
            chunks.append({
                'chunk_id': data['chunk_id'],
                'entities': data['entities'],
                'relationships': data.get('relationships', []),
                'file': str(entity_file)
            })
    
    return chunks

def get_subgraph(graph, entity_ids):
    """Extract subgraph for given entities."""
    nodes = [n for n in graph['nodes'] if n['id'] in entity_ids]
    edges = [e for e in graph['edges'] 
             if e['from'] in entity_ids and e['to'] in entity_ids]
    
    return {'nodes': nodes, 'edges': edges}

def main():
    query = sys.argv[1]
    graph_file = sys.argv[2]
    graph_chunks_dir = sys.argv[3]
    output_file = sys.argv[4]
    max_depth = int(sys.argv[5]) if len(sys.argv) > 5 else 2
    
    # Load graph
    graph = load_graph(graph_file)
    
    # Search for entities matching query
    query_terms = query.lower().split()
    matching_entities = search_entities(graph, query_terms)
    
    if not matching_entities:
        result = {
            'query': query,
            'status': 'no_matches',
            'message': 'No entities found matching the query',
            'suggestions': ['Try broader terms', 'Check spelling', 'Use key concepts']
        }
    else:
        # Get related entities (multi-hop)
        entity_ids = [e['id'] for e in matching_entities]
        related_ids = get_related_entities(graph, entity_ids, depth=max_depth)
        
        # Find chunks containing these entities
        relevant_chunks = find_chunks_with_entities(related_ids, graph_chunks_dir)
        
        # Get subgraph
        subgraph = get_subgraph(graph, related_ids)
        
        result = {
            'query': query,
            'status': 'success',
            'matching_entities': matching_entities,
            'total_related_entities': len(related_ids),
            'relevant_chunks': len(relevant_chunks),
            'chunks': relevant_chunks[:20],  # Limit to top 20
            'subgraph': {
                'nodes': len(subgraph['nodes']),
                'edges': len(subgraph['edges']),
                'data': subgraph
            },
            'statistics': {
                'direct_matches': len(matching_entities),
                'total_related': len(related_ids),
                'chunks_found': len(relevant_chunks),
                'search_depth': max_depth
            }
        }
    
    # Save result
    os.makedirs(os.path.dirname(output_file), exist_ok=True)
    with open(output_file, 'w') as f:
        json.dump(result, f, indent=2)
    
    # Print summary
    print(json.dumps({
        'status': 'success',
        'query': query,
        'entities_found': len(matching_entities) if matching_entities else 0,
        'related_entities': len(related_ids) if matching_entities else 0,
        'chunks_found': len(relevant_chunks) if matching_entities else 0,
        'output_file': output_file
    }))

if __name__ == "__main__":
    main()
