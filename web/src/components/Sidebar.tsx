/**
 * src/components/Sidebar.tsx
 * Left navigation sidebar.
 */
import { NavLink } from 'react-router-dom';

export function Sidebar() {
  const linkClass = ({ isActive }: { isActive: boolean }) =>
    `flex items-center gap-3 px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
      isActive
        ? 'bg-[#7C3AED] text-white'
        : 'text-[#6B7280] hover:bg-[#374151] hover:text-[#F9FAFB]'
    }`;

  return (
    <aside className="w-56 bg-[#1F2937] flex flex-col border-r border-[#374151] shrink-0">
      <div className="p-5 border-b border-[#374151]">
        <h1 className="text-xl font-bold text-[#7C3AED]">🎵 gtmpc</h1>
        <p className="text-xs text-[#6B7280] mt-0.5">Music Player</p>
      </div>
      <nav className="p-3 flex flex-col gap-1">
        <NavLink to="/library" className={linkClass}>
          <span>📚</span> Library
        </NavLink>
        <NavLink to="/playlists" className={linkClass}>
          <span>📋</span> Playlists
        </NavLink>
      </nav>
    </aside>
  );
}
