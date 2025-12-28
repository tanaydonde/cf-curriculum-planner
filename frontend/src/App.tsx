import Roadmap from './Roadmap';

function App() {
  return (
    <div className="h-screen w-screen overflow-hidden"
      style={{ background: 'linear-gradient(180deg, #0b1220 0%, #0e1628 100%)' }}
    >
      <nav className="px-6 py-3 border-b border-slate-800 bg-transparent text-gray-200 font-mono text-sm tracking-wider">
        <span className="text-sky-400">CF</span>-CURRICULUM-PLANNER // v1.0
      </nav>
      <main className="h-full w-full">
        <Roadmap />
      </main>
    </div>
  );
}

export default App;