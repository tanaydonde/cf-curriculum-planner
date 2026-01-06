import { useCallback, useEffect, useState } from 'react';
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
type Difficulty = 'warmup' | 'growth' | 'challenge';

const DIFFICULTY_CONFIG = {
  warmup: { label: 'Warmup', inc: -50, color: 'text-emerald-400', border: 'border-emerald-400/20', hover: 'hover:bg-emerald-400/10' },
  growth: { label: 'Growth', inc: 25, color: 'text-sky-400', border: 'border-sky-400/20', hover: 'hover:bg-sky-400/10' },
  challenge: { label: 'Challenge', inc: 100, color: 'text-rose-400', border: 'border-rose-400/20', hover: 'hover:bg-rose-400/10' },
};

interface SubmitModalProps {
  isOpen: boolean;
  onClose: () => void;
  problem: Problem | null;
  onSubmit: (timeMinutes: number) => void;
  isSubmitting: boolean;
  isSuccess: boolean;
  error: string | null;
}

interface BackendNode {
  id: number;
  slug: string;
  display_name: string;
}

interface BackendEdge {
  from: number;
  to: number;
}

interface Problem {
  id: string;
  name: string;
  rating: number;
  tags: string[];
}

const SubmitModal = ({ isOpen, onClose, problem, onSubmit, isSubmitting, isSuccess, error }: SubmitModalProps) => {
  const [time, setTime] = useState<string>('');

  if (!isOpen || !problem) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/60 backdrop-blur-sm">
      <div className="bg-slate-900 border border-slate-700 rounded-xl shadow-2xl w-full max-w-md p-6 relative animate-in fade-in zoom-in duration-200">
        
        {isSuccess ? (
          <div className="flex flex-col items-center justify-center py-6 text-center animate-in fade-in slide-in-from-bottom-2 duration-300">
            <div className="w-16 h-16 bg-emerald-500/20 rounded-full flex items-center justify-center mb-4 ring-1 ring-emerald-500/50 shadow-[0_0_20px_rgba(16,185,129,0.2)]">
              <svg className="w-8 h-8 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <h3 className="text-xl font-bold text-gray-100 mb-1">Great Job!</h3>
            <p className="text-slate-400 text-sm">
              Problem marked as solved. Mastery updated.
            </p>
          </div>
        ) : (
          <>
            <h3 className="text-xl font-bold text-gray-100 mb-1">Verify Solution</h3>
            <p className="text-slate-400 text-sm mb-6">
              Did you solve <span className="text-sky-400 font-mono">{problem.id}</span> on Codeforces?
            </p>

            <div className="space-y-3 mb-6">
              <label className="block text-xs font-semibold text-slate-500 uppercase tracking-wider">
                Time Taken (Optional)
              </label>
              <div className="relative">
                <input 
                  type="number" 
                  value={time}
                  onChange={(e) => setTime(e.target.value)}
                  placeholder="0"
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-4 py-3 text-gray-200 focus:ring-2 focus:ring-sky-500 focus:border-transparent outline-none transition-all placeholder:text-slate-600 [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
                />
                <span className="absolute right-4 top-3.5 text-slate-500 text-sm">min</span>
              </div>
              <p className="text-[10px] text-slate-500">
                * Leave blank if you didn't time yourself.
              </p>
            </div>

            {error && (
              <div className="mb-4 p-3 bg-rose-500/10 border border-rose-500/20 rounded-lg text-rose-400 text-xs">
                {error}
              </div>
            )}

            <div className="flex gap-3 mt-2">
              <button 
                onClick={onClose}
                className="flex-1 px-4 py-2.5 rounded-lg bg-slate-800 text-slate-300 font-medium hover:bg-slate-700 transition-colors text-sm"
                disabled={isSubmitting}
              >
                Cancel
              </button>
              <button 
                onClick={() => onSubmit(time === '' ? 0 : parseInt(time))}
                disabled={isSubmitting}
                className="flex-1 px-4 py-2.5 rounded-lg bg-sky-500 text-white font-bold hover:bg-sky-400 transition-colors text-sm flex justify-center items-center gap-2 shadow-lg shadow-sky-500/20"
              >
                {isSubmitting ? (
                  <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"/>
                ) : (
                  <>Verify & Submit</>
                )}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
};

const getProblemLink = (problemId: string) => {
  const match = problemId.match(/^(\d+)(.+)$/);
  if (match) {
    return `https://codeforces.com/problemset/problem/${match[1]}/${match[2]}`;
  }
  return `https://codeforces.com/problemset/problem/${problemId}`;
};

const Training = () => {
  const [nodes, setNodes] = useState<Node[]>([]);
  const [edges, setEdges] = useState<Edge[]>([]);

  const [selectedTopic, setSelectedTopic] = useState<{slug: string, label: string} | null>(null);
  const [difficulty, setDifficulty] = useState<Difficulty>('growth');
  const [problems, setProblems] = useState<Problem[]>([]);
  const [loadingProblems, setLoadingProblems] = useState(false);

  const [showInfo, setShowInfo] = useState(true);

  const [modalOpen, setModalOpen] = useState(false);
  const [targetProblem, setTargetProblem] = useState<Problem | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [isSuccess, setIsSuccess] = useState(false);

  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchGraph = async () => {
      try {
        const res = await fetch('http://localhost:8080/api/graph');
        if (!res.ok) throw new Error('failed to fetch graph');
        const data: { nodes: BackendNode[]; edges: BackendEdge[] } = await res.json();

        const formattedNodes: Node[] = data.nodes.map((node) => ({
          id: node.id.toString(),
          data: { label: node.display_name, slug: node.slug },
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
          },
        }));

        const formattedEdges: Edge[] = data.edges.map((edge, index) => ({
          id: `e-${index}`,
          source: edge.from.toString(),
          target: edge.to.toString(),
          markerEnd: { type: MarkerType.ArrowClosed },
          style: { stroke: '#e5e7eb', strokeWidth: 2, opacity: 0.85 },
        }));

        const bottomPadding = 75;
        const maxY = Math.max(...formattedNodes.map((n) => n.position.y));

        const spacerNode: Node = {
          id: '__bottom_spacer__',
          data: { label: '' },
          position: { x: 0, y: maxY + bottomPadding },
          style: { opacity: 0, width: 1, height: 1, pointerEvents: 'none' },
          selectable: false,
          draggable: false,
        };

        setNodes([...formattedNodes, spacerNode]);
        setEdges(formattedEdges);
      } catch (err) {
        console.error("can't load graph:", err)
      } finally {
        setLoading(false);
      }
    }
    fetchGraph()
  }, []);

  const onNodeClick = useCallback((event: React.MouseEvent, node: Node) => {
    if (node.id === '__bottom_spacer__') return;
    
    setSelectedTopic({
      slug: node.data.slug as string,
      label: node.data.label as string
    });
    setDifficulty('growth'); 
  }, []);

  useEffect(() => {
    if (!selectedTopic) return;

    const fetchProblems = async () => {
      setLoadingProblems(true);
      const handle = sessionStorage.getItem('cf_handle');
      if (!handle) return;

      const inc = DIFFICULTY_CONFIG[difficulty].inc;
      
      try {
        const res = await fetch(`http://localhost:8080/api/problems/${selectedTopic.slug}?handle=${handle}&inc=${inc}`);
        if (!res.ok) throw new Error('failed to fetch');
        const data = await res.json();
        setProblems(data || []);
      } catch (err) {
        console.error(err);
        setProblems([]);
      } finally {
        setLoadingProblems(false);
      }
    };

    fetchProblems();
  }, [selectedTopic, difficulty]);

  const handleOpenVerify = (problem: Problem) => {
    setTargetProblem(problem);
    setSubmitError(null);
    setIsSuccess(false);
    setModalOpen(true);
  };

  const handleSubmitVerification = async (timeMinutes: number) => {
    if (!targetProblem) return;
    setIsSubmitting(true);
    setSubmitError(null);

    const handle = sessionStorage.getItem('cf_handle');
    if (!handle) {
        setSubmitError("No handle found. Please log in.");
        setIsSubmitting(false);
        return;
    }

    try {
      const res = await fetch(`http://localhost:8080/api/submit/${handle}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          problem_id: targetProblem.id,
          time_spent_minutes: timeMinutes
        })
      });

      if (!res.ok) {
        const errMsg = await res.text();
        throw new Error(errMsg || "Verification failed");
      }
      
      setIsSubmitting(false);
      setIsSuccess(true);

      setTimeout(() => {
        setModalOpen(false);
        setIsSuccess(false);
        setProblems(prev => prev.filter(p => p.id !== targetProblem.id));
      }, 2000);
      
    } catch (err: any) {
      let msg = err.message;
      if (msg.includes("not solved")) msg = "Codeforces says you haven't solved this yet! Please wait a minute if you just submitted.";
      else if (msg.includes("already solved")) msg = "You've already tracked this problem! We'll update your list.";
      
      setSubmitError(msg);
      setIsSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="h-full flex items-center justify-center text-slate-500 font-mono animate-pulse">
        LOADING...
      </div>
    );
  }

  return (
    <div className="w-full h-[calc(100vh-64px)] relative flex">
      
      {showInfo && (
        <div className="absolute top-4 left-4 z-10 bg-slate-900/90 backdrop-blur border border-slate-700 p-4 rounded-xl shadow-xl max-w-xs pointer-events-auto transition-all">
          <button 
            onClick={() => setShowInfo(false)}
            className="absolute top-2 right-2 text-slate-500 hover:text-white transition-colors"
          >
            ✕
          </button>

          <h3 className="text-sky-400 font-bold text-sm flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-sky-400 animate-pulse"/>
            Training
          </h3>
          <p className="text-slate-400 text-xs mt-2 leading-relaxed">
            Select any topic node to generate a personalized problem set tailored to your rating.
          </p>
          <div className="mt-3 flex gap-2 text-[10px] text-slate-500">
            <span className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-emerald-400/20 border border-emerald-400/50"></div> Warmup</span>
            <span className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-sky-400/20 border border-sky-400/50"></div> Growth</span>
            <span className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-rose-400/20 border border-rose-400/50"></div> Challenge</span>
          </div>
        </div>
      )}

      <div className="flex-1 h-full">
        <ReactFlow
          key={nodes.length}
          nodes={nodes}
          edges={edges}
          onNodeClick={onNodeClick}
          nodesDraggable={false}
          nodesConnectable={false}
          elementsSelectable={true}
          panOnDrag={true}
          zoomOnScroll={true}
          fitView
          maxZoom={1.5}
          minZoom={0.5}
        >
          <Background
            variant={BackgroundVariant.Dots}
            gap={24}
            size={1}
            color="rgba(255,255,255,0.05)"
          />
        </ReactFlow>
      </div>

      <SubmitModal 
        isOpen={modalOpen}
        onClose={() => setModalOpen(false)}
        problem={targetProblem}
        onSubmit={handleSubmitVerification}
        isSubmitting={isSubmitting}
        isSuccess={isSuccess}
        error={submitError}
      />

      <div 
        className={`fixed top-[64px] right-0 h-[calc(100vh-64px)] w-96 bg-slate-900/95 backdrop-blur-md border-l border-slate-800 shadow-2xl transform transition-transform duration-300 ease-in-out z-30 flex flex-col ${
          selectedTopic ? 'translate-x-0' : 'translate-x-full'
        }`}
      >
        {selectedTopic && (
          <>
            <div className="p-6 border-b border-slate-800 flex justify-between items-center bg-slate-900">
              <div>
                <h2 className="text-xl font-bold text-gray-100">{selectedTopic.label}</h2>
                <p className="text-xs text-slate-400 mt-1">Recommended Practice</p>
              </div>
              <button 
                onClick={() => setSelectedTopic(null)}
                className="text-slate-400 hover:text-white transition-colors p-1"
              >
                ✕
              </button>
            </div>

            <div className="p-4 grid grid-cols-3 gap-2 border-b border-slate-800">
              {(Object.keys(DIFFICULTY_CONFIG) as Difficulty[]).map((level) => (
                <button
                  key={level}
                  onClick={() => setDifficulty(level)}
                  className={`py-2 px-1 text-xs font-semibold rounded transition-all border ${
                    difficulty === level 
                      ? `${DIFFICULTY_CONFIG[level].color} ${DIFFICULTY_CONFIG[level].border} bg-slate-800` 
                      : 'text-slate-500 border-transparent hover:bg-slate-800'
                  }`}
                >
                  {DIFFICULTY_CONFIG[level].label}
                </button>
              ))}
            </div>

            <div className="flex-1 overflow-y-auto p-4 space-y-3">
              {loadingProblems ? (
                 <div className="text-center py-10 text-slate-500 text-sm animate-pulse">Analyzing...</div>
              ) : problems.length > 0 ? (
                problems.map((prob) => (
                  <div 
                    key={prob.id}
                    className="relative group block rounded-lg border border-slate-800 bg-slate-900/50 transition-all hover:bg-slate-800 hover:border-slate-700"
                  >
                    <a 
                      href={getProblemLink(prob.id)}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="block p-4 pr-12" 
                    >
                      <div className="flex justify-between items-start mb-2">
                        <span className="font-semibold text-gray-200 group-hover:text-sky-400 transition-colors">
                          {prob.id} - {prob.name}
                        </span>
                        <span className={`text-xs font-bold px-2 py-0.5 rounded bg-slate-950 ${
                          prob.rating >= 1600 ? 'text-yellow-500' :
                          prob.rating >= 1400 ? 'text-cyan-400' :
                          prob.rating >= 1200 ? 'text-green-400' : 'text-gray-400'
                        }`}>
                          {prob.rating}
                        </span>
                      </div>
                      <div className="flex flex-wrap gap-1.5 mt-2">
                        {prob.tags.slice(0, 3).map(tag => (
                          <span key={tag} className="text-[10px] text-slate-400 bg-slate-800 px-1.5 py-0.5 rounded">
                            {tag}
                          </span>
                        ))}
                      </div>
                    </a>

                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleOpenVerify(prob);
                      }}
                      className="absolute right-3 top-1/2 -translate-y-1/2 w-8 h-8 flex items-center justify-center rounded-full bg-slate-800 border border-slate-600 text-slate-400 hover:bg-emerald-500 hover:text-white hover:border-emerald-400 hover:shadow-lg hover:shadow-emerald-500/20 transition-all z-10"
                      title="Verify & Mark Complete"
                    >
                      ✓
                    </button>
                  </div>
                ))
              ) : (
                <div className="text-center py-10 text-slate-500 text-sm">
                  No problems found.
                </div>
              )}
            </div>
            
            <div className="p-3 border-t border-slate-800 text-[10px] text-center text-slate-600">
              Problems tailored to your current rating
            </div>
          </>
        )}
      </div>
    </div>
  );
};

export default Training;