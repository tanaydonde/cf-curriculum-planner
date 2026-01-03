import { useState } from 'react';

interface LandingPageProps {
  onSuccess: (handle: string) => void;
}

const LandingPage = ({ onSuccess }: LandingPageProps) => {
  const [handle, setHandle] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!handle.trim()) return;

    setLoading(true);
    setError('');

    try {
      const res = await fetch(`http://localhost:8080/api/sync/${handle}`, {
        method: 'POST',
      });

      if (!res.ok) {
        const errorText = await res.text();
        throw new Error(errorText || 'Failed to sync handle');
      }

      onSuccess(handle);
      
    } catch (err: any) {
      console.error(err);
      setError(err.message.includes('not found') 
        ? `Handle '${handle}' not found` 
        : 'Error syncing handle. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex flex-col items-center justify-center h-full w-full text-gray-200">
      <div className="w-full max-w-md p-8 bg-slate-900/50 border border-slate-800 rounded-xl shadow-2xl backdrop-blur-sm">
        
        <div className="text-center mb-10">
          <h1 className="text-4xl font-bold tracking-tight text-white mb-2">
            <span className="text-sky-400">CF</span> Mastery
          </h1>
          <p className="text-slate-400 text-sm">
            Enter your handle to generate your curriculum
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="space-y-2">
            <label htmlFor="handle" className="text-xs font-mono text-slate-500 uppercase tracking-wider ml-1">
              Codeforces Handle
            </label>
            <input
              id="handle"
              type="text"
              value={handle}
              onChange={(e) => setHandle(e.target.value)}
              placeholder="tourist"
              disabled={loading}
              className="w-full bg-slate-950 border border-slate-700 rounded-lg px-4 py-3 text-white placeholder-slate-600 focus:outline-none focus:ring-2 focus:ring-sky-500/50 focus:border-sky-500 transition-all"
            />
          </div>

          {error && (
            <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-lg">
              <p className="text-red-400 text-sm text-center font-medium">
                {error}
              </p>
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            className={`w-full py-3.5 px-4 rounded-lg font-medium text-sm transition-all duration-200 
              ${loading 
                ? 'bg-slate-800 text-slate-500 cursor-not-allowed' 
                : 'bg-sky-600 hover:bg-sky-500 text-white shadow-lg shadow-sky-900/20 active:scale-[0.98]'
              }`}
          >
            {loading ? (
              <span className="flex items-center justify-center gap-2">
                <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                Syncing History...
              </span>
            ) : (
              'Initialize Roadmap'
            )}
          </button>
        </form>
      </div>
    </div>
  );
};

export default LandingPage;