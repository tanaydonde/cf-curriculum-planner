import { useEffect, useState } from 'react';
import {
  Radar,
  RadarChart,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
  ResponsiveContainer,
  Tooltip,
} from 'recharts';

interface MasteryResult {
  current: number;
  peak: number;
}

interface TopicData {
  topic: string;
  current: number;
  peak: number;
  decay: number;
  fullMark: number; 
}

const LABEL_MAP: Record<string, string> = {
    "dynamic programming": "DP",
    "data structures": "DS",
    "meet-in-the-middle": "MitM",
};

const lerp = (start: number, end: number, factor: number) => start + (end - start) * factor;

const lerpColor = (c1: number[], c2: number[], factor: number) => {
  const r = Math.round(lerp(c1[0], c2[0], factor));
  const g = Math.round(lerp(c1[1], c2[1], factor));
  const b = Math.round(lerp(c1[2], c2[2], factor));
  return `rgb(${r}, ${g}, ${b})`;
};


const getDecayColorStyle = (decay: number) => {
  const C_GOOD = [14, 165, 233];
  const C_OK = [34, 211, 238];
  const C_WARN = [251, 146, 60]; 
  const C_CRIT = [244, 63, 94]; 

  if (decay <= 150) {
    return lerpColor(C_GOOD, C_OK, decay / 150);
  }
  else if (decay <= 300) {
    return lerpColor(C_OK, C_WARN, (decay - 150) / 150);
  }
  else {
    const factor = Math.min((decay - 300) / 200, 1); 
    return lerpColor(C_WARN, C_CRIT, factor);
  }
};

const getRatingColor = (rating: number) => {
  if (rating < 1200) return "text-gray-400";
  if (rating < 1400) return "text-green-500";
  if (rating < 1600) return "text-cyan-400";
  if (rating < 1900) return "text-blue-500";
  if (rating < 2100) return "text-purple-500";
  if (rating < 2300) return "text-orange-400";
  if (rating < 2400) return "text-orange-500";
  return "text-red-500";
};

const calculateWeightedRating = (values: number[]) => {
  let num = 0;
  let den = 0;

  for (const v of values) {
    if (v <= 0) continue;
    const v4 = Math.pow(v, 4);
    const v3 = Math.pow(v, 3);

    num += v4;
    den += v3;
  }

  if (den === 0) return 0;
  return Math.round(num / den);
};

const formatLabel = (slug: string) => LABEL_MAP[slug] || slug.charAt(0).toUpperCase() + slug.slice(1);

