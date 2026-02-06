import React, { useEffect, useState, useRef } from 'react';
import { 
  Zap, 
  Terminal, 
  Shield, 
  AlertCircle, 
  CheckCircle2, 
  RefreshCw,
  Search,
  GitBranch,
  Cpu
} from 'lucide-react';
import clsx from 'clsx';

interface Thought {
  timestamp: string;
  type: string;
  content: string;
  iteration?: number;
}

interface ThoughtStreamProps {
  jobId: string | null;
}

const getThoughtIcon = (type: string) => {
  switch (type) {
    case 'thought': return <Cpu className="w-4 h-4 text-indigo-400" />;
    case 'action': return <Terminal className="w-4 h-4 text-emerald-400" />;
    case 'observation': return <Search className="w-4 h-4 text-blue-400" />;
    case 'error': return <AlertCircle className="w-4 h-4 text-rose-400" />;
    case 'success': return <CheckCircle2 className="w-4 h-4 text-emerald-400" />;
    case 'git': return <GitBranch className="w-4 h-4 text-purple-400" />;
    default: return <RefreshCw className="w-4 h-4 text-slate-400" />;
  }
};

const ThoughtStream: React.FC<ThoughtStreamProps> = ({ jobId }) => {
  const [thoughts, setThoughts] = useState<Thought[]>([]);
  const [status, setStatus] = useState<string>('');
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!jobId) {
      setThoughts([]);
      setStatus('');
      return;
    }

    const eventSource = new EventSource(`http://localhost:8080/api/jobs/${jobId}/stream`);

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.id) { // Initial data or status update
          setStatus(data.status);
          if (data.thought_trace) {
            setThoughts(data.thought_trace);
          }
        } else if (data.content) { // Single thought step
          setThoughts((prev) => [...prev, data]);
        }
      } catch (err) {
        console.error('Failed to parse event data:', err);
      }
    };

    eventSource.onerror = () => {
      eventSource.close();
    };

    return () => {
      eventSource.close();
    };
  }, [jobId]);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [thoughts]);

  return (
    <div className="flex flex-col h-full bg-slate-900 min-h-[500px]">
      <div className="flex-1 overflow-y-auto p-0 scrollbar-thin">
        {!jobId ? (
          <div className="h-full flex flex-col items-center justify-center text-slate-600 p-8 text-center space-y-4">
            <div className="w-16 h-16 rounded-full border-2 border-dashed border-slate-800 flex items-center justify-center">
               <Shield className="w-8 h-8 opacity-20" />
            </div>
            <div>
              <p className="text-sm font-bold uppercase tracking-widest text-slate-500">Pipeline Offline</p>
              <p className="text-xs mt-1">Select a job from the log to view live reasoning.</p>
            </div>
          </div>
        ) : thoughts.length === 0 ? (
          <div className="p-6 space-y-4">
             <div className="animate-pulse flex space-x-4">
                <div className="flex-1 space-y-6 py-1">
                  <div className="h-2 bg-slate-800 rounded"></div>
                  <div className="space-y-3">
                    <div className="grid grid-cols-3 gap-4">
                      <div className="h-2 bg-slate-800 rounded col-span-2"></div>
                      <div className="h-2 bg-slate-800 rounded col-span-1"></div>
                    </div>
                    <div className="h-2 bg-slate-800 rounded"></div>
                  </div>
                </div>
              </div>
          </div>
        ) : (
          <div className="p-0">
            {thoughts.map((thought, idx) => (
              <div key={idx} className="border-b border-slate-800 last:border-0 hover:bg-slate-800/20 transition-colors">
                <div className="p-4 flex gap-4">
                  <div className="mt-1">
                    {getThoughtIcon(thought.type)}
                  </div>
                  <div className="flex-1 space-y-2">
                    <div className="flex items-center justify-between">
                       <span className="text-[10px] font-mono text-slate-500 font-bold uppercase tracking-widest">
                         [{thought.type}] {new Date(thought.timestamp).toLocaleTimeString()}
                       </span>
                       <span className="text-[10px] font-mono text-slate-600">
                         #0{thought.iteration || 0}
                       </span>
                    </div>
                    <div className="text-sm font-medium text-slate-300 leading-relaxed">
                      {thought.content}
                    </div>
                    
                    {/* Tool Call Rendering */}
                    {thought.type === 'action' && thought.content.includes('{') && (
                      <div className="mt-3 bg-slate-950/80 border border-slate-700/50 rounded-lg p-3 overflow-hidden group">
                        <div className="flex items-center gap-2 mb-2 pb-2 border-b border-slate-800">
                          <Terminal className="w-3 h-3 text-emerald-400" />
                          <span className="text-[10px] font-mono text-emerald-400">TOOL_INVOCATION</span>
                        </div>
                        <pre className="text-xs font-mono text-slate-400 overflow-x-auto selection:bg-indigo-500/20">
                          {thought.content}
                        </pre>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            ))}
            
            {status !== 'SUCCESS' && status !== 'FAILED' && status !== 'CANCELLED' && (
              <div className="p-6 flex items-center gap-4 animate-pulse border-t border-slate-800">
                <div className="animate-spin text-indigo-500">
                  <RefreshCw className="w-4 h-4" />
                </div>
                <span className="text-xs font-mono text-indigo-400 uppercase tracking-widest font-bold">Agent processing...</span>
              </div>
            )}
          </div>
        )}
        <div ref={scrollRef} />
      </div>
    </div>
  );
};

export default ThoughtStream;
