import React from 'react';
import { Shield, LayoutDashboard, Database, Settings, LogOut, Terminal } from 'lucide-react';

interface SidebarProps {
  onNavChange?: (page: string) => void;
}

export const Sidebar: React.FC<SidebarProps> = ({ onNavChange }) => {
  return (
    <aside className="w-64 border-r border-slate-800 bg-[#020617] flex flex-col h-screen sticky top-0">
      <div className="p-6 flex items-center gap-3">
        <div className="bg-indigo-600 p-2 rounded-lg">
          <Shield className="w-6 h-6 text-white" />
        </div>
        <div>
          <h1 className="text-xl font-bold tracking-tight text-white">Sentinel</h1>
          <p className="text-[10px] uppercase tracking-widest text-slate-500 font-semibold">Remediator</p>
        </div>
      </div>

      <nav className="flex-1 px-4 py-4">
        <ul className="space-y-2">
          <li>
            <button className="w-full flex items-center gap-3 px-3 py-2 text-sm font-medium transition-colors rounded-lg bg-indigo-600/10 text-indigo-400 group">
              <LayoutDashboard className="w-4 h-4" />
              Overview
            </button>
          </li>
          <li>
            <button className="w-full flex items-center gap-3 px-3 py-2 text-sm font-medium transition-colors text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg group">
              <Terminal className="w-4 h-4" />
              Remediation Jobs
            </button>
          </li>
          <li>
            <button className="w-full flex items-center gap-3 px-3 py-2 text-sm font-medium transition-colors text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg group">
              <Database className="w-4 h-4" />
              Fix Memory
            </button>
          </li>
          <li>
            <button className="w-full flex items-center gap-3 px-3 py-2 text-sm font-medium transition-colors text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg group">
              <Settings className="w-4 h-4" />
              Settings
            </button>
          </li>
        </ul>
      </nav>

      <div className="p-4 border-t border-slate-800">
        <div className="flex items-center gap-3 px-2 py-2 mb-4">
          <div className="w-8 h-8 rounded-full bg-gradient-to-tr from-indigo-500 to-purple-500 flex items-center justify-center text-xs font-bold">
            JD
          </div>
          <div className="flex-1 overflow-hidden">
            <p className="text-sm font-medium text-white truncate">Admin User</p>
            <p className="text-xs text-slate-500 truncate">admin@sentinel.io</p>
          </div>
        </div>
        <button className="w-full flex items-center gap-3 px-3 py-2 text-sm font-medium transition-colors text-slate-400 hover:text-rose-400 rounded-lg group">
          <LogOut className="w-4 h-4" />
          Sign Out
        </button>
      </div>
    </aside>
  );
};
