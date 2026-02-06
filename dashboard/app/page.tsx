'use client';

import React, { useState } from 'react';
import { Shield, Zap, Search, Activity, ChevronRight } from 'lucide-react';
import JobList from '@/components/JobList';
import SubmitForm from '@/components/SubmitForm';
import ThoughtStream from '@/components/ThoughtStream';

export default function Home() {
  const [activeJobId, setActiveJobId] = useState<string | null>(null);

  const stats = [
    { label: "Critical Fixed", value: "12", icon: Shield, color: "text-rose-400" },
    { label: "Active Jobs", value: "2", icon: Zap, color: "text-indigo-400" },
    { label: "Fix Rate", value: "94%", icon: Activity, color: "text-emerald-400" },
    { label: "Avg Time", value: "2.4m", icon: Search, color: "text-slate-400" },
  ];

  return (
    <div className="p-8 max-w-7xl mx-auto space-y-8">
      {/* Page Header */}
      <header className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-white tracking-tight">Security Command Center</h1>
          <p className="text-slate-400 mt-1">Monitor and orchestrate autonomous security remediations.</p>
        </div>
        <div className="flex gap-2">
           <button className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-sm font-medium rounded-lg border border-slate-700 transition-colors">
            View Analytics
           </button>
           <button className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-sm font-medium rounded-lg shadow-lg shadow-indigo-500/20 transition-all">
            New Workspace
           </button>
        </div>
      </header>

      {/* Stats Row */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat) => (
          <div key={stat.label} className="p-6 bg-slate-900/50 border border-slate-800 rounded-xl hover:border-slate-700 transition-colors group">
            <div className="flex items-center gap-4">
              <div className={`p-3 rounded-lg bg-slate-800 ${stat.color} group-hover:scale-110 transition-transform`}>
                <stat.icon className="w-5 h-5" />
              </div>
              <div>
                <p className="text-xs font-semibold uppercase tracking-wider text-slate-500">{stat.label}</p>
                <p className="text-2xl font-bold text-white">{stat.value}</p>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Main Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
        {/* Left Column: Form & List */}
        <div className="lg:col-span-5 space-y-8">
          <section className="bg-slate-900 border border-slate-800 rounded-2xl overflow-hidden shadow-xl">
            <div className="p-4 border-b border-slate-800 bg-slate-800/50 flex items-center justify-between">
              <h2 className="text-sm font-bold uppercase tracking-widest text-slate-300">Submit Vulnerability Scan</h2>
              <Shield className="w-4 h-4 text-indigo-400" />
            </div>
            <div className="p-6">
              <SubmitForm onJobCreated={(id) => setActiveJobId(id)} />
            </div>
          </section>

          <section className="bg-slate-900 border border-slate-800 rounded-2xl overflow-hidden shadow-xl">
            <div className="p-4 border-b border-slate-800 bg-slate-800/50 flex items-center justify-between">
              <h2 className="text-sm font-bold uppercase tracking-widest text-slate-300">Remediation Logs</h2>
              <Activity className="w-4 h-4 text-emerald-400" />
            </div>
            <JobList onSelectJob={setActiveJobId} activeJobId={activeJobId} />
          </section>
        </div>

        {/* Right Column: Thought Stream */}
        <div className="lg:col-span-7 flex flex-col">
          <section className="flex-1 bg-slate-900 border border-slate-800 rounded-2xl overflow-hidden shadow-xl flex flex-col">
            <div className="p-4 border-b border-slate-800 bg-slate-800/50 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <h2 className="text-sm font-bold uppercase tracking-widest text-slate-300 font-mono">Agent Reasoning Pipeline</h2>
                {activeJobId && (
                  <span className="px-2 py-0.5 rounded-full bg-indigo-500/10 text-indigo-400 text-[10px] font-mono border border-indigo-500/20">
                    ID: {activeJobId.substring(0, 8)}...
                  </span>
                )}
              </div>
              <Zap className="w-4 h-4 text-amber-400" />
            </div>
            <div className="flex-1 min-h-[500px]">
              <ThoughtStream jobId={activeJobId} />
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}
