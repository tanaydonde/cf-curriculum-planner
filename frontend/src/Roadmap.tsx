import { useEffect, useState } from 'react';
import { ReactFlow, Background} from '@xyflow/react';
import type { Node, Edge } from '@xyflow/react';
import { MarkerType } from '@xyflow/react';
import { BackgroundVariant } from '@xyflow/react';
import '@xyflow/react/dist/style.css';

const POSITION_MAP: Record<string, { x: number, y: number }> = {
  "implementation": {x: 400, y: 0},

  "ad hoc": {x: 0, y: 150},
  "sortings": {x: 150, y: 150},
  "data structures": {x: 300, y: 150},
  "greedy": {x: 450, y: 150},
  "strings": {x: 600, y: 150},
  "math": {x: 750, y: 150},

  "searching": {x: 200, y: 300},
  "advanced math": {x: 700, y: 300},
  "geometry": {x: 850, y: 300},

  "two pointers": {x: 0, y: 450},
  "meet in the middle": {x: 150, y: 450},
  "graphs": {x: 300, y: 450},
  "dynamic programming": {x: 450, y: 450},
  "advanced strings": {x: 600, y: 450},

  "advanced graphs": {x: 225, y: 600},
  "trees": {x: 375, y: 600},
  
  "tree dp": {x: 600, y: 750},
};

interface BackendNode {
  id: number;
  slug: string;
  display_name: string;
}

interface BackendEdge {
  from: number;
  to: number;
}

const Roadmap = () => {
  const [nodes, setNodes] = useState<Node[]>([]);
  const [edges, setEdges] = useState<Edge[]>([]);

  useEffect(() => {
    fetch('http://localhost:8080/api/graph')
      .then(res => res.json())
      .then((data: { nodes: BackendNode[], edges: BackendEdge[] }) => {
        
        const formattedNodes: Node[] = data.nodes.map((node) => ({
          id: node.id.toString(),
          data: {label: node.display_name},
          position: POSITION_MAP[node.slug] || { x: 0, y: 0 },
          style: {
            background: '#1e293b',
            color: '#e5e7eb',
            borderRadius: 10,
            border: '1px solid #374151', 
            padding: '12px 14px',
            width: 140,
            fontSize: 12,
            fontWeight: 500,
            textAlign: 'center',
            boxShadow: '0 0 0 1px rgba(56,189,248,0.35), 0 10px 30px rgba(0,0,0,0.55)',
          }
        }));

        const formattedEdges: Edge[] = data.edges.map((edge, index) => ({
          id: `e-${index}`,
          source: edge.from.toString(),
          target: edge.to.toString(),
          //type: 'smoothstep',
          markerEnd: { type: MarkerType.ArrowClosed },
          style: {
            stroke: '#e5e7eb',   // light gray
            strokeWidth: 2,
            opacity: 0.85,
          },
        }));

        const bottomPadding = 75;

        const maxY = Math.max(
          ...formattedNodes.map((n) => n.position.y + (n.height ?? 0))
        );

        const spacerNode: Node = {
            id: '__bottom_spacer__',
            data: { label: '' },
            position: { x: 0, y: maxY + bottomPadding },
            style: {
                opacity: 0,
                width: 1,
                height: 1,
                pointerEvents: 'none',
            },
            selectable: false,
            draggable: false,
        };

        setNodes([...formattedNodes, spacerNode]);
        setEdges(formattedEdges);
      })
      .catch(err => console.error("can't load graph:", err));
  }, []);

  return (
    <div className="w-full h-full">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        elevateEdgesOnSelect
        nodesDraggable={false}
        nodesConnectable={false}
        elementsSelectable={true}
        panOnDrag={true}
        zoomOnScroll={true}
        fitView
        //fitViewOptions={{ padding: 0.2}}
      >
        <Background
          variant={BackgroundVariant.Lines}
          gap={36}
          size={1}
          color="rgba(255,255,255,0.04)"
        />
      </ReactFlow>
    </div>
  );
};

export default Roadmap;