/**
 * src/components/Layout.tsx
 * Main application shell: top navbar, left sidebar, content area, and NowPlaying bar.
 */
import { Outlet } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { Sidebar } from './Sidebar';
import { NowPlaying } from './NowPlaying';

export function Layout() {
  const { username, logout } = useAuth();

  return (
    <div className="flex flex-col h-screen bg-[#111827]">
      {/* Top navbar */}
      <header className="h-14 bg-[#1F2937] border-b border-[#374151] flex items-center justify-between px-6 shrink-0 z-40">
        <span className="text-[#7C3AED] font-bold text-lg">gtmpc</span>
        <div className="flex items-center gap-4">
          <span className="text-[#6B7280] text-sm">{username}</span>
          <button
            onClick={logout}
            className="text-xs text-[#6B7280] hover:text-[#EF4444] transition-colors px-3 py-1 rounded border border-[#374151] hover:border-[#EF4444]"
          >
            Log Out
          </button>
        </div>
      </header>

      {/* Middle: sidebar + content */}
      <div className="flex flex-1 min-h-0">
        <Sidebar />
        <main className="flex-1 overflow-y-auto pb-20 p-6">
          <Outlet />
        </main>
      </div>

      {/* Fixed bottom player */}
      <NowPlaying />
    </div>
  );
}
