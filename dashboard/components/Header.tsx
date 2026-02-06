import { Shield } from 'lucide-react';

export default function Header() {
  return (
    <header className="sticky top-0 z-50 glass border-b border-surface-700/50">
      <div className="max-w-7xl mx-auto px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-primary-500 to-accent-500 flex items-center justify-center">
              <Shield className="w-6 h-6 text-white" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-white">Sentinel</h1>
              <p className="text-xs text-surface-200">Remediator v1.0</p>
            </div>
          </div>
          
          <nav className="hidden md:flex items-center gap-6">
            <a href="#" className="text-sm text-surface-200 hover:text-primary-400 transition-colors">
              Dashboard
            </a>
            <a href="#" className="text-sm text-surface-200 hover:text-primary-400 transition-colors">
              History
            </a>
            <a href="#" className="text-sm text-surface-200 hover:text-primary-400 transition-colors">
              Settings
            </a>
          </nav>

          <div className="flex items-center gap-3">
            <div className="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-500/20 text-emerald-400 text-xs">
              <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
              System Online
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}
