import { useState, useEffect } from 'react';

interface RecentSolve {
  id: string;
  name: string;
  rating: number;
  tags: string[];
  solvedAt: string;
}

const RecentActivity = () => {
  const [solves, setSolves] = useState<RecentSolve[]>([]);
  const [unsolved, setUnsolved] = useState<RecentSolve[]>([]);
  const [loading, setLoading] = useState(true);
  const handle = sessionStorage.getItem('cf_handle');

  useEffect(() => {
    if (!handle) return;

    const fetchRecent = async () => {
      setLoading(true)
      try {
        const [resSolved, resUnsolved] = await Promise.all([
          fetch(`http://localhost:8080/api/recent/solved/${handle}`),
          fetch(`http://localhost:8080/api/recent/unsolved/${handle}`),
        ]);

        const solvedData = resSolved.ok ? await resSolved.json() : [];
        const unsolvedData = resUnsolved.ok ? await resUnsolved.json() : [];

        setSolves(solvedData || []);
        setUnsolved(unsolvedData || []);
      } catch (err) {
        console.error("Error fetching recent activity:", err);
      } finally {
        setLoading(false);
      }
    };

    fetchRecent();
  }, []);

  const timeAgo = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();

    const seconds = Math.max(0, Math.floor((now.getTime() - date.getTime()) / 1000));
    
    if (seconds < 60) return "Just Now";
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return `${minutes}m ago`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}h ago`;
    const days = Math.floor(hours / 24);
    return `${days}d ago`;
  };

  const getLink = (id: string) => {
    const match = id.match(/^(\d+)(.+)$/);
    if (match) {
      return `https://codeforces.com/problemset/problem/${match[1]}/${match[2]}`;
    }
    return `https://codeforces.com/problemset/problem/${id}`;
  };

  const getRatingColor = (rating: number) => {
    if (rating < 1000) return "text-gray-400 border-gray-500/30 bg-gray-500/10";
    if (rating < 1400) return "text-green-400 border-green-500/30 bg-green-500/10";
    if (rating < 1600) return "text-cyan-400 border-cyan-500/30 bg-cyan-500/10";
    if (rating < 1800) return "text-blue-400 border-blue-500/30 bg-blue-500/10";
    if (rating < 2000) return "text-violet-400 border-violet-500/30 bg-violet-500/10";
    if (rating < 2400) return "text-orange-400 border-orange-500/30 bg-orange-500/10";
    return "text-red-400 border-red-500/30 bg-red-500/10";
  };

  if (loading) {
    return (
      <div className="h-full flex items-center justify-center text-slate-500 font-mono animate-pulse">
        LOADING...
      </div>
    );
  }

  if (solves.length === 0 && unsolved.length === 0) {
    return (
      <div className="text-center py-8 border border-slate-800 rounded-xl bg-slate-900/30">
        <p className="text-slate-500 text-sm">No recent activity detected.</p>
      </div>
    );
  }

  const ProblemRow = ({ item, isUnsolved = false }: { item: RecentSolve, isUnsolved?: boolean }) => (
    <a 
      href={getLink(item.id)}
      target="_blank"
      rel="noreferrer"
      className={`group flex justify-between items-center border rounded-lg px-4 py-3 transition-all duration-200 
        ${isUnsolved 
          ? 'bg-red-900/10 hover:bg-red-900/20 border-red-900/30 hover:border-red-700/50' 
          : 'bg-slate-800/20 hover:bg-slate-800/50 border-slate-800 hover:border-slate-700'
        }`}
    >
      <div className="flex items-center gap-4 min-w-0">
        <span className={`font-mono text-sm font-bold shrink-0 min-w-[3rem] 
          ${isUnsolved ? 'text-red-400/80 group-hover:text-red-400' : 'text-sky-500/80 group-hover:text-sky-400'}`}>
          {item.id}
        </span>
        <span className="text-sm text-slate-400 group-hover:text-slate-200 truncate font-medium">
          {item.name}
        </span>
      </div>

      <div className="flex items-center gap-4 pl-4 shrink-0">
        {item.rating > 0 && (
          <span className={`text-[10px] px-2 py-0.5 rounded border font-mono font-medium ${getRatingColor(item.rating)}`}>
            {item.rating}
          </span>
        )}
        <span className="text-xs text-slate-600 font-mono w-16 text-right">
          {timeAgo(item.solvedAt)}
        </span>
      </div>
    </a>
  );

  return (
    <div className="relative flex flex-col bg-slate-900/50 border border-slate-800 rounded-xl h-[100dvh] overflow-hidden">
        <div className="min-h-0 flex-1 overflow-y-auto p-4 pb-20 space-y-8">
        {solves.length > 0 && (
            <div>
            <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-bold text-white flex items-center gap-2">
                <span className="w-2 h-2 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]"></span>
                Recent Solves
                </h3>
                <span className="text-xs text-slate-500 font-mono">LIVE FEED</span>
            </div>
            <div className="space-y-2">
                {solves.map((s) => <ProblemRow key={s.id} item={s} />)}
            </div>
            </div>
        )}

        {unsolved.length > 0 && (
            <div>
            <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-bold text-white flex items-center gap-2">
                <span className="w-2 h-2 rounded-full bg-amber-500 shadow-[0_0_8px_rgba(245,158,11,0.5)]"></span>
                Unresolved
                </h3>
                <span className="text-xs text-amber-500/50 font-mono">ATTEMPTED</span>
            </div>
            <div className="space-y-2">
                {unsolved.map((u) => (
                <ProblemRow key={u.id} item={u} isUnsolved />
                ))}
            </div>
            </div>
        )}
        </div>
    </div>
    );
};

export default RecentActivity;