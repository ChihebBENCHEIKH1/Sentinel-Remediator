import React, { useEffect, useState } from 'react';
import { Activity, Shield, ChevronRight } from 'lucide-react';
import clsx from 'clsx';

interface Job {
  id: string;
  status: string;
  progress: number;
  fixed_count: number;
  total_count: number;
  created_at: string;
}

interface JobListProps {
  onSelectJob: (id: string) => void;
  activeJobId: string | null;
}

const getStatusStyles = (status: string) => {
  switch (status) {
    case 'PENDING': return 'bg-slate-800 text-slate-400';
    case 'REASONING': return 'bg-indigo-500/10 text-indigo-400 border border-indigo-500/20';
    case 'APPLYING_FIX': return 'bg-blue-500/10 text-blue-400 border border-blue-500/20';
    case 'BUILDING': return 'bg-amber-500/10 text-amber-400 border border-amber-500/20';
    case 'SUCCESS': return 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20';
    case 'FAILED': return 'bg-rose-500/10 text-rose-400 border border-rose-500/20';
    default: return 'bg-slate-800 text-slate-400';
  }
};

const JobList: React.FC<JobListProps> = ({ onSelectJob, activeJobId }) => {
  const [jobs, setJobs] = useState<Job[]>([]);

  useEffect(() => {
    const fetchJobs = async () => {
      try {
        const response = await fetch('http://localhost:8080/api/jobs');
        if (response.ok) {
          const data = await response.json();
          setJobs(data);
        }
      } catch (err) {
        console.error('Failed to fetch jobs:', err);
      }
    };

    fetchJobs();
    const interval = setInterval(fetchJobs, 3000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="divide-y divide-slate-800 max-h-[400px] overflow-y-auto scrollbar-thin">
      {jobs.length === 0 ? (
        <div className="p-12 text-center">
          <div className="inline-flex p-4 rounded-full bg-slate-800 text-slate-500 mb-4">
             <Shield className="w-8 h-8 opacity-20" />
          </div>
          <p className="text-sm text-slate-500">No remediation history found.</p>
        </div>
      ) : (
        jobs.map((job) => (
          <button
            key={job.id}
            onClick={() => onSelectJob(job.id)}
            className={clsx(
              "w-full text-left p-4 transition-all hover:bg-slate-800/50 flex items-center justify-between group outline-none",
              activeJobId === job.id && "bg-indigo-600/5 border-l-2 border-indigo-500"
            )}
          >
            <div className="flex-1 min-w-0 mr-4">
              <div className="flex items-center gap-2 mb-1">
                <span className="text-xs font-mono font-bold text-slate-300">
                  {job.id.substring(0, 8)}
                </span>
                <span className={clsx(
                  "text-[10px] px-1.5 py-0.5 rounded font-bold uppercase",
                  getStatusStyles(job.status)
                )}>
                  {job.status}
                </span>
              </div>
              <div className="flex items-center gap-3 text-[10px] text-slate-500">
                <span className="flex items-center gap-1">
                  <Activity className="w-3 h-3" />
                   {Math.round(job.progress * 100)}%
                </span>
                <span>•</span>
                <span>{job.fixed_count}/{job.total_count} Fixed</span>
                <span>•</span>
                <span>{new Date(job.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
              </div>
            </div>
            <ChevronRight className={clsx(
              "w-4 h-4 text-slate-600 transition-transform group-hover:translate-x-1",
              activeJobId === job.id && "text-indigo-500"
            )} />
          </button>
        ))
      )}
    </div>
  );
};

export default JobList;
