import { useState, useEffect } from 'react';
import { Routes, Route, Navigate, useNavigate, useLocation, NavLink } from 'react-router-dom';
import Roadmap from './Roadmap';
import LandingPage from './LandingPage';
import Stats from './Stats';

function App() {
  const [handle, setHandle] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    const storedHandle = localStorage.getItem('cf_handle');
    setHandle(storedHandle);
    setLoading(false);
  }, []);

  useEffect(() => {
    if (!loading && !handle && location.pathname !== '/') {
      navigate('/');
    }
    if (!loading && handle && location.pathname === '/') {
      navigate('/skill-tree');
    }
  }, [handle, loading, navigate, location]);

  const handleLoginSuccess = (newHandle: string) => {
    localStorage.setItem('cf_handle', newHandle);
    setHandle(newHandle);
    navigate('/skill-tree');
  };

  const handleLogout = () => {
    localStorage.removeItem('cf_handle');
    setHandle(null);
    navigate('/');
  };

  if (loading) return null;

  return (
    <div className="h-screen w-screen overflow-hidden flex flex-col"
      style={{ background: 'linear-gradient(180deg, #0b1220 0%, #0e1628 100%)' }}
    >
      {handle && (
        <nav className="px-6 py-3 border-b border-slate-800 bg-slate-900/50 backdrop-blur-sm text-gray-200 font-mono text-sm tracking-wider flex justify-between items-center z-20">
          <div className="flex items-center gap-8">
            <div>
              <span className="text-sky-400">CF</span>-MASTERY
            </div>
            
            <div className="flex gap-4">
              <NavLink 
                to="/skill-tree" 
                className={({ isActive }) => 
                  `hover:text-sky-400 transition-colors ${isActive ? 'text-sky-400 font-bold' : 'text-slate-400'}`
                }
              >
                SKILL-TREE
              </NavLink>
              <NavLink 
                to="/stats" 
                className={({ isActive }) => 
                  `hover:text-sky-400 transition-colors ${isActive ? 'text-sky-400 font-bold' : 'text-slate-400'}`
                }
              >
                STATS
              </NavLink>
            </div>
          </div>
          
          <div className="flex items-center gap-4">
            <span className="text-slate-400 text-xs">
              Logged in as <span className="text-white font-semibold">{handle}</span>
            </span>
            <button 
              onClick={handleLogout}
              className="text-xs text-red-400 hover:text-red-300 transition-colors border border-red-400/20 px-2 py-1 rounded hover:bg-red-400/10"
            >
              LOGOUT
            </button>
          </div>
        </nav>
      )}

      <main className="flex-1 overflow-hidden relative">
        <Routes>
          <Route path="/" element={<LandingPage onSuccess={handleLoginSuccess} />} />
          <Route path="/skill-tree" element={handle ? <Roadmap /> : <Navigate to="/" />} />
          <Route path="/stats" element={handle ? <Stats /> : <Navigate to="/" />} />
          <Route path="*" element={<Navigate to="/" />} />
        </Routes>
      </main>
    </div>
  );
}

export default App;