const Stats = () => {
  const [data, setData] = useState<TopicData[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStats = async () => {
      const handle = localStorage.getItem('cf_handle');
      if (!handle) return;

      try {
        const response = await fetch(`http://localhost:8080/api/stats/${handle}`); 
        const json: Record<string, MasteryResult> = await response.json();

        const chartData: TopicData[] = Object.entries(json).map(([key, value]) => ({
            topic: formatLabel(key),
            current: Math.round(value.current),
            peak: Math.round(value.peak),
            decay: Math.round(value.peak - value.current),
            fullMark: 2000, 
        }));

        chartData.sort((a, b) => a.topic.localeCompare(b.topic));
        
        setData(chartData);
      } catch (err) {
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
  }, []);

  const currentRating = calculateWeightedRating(data.map(d => d.current));
  const peakRating = calculateWeightedRating(data.map(d => d.peak));

  if (loading) return <div className="text-white p-10">Loading Stats...</div>;

  return (
    <div className="h-full w-full overflow-y-auto bg-slate-900/50 p-8 scrollbar-thin scrollbar-thumb-slate-700">
      <div className="max-w-7xl mx-auto space-y-8">
        
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          
          <div className="bg-slate-900 border border-slate-800 p-6 rounded-xl shadow-lg relative overflow-hidden">
            <div className={`absolute top-0 right-0 w-32 h-32 opacity-5 blur-3xl rounded-full -mr-10 -mt-10 pointer-events-none bg-current ${getRatingColor(currentRating)}`}></div>
            
            <h3 className="text-slate-400 text-xs font-mono uppercase tracking-wider mb-2">Effective Rating</h3>
            <div className={`text-4xl font-bold ${getRatingColor(currentRating)}`}>
              {currentRating}
            </div>
            <div className="text-slate-500 text-xs mt-1">
              Weighted by your strongest topics
            </div>
          </div>

          <div className="bg-slate-900 border border-slate-800 p-6 rounded-xl shadow-lg">
            <h3 className="text-slate-400 text-xs font-mono uppercase tracking-wider mb-2">Peak Potential</h3>
            <div className={`text-4xl font-bold ${getRatingColor(peakRating)}`}>
              {peakRating}
            </div>
            <div className="text-slate-500 text-xs mt-1">
              Theoretical max if rust is removed
            </div>
          </div>

          <StatCard 
            title="Decay Factor" 
            value={`-${peakRating - currentRating}`} 
            subtitle="Rating points lost to inactivity"
            type="decay"
          />
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 h-[500px]">
          
          <div className="bg-slate-900 border border-slate-800 rounded-xl p-4 shadow-xl flex flex-col">
            <h3 className="text-slate-400 text-sm font-mono uppercase tracking-wider mb-4 ml-2">Skill Fingerprint</h3>
            <div className="flex-1 min-h-0">
              <ResponsiveContainer width="100%" height="100%">
                <RadarChart cx="50%" cy="50%" outerRadius="70%" data={data}>
                  <PolarGrid stroke="#334155" />
                  <PolarAngleAxis 
                    dataKey="topic" 
                    tick={{ fill: '#94a3b8', fontSize: 10 }} 
                  />
                  <PolarRadiusAxis angle={30} domain={[0, 'auto']} tick={false} axisLine={false} />
                  
                  <Radar
                    name="Peak"
                    dataKey="peak"
                    stroke="#0ea5e9"
                    strokeWidth={2}
                    strokeDasharray="4 4"
                    fill="#0ea5e9"
                    fillOpacity={0.1}
                  />
                  
                  <Radar
                    name="Current"
                    dataKey="current"
                    stroke="#38bdf8"
                    strokeWidth={2}
                    fill="#38bdf8"
                    fillOpacity={0.6}
                  />
                  <Tooltip 
                    contentStyle={{ backgroundColor: '#0f172a', borderColor: '#1e293b', color: '#f1f5f9' }}
                    itemStyle={{ color: '#e2e8f0' }}
                  />
                </RadarChart>
              </ResponsiveContainer>
            </div>
          </div>

          <div className="bg-slate-900 border border-slate-800 rounded-xl p-4 shadow-xl flex flex-col">
            <h3 className="text-slate-400 text-sm font-mono uppercase tracking-wider mb-4 ml-2">Highest Decay (Needs Practice)</h3>
            <div className="flex-1 min-h-0 overflow-y-auto pr-2">
              {[...data]
                .sort((a, b) => b.decay - a.decay)
                .map((item) => {
                  const colorStyle = getDecayColorStyle(item.decay);
                  return (
                    <div key={item.topic} className="mb-4 last:mb-0 group">
                      <div className="flex justify-between text-sm mb-1">
                        <span className="font-medium text-slate-200">{item.topic}</span>
                        <span className="font-mono text-xs">
                          <span className="text-sky-400">{item.current}</span>
                          <span className="text-slate-600"> / </span>
                          <span className="text-slate-400">{item.peak}</span>
                        </span>
                      </div>
                      <div className="h-2 w-full bg-slate-800 rounded-full overflow-hidden relative">
                        <div 
                          className="absolute top-0 left-0 h-full opacity-20" 
                          style={{ 
                            width: `${Math.min((item.peak / 3000) * 100, 100)}%`,
                            backgroundColor: colorStyle 
                          }} 
                        />
                        <div 
                          className="absolute top-0 left-0 h-full rounded-full transition-all duration-500"
                          style={{ 
                            width: `${Math.min((item.current / 3000) * 100, 100)}%`,
                            backgroundColor: colorStyle
                          }}
                        />
                      </div>
                      {item.decay > 10 && (
                        <div 
                          className="text-[10px] mt-1 text-right opacity-0 group-hover:opacity-100 transition-opacity"
                          style={{ color: colorStyle }}
                        >
                          -{item.decay} pts due to decay
                        </div>
                      )}
                    </div>
                  );
                })}
            </div>
          </div>

        </div>
      </div>
    </div>
  );
};

const StatCard = ({ title, value, subtitle, type = 'neutral' }: any) => {
  const colors = {
    neutral: "text-white",
    peak: "text-sky-400",
    decay: "text-orange-400"
  };

  return (
    <div className="bg-slate-900 border border-slate-800 p-6 rounded-xl shadow-lg">
      <h3 className="text-slate-400 text-xs font-mono uppercase tracking-wider mb-2">{title}</h3>
      <div className={`text-3xl font-bold ${colors[type as keyof typeof colors]}`}>{value}</div>
      <div className="text-slate-500 text-xs mt-1">{subtitle}</div>
    </div>
  );
};

export default Stats;