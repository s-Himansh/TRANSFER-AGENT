"use client";

import { useState, useEffect, useRef, useCallback } from "react";

const API = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

interface Transfer {
  id: string;
  file_name: string;
  file_size: number;
  check_sum: string;
  status: string;
  progress: number;
  mode: string;
  remote_addr: string;
  link: string;
  created_at: string;
}

interface FileRecord {
  name: string;
  size: number;
  check_sum: string;
}

export default function Dashboard() {
  const [transfers, setTransfers] = useState<Transfer[]>([]);
  const [files, setFiles] = useState<FileRecord[]>([]);
  const [uploading, setUploading] = useState(false);
  const [dragOver, setDragOver] = useState(false);
  const [peerAddr, setPeerAddr] = useState("");
  const [peerFile, setPeerFile] = useState("");
  const [activeTransfer, setActiveTransfer] = useState<string | null>(null);
  const [progress, setProgress] = useState(0);
  const wsRef = useRef<WebSocket | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const fetchData = useCallback(async () => {
    try {
      const [tRes, fRes] = await Promise.all([
        fetch(`${API}/api/transfers`),
        fetch(`${API}/api/files`),
      ]);
      const tData = await tRes.json();
      const fData = await fRes.json();
      setTransfers(tData.transfers || []);
      setFiles(fData.files || []);
    } catch {}
  }, []);

  useEffect(() => {
    fetchData();
    const ws = new WebSocket(`${API.replace("http", "ws")}/ws`);
    wsRef.current = ws;
    ws.onmessage = (e) => {
      const event = JSON.parse(e.data);
      if (event.type === "progress") {
        setProgress(event.data.progress);
      }
      if (event.type === "complete") {
        setUploading(false);
        setActiveTransfer(null);
        setProgress(0);
        fetchData();
      }
      if (event.type === "error") {
        setUploading(false);
        setActiveTransfer(null);
        setProgress(0);
        fetchData();
      }
    };
    return () => ws.close();
  }, [fetchData]);

  const handleUpload = async (file: File) => {
    setUploading(true);
    setProgress(0);
    const form = new FormData();
    form.append("file", file);
    try {
      const res = await fetch(`${API}/api/upload`, { method: "POST", body: form });
      const data = await res.json();
      setActiveTransfer(data.id);
      fetchData();
    } catch {
      setUploading(false);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    const file = e.dataTransfer.files[0];
    if (file) handleUpload(file);
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) handleUpload(file);
  };

  const handleSendPeer = async () => {
    if (!peerFile || !peerAddr) return;
    setUploading(true);
    setProgress(0);
    try {
      await fetch(`${API}/api/send`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ file_name: peerFile, addr: peerAddr }),
      });
      fetchData();
    } catch {
      setUploading(false);
    }
  };

  const handleDelete = async (id: string) => {
    await fetch(`${API}/api/transfers/${id}`, { method: "DELETE" });
    fetchData();
  };

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1048576) return `${(bytes / 1024).toFixed(1)} KB`;
    if (bytes < 1073741824) return `${(bytes / 1048576).toFixed(1)} MB`;
    return `${(bytes / 1073741824).toFixed(1)} GB`;
  };

  const statusColor = (s: string) => {
    switch (s) {
      case "completed": return "text-emerald-400";
      case "sending": case "uploading": case "queued": return "text-blue-400";
      case "failed": return "text-red-400";
      default: return "text-slate-400";
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 text-white p-6 md:p-10">
      <div className="max-w-6xl mx-auto space-y-8">
        <header>
          <h1 className="text-3xl font-bold tracking-tight">Transfer Agent</h1>
          <p className="text-slate-400 mt-1">Upload, share, and transfer files securely</p>
        </header>

        {/* Upload Zone */}
        <div
          onDragOver={(e) => { e.preventDefault(); setDragOver(true); }}
          onDragLeave={() => setDragOver(false)}
          onDrop={handleDrop}
          onClick={() => fileInputRef.current?.click()}
          className={`border-2 border-dashed rounded-2xl p-12 text-center cursor-pointer transition-all ${
            dragOver ? "border-blue-500 bg-blue-500/10" : "border-slate-700 hover:border-slate-500"
          }`}
        >
          <input ref={fileInputRef} type="file" className="hidden" onChange={handleFileSelect} />
          <div className="text-4xl mb-3">&#128228;</div>
          {uploading ? (
            <div className="space-y-3">
              <p className="text-slate-300">Uploading...</p>
              <div className="w-64 mx-auto bg-slate-800 rounded-full h-2">
                <div
                  className="bg-blue-500 h-2 rounded-full transition-all duration-300"
                  style={{ width: `${progress}%` }}
                />
              </div>
              <p className="text-sm text-slate-400">{progress.toFixed(0)}%</p>
            </div>
          ) : (
            <>
              <p className="text-slate-300 text-lg">Drop a file here or click to browse</p>
              <p className="text-slate-500 text-sm mt-1">Any file type, up to 512 MB</p>
            </>
          )}
        </div>

        {/* Send to Peer */}
        <div className="bg-slate-900/60 backdrop-blur-sm rounded-2xl p-6 border border-slate-800">
          <h2 className="text-lg font-semibold mb-4">Send to Remote Peer</h2>
          <div className="flex flex-col sm:flex-row gap-3">
            <select
              value={peerFile}
              onChange={(e) => setPeerFile(e.target.value)}
              className="flex-1 bg-slate-800 border border-slate-700 rounded-lg px-4 py-2.5 text-sm focus:outline-none focus:border-blue-500"
            >
              <option value="">Select a file...</option>
              {files.map((f) => (
                <option key={f.name} value={f.name}>{f.name} ({formatSize(f.size)})</option>
              ))}
            </select>
            <input
              value={peerAddr}
              onChange={(e) => setPeerAddr(e.target.value)}
              placeholder="Receiver IP:port (e.g. 192.168.1.5:6789)"
              className="flex-1 bg-slate-800 border border-slate-700 rounded-lg px-4 py-2.5 text-sm focus:outline-none focus:border-blue-500"
            />
            <button
              onClick={handleSendPeer}
              disabled={!peerFile || !peerAddr || uploading}
              className="px-6 py-2.5 bg-blue-600 hover:bg-blue-500 disabled:bg-slate-700 disabled:text-slate-500 rounded-lg text-sm font-medium transition-colors"
            >
              Send
            </button>
          </div>
        </div>

        <div className="grid md:grid-cols-2 gap-8">
          {/* Transfer History */}
          <div className="bg-slate-900/60 backdrop-blur-sm rounded-2xl p-6 border border-slate-800">
            <h2 className="text-lg font-semibold mb-4">Transfer History</h2>
            {transfers.length === 0 ? (
              <p className="text-slate-500 text-sm">No transfers yet</p>
            ) : (
              <div className="space-y-3 max-h-96 overflow-y-auto">
                {transfers.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()).map((t) => (
                  <div key={t.id} className="flex items-center justify-between bg-slate-800/50 rounded-lg p-3">
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium truncate">{t.file_name}</p>
                      <p className="text-xs text-slate-400">
                        {formatSize(t.file_size)} &middot; {t.mode} &middot;{" "}
                        <span className={statusColor(t.status)}>{t.status}</span>
                      </p>
                      {t.status === "sending" && (
                        <div className="w-full bg-slate-700 rounded-full h-1.5 mt-2">
                          <div
                            className="bg-blue-500 h-1.5 rounded-full transition-all"
                            style={{ width: `${t.progress}%` }}
                          />
                        </div>
                      )}
                    </div>
                    <div className="flex gap-2 ml-3">
                      {t.status === "completed" && (
                        <a
                          href={`${API}${t.link}`}
                          target="_blank"
                          rel="noopener"
                          className="text-xs bg-slate-700 hover:bg-slate-600 px-3 py-1 rounded-md transition-colors"
                        >
                          Share
                        </a>
                      )}
                      <button
                        onClick={() => handleDelete(t.id)}
                        className="text-xs bg-slate-700 hover:bg-red-600 px-3 py-1 rounded-md transition-colors"
                      >
                        Delete
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Received Files */}
          <div className="bg-slate-900/60 backdrop-blur-sm rounded-2xl p-6 border border-slate-800">
            <h2 className="text-lg font-semibold mb-4">Received Files</h2>
            {files.length === 0 ? (
              <p className="text-slate-500 text-sm">No files received yet</p>
            ) : (
              <div className="space-y-3 max-h-96 overflow-y-auto">
                {files.map((f) => (
                  <div key={f.name} className="flex items-center justify-between bg-slate-800/50 rounded-lg p-3">
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium truncate">{f.name}</p>
                      <p className="text-xs text-slate-400">{formatSize(f.size)}</p>
                    </div>
                    <a
                      href={`${API}/api/files/${f.name}`}
                      download
                      className="text-xs bg-slate-700 hover:bg-blue-600 px-3 py-1 rounded-md transition-colors ml-3"
                    >
                      Download
                    </a>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
