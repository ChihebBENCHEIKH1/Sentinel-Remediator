import React, { useState } from 'react';
import { Shield, Zap, Download, AlertCircle } from 'lucide-react';
import clsx from 'clsx';

interface SubmitFormProps {
  onJobCreated?: (jobId: string) => void;
}

const SubmitForm: React.FC<SubmitFormProps> = ({ onJobCreated }) => {
  const [scanJson, setScanJson] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadSample = () => {
    const sample = {
      scan_result: {
        scan_id: "scan-001",
        image_name: "myapp",
        image_tag: "latest",
        repo_url: "https://github.com/example/myapp",
        branch: "main",
        vulnerabilities: [
          {
            id: "VULN-001",
            severity: "HIGH",
            type: "RUN_AS_ROOT",
            title: "Container runs as root user",
            description: "The container is configured to run as the root user...",
            file_path: "Dockerfile",
            line_number: 1,
            suggestion: "Add a non-root user and use the USER directive."
          }
        ]
      }
    };
    setScanJson(JSON.stringify(sample, null, 2));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!scanJson) {
      setError("Scan JSON is required");
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await fetch('http://localhost:8080/api/remediate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: scanJson,
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Failed to submit remediation job');
      }

      if (onJobCreated) {
        onJobCreated(data.job_id);
      }
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <label className="text-xs font-semibold text-slate-500 uppercase tracking-widest pl-1">
          Input Data Source
        </label>
        <div className="relative group">
          <textarea
            value={scanJson}
            onChange={(e) => setScanJson(e.target.value)}
            className="w-full h-80 bg-slate-950 border border-slate-700 rounded-xl p-4 text-sm font-mono text-slate-300 focus:outline-none focus:ring-2 focus:ring-indigo-500/50 focus:border-indigo-500 transition-all resize-none group-hover:border-slate-600"
            placeholder='Paste security scan JSON here...'
          />
        </div>
      </div>

      {error && (
        <div className="text-xs text-rose-400 bg-rose-400/10 border border-rose-400/20 p-3 rounded-lg flex items-center gap-2">
          <AlertCircle className="w-4 h-4 flex-shrink-0" />
          {error}
        </div>
      )}

      <div className="flex gap-3">
        <button
          onClick={loadSample}
          type="button"
          className="flex-1 px-4 py-3 bg-slate-800 hover:bg-slate-700 text-slate-300 text-sm font-semibold rounded-xl border border-slate-700 transition-all flex items-center justify-center gap-2"
        >
          <Download className="w-4 h-4" />
          Load Sample
        </button>
        <button
          type="submit"
          disabled={loading}
          className="flex-[2] bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white font-bold py-3 rounded-xl shadow-lg shadow-indigo-600/20 transform transition-all active:scale-[0.98] disabled:active:scale-100 flex items-center justify-center gap-2"
        >
          {loading ? (
            <Zap className="w-4 h-4 animate-spin" />
          ) : (
            <Shield className="w-4 h-4" />
          )}
          {loading ? "Initializing..." : "Run Remediation Agent"}
        </button>
      </div>
    </form>
  );
};

export default SubmitForm;
