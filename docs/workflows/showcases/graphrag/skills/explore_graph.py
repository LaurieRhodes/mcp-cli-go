#!/usr/bin/env python3
"""
Explore the GraphRAG knowledge graph structure
"""

import sys
import json
from collections import Counter

def load_graph(graph_file):
    with open(graph_file, 'r') as f:
        return json.load(f)

def get_stats(graph):
    """Get graph statistics."""
    entity_types = Counter(n['type'] for n in graph['nodes'])
    relationship_types = Counter(e['type'] for e in graph['edges'])
    
    return {
        'total_nodes': len(graph['nodes']),
        'total_edges': len(graph['edges']),
        'entity_types': dict(entity_types.most_common()),
        'relationship_types': dict(relationship_types.most_common())
    }

def get_most_connected(graph, top_n=20):
    """Find most connected entities."""
    connections = Counter()
    for edge in graph['edges']:
        connections[edge['from']] += 1
        connections[edge['to']] += 1
    
    entity_map = {n['id']: n for n in graph['nodes']}
    
    result = []
    for entity_id, count in connections.most_common(top_n):
        if entity_id in entity_map:
            result.append({
                'id': entity_id,
                'text': entity_map[entity_id]['text'],
                'type': entity_map[entity_id]['type'],
                'connections': count
            })
    
    return result

def list_by_type(graph, entity_type, limit=50):
    """List entities of a specific type."""
    return [
        {'id': n['id'], 'text': n['text'], 'type': n['type']}
        for n in graph['nodes']
        if n['type'].upper() == entity_type.upper()
    ][:limit]

def main():
    command = sys.argv[1]
    graph_file = sys.argv[2]
    output_file = sys.argv[3]
    
    graph = load_graph(graph_file)
    
    if command == 'stats':
        result = get_stats(graph)
    elif command == 'top':
        top_n = int(sys.argv[4]) if len(sys.argv) > 4 else 20
        result = {
            'command': 'most_connected',
            'entities': get_most_connected(graph, top_n)
        }
    elif command == 'list':
        entity_type = sys.argv[4]
        limit = int(sys.argv[5]) if len(sys.argv) > 5 else 50
        result = {
            'command': 'list_by_type',
            'entity_type': entity_type,
            'entities': list_by_type(graph, entity_type, limit)
        }
    else:
        result = {'error': f'Unknown command: {command}'}
    
    with open(output_file, 'w') as f:
        json.dump(result, f, indent=2)
    
    print(json.dumps({'status': 'success', 'output_file': output_file}))

if __name__ == "__main__":
    main()
