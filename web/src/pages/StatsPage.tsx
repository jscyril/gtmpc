/**
 * src/pages/StatsPage.tsx
 * Session statistics dashboard — displays play count, listening time, liked
 * tracks, top artist, mean/stddev duration, and a visual artist breakdown chart.
 *
 * All data is driven by StatsContext (in-memory, client-side).
 * Mirrors the statistical computations in pkg/stats/stats.go (Go backend).
 */
import { useStats } from '../context/StatsContext';
import { usePlayer } from '../context/PlayerContext';
import type { Track } from '../api/types';

// ── Sub-components ─────────────────────────────────────────────────────────────

function StatCard({
  icon, label, value, sub, accent = false,
}: {
  icon: string;
  label: string;
  value: string | number;
  sub?: string;
  accent?: boolean;
}) {
  return (
    <div
      className={`rounded-2xl p-5 flex flex-col gap-1 border transition-all ${
        accent
          ? 'bg-gradient-to-br from-[#4C1D95] to-[#7C3AED] border-[#7C3AED] shadow-lg shadow-purple-900/30'
          : 'bg-[#1F2937] border-[#374151] hover:border-[#7C3AED]'
      }`}
    >
      <span className="text-2xl">{icon}</span>
      <span className="text-[#9CA3AF] text-xs uppercase tracking-widest mt-1">{label}</span>
      <span className={`text-3xl font-extrabold ${accent ? 'text-white' : 'text-[#F9FAFB]'}`}>{value}</span>
      {sub && <span className="text-[#6B7280] text-xs mt-0.5">{sub}</span>}
    </div>
  );
}

function MathPanel({ mean, stddev }: { mean: string; stddev: string }) {
  return (
    <div className="rounded-2xl p-5 border border-[#374151] bg-[#1F2937] flex flex-col gap-3">
      <h3 className="text-[#7C3AED] font-semibold text-sm uppercase tracking-widest">
        📐 Statistical Analysis
      </h3>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <p className="text-[#6B7280] text-xs">Mean Track Duration</p>
          <p className="text-[#F9FAFB] text-xl font-bold mt-0.5">{mean}</p>
          <p className="text-[#6B7280] text-xs mt-1">arithmetic mean (μ)</p>
        </div>
        <div>
          <p className="text-[#6B7280] text-xs">Std Deviation</p>
          <p className="text-[#10B981] text-xl font-bold mt-0.5">{stddev}</p>
          <p className="text-[#6B7280] text-xs mt-1">population σ</p>
        </div>
      </div>
      <p className="text-[#4B5563] text-xs border-t border-[#374151] pt-2 mt-1">
        σ = √(Σ(xᵢ − μ)² / n) &nbsp;·&nbsp; computed over {'\u00A0'}session plays
      </p>
    </div>
  );
}

function ArtistChart({ chart }: { chart: ReturnType<typeof useStats>['summary']['artistChart'] }) {
  if (chart.length === 0) return null;
  return (
    <div className="rounded-2xl p-5 border border-[#374151] bg-[#1F2937] flex flex-col gap-3">
      <h3 className="text-[#7C3AED] font-semibold text-sm uppercase tracking-widest">
        🎤 Artist Breakdown
      </h3>
      <div className="flex flex-col gap-2">
        {chart.map(({ artist, count, pct }) => (
          <div key={artist} className="flex items-center gap-3">
            <span className="text-[#9CA3AF] text-sm w-36 truncate shrink-0">{artist}</span>
            <div className="flex-1 bg-[#374151] rounded-full h-2 overflow-hidden">
              <div
                className="h-2 rounded-full bg-gradient-to-r from-[#7C3AED] to-[#10B981] transition-all duration-700"
                style={{ width: `${pct}%` }}
              />
            </div>
            <span className="text-[#6B7280] text-xs w-6 text-right shrink-0">{count}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

// ── Main page ──────────────────────────────────────────────────────────────────

export function StatsPage() {
  const { summary, likedIds, toggleLike, clearStats } = useStats();
  const { currentTrack } = usePlayer();

  const empty = summary.tracksPlayed === 0;

  return (
    <div className="flex flex-col gap-6 h-full">
      {/* Header */}
      <div className="flex items-center justify-between flex-wrap gap-3">
        <div>
          <h2 className="text-2xl font-bold text-[#F9FAFB]">Session Stats</h2>
          <p className="text-[#6B7280] text-sm mt-0.5">
            {empty ? 'Play some music to see your stats!' : `${summary.tracksPlayed} plays this session`}
          </p>
        </div>
        {!empty && (
          <button
            onClick={clearStats}
            className="px-4 py-2 rounded-lg text-sm border border-[#374151] text-[#6B7280] hover:text-[#EF4444] hover:border-[#EF4444] transition-colors"
          >
            🗑 Clear Session
          </button>
        )}
      </div>

      {empty ? (
        <div className="flex-1 flex flex-col items-center justify-center gap-4 text-center">
          <span className="text-6xl opacity-40">📊</span>
          <p className="text-[#6B7280] text-lg">No tracks played yet.</p>
          <p className="text-[#4B5563] text-sm">Head to the Library and start listening!</p>
        </div>
      ) : (
        <>
          {/* Stat cards grid */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <StatCard
              icon="🎵"
              label="Songs Played"
              value={summary.tracksPlayed}
              accent
            />
            <StatCard
              icon="♥"
              label="Songs Liked"
              value={summary.tracksLiked}
              sub={likedIds.size > 0 ? [...likedIds].slice(0, 2).join(', ').slice(0, 20) + '…' : undefined}
            />
            <StatCard
              icon="⏱"
              label="Listen Time"
              value={summary.formattedTime}
            />
            <StatCard
              icon="🏆"
              label="Top Artist"
              value={summary.topArtist || '—'}
              sub={summary.topArtist ? `${summary.artistCounts[summary.topArtist]} plays` : undefined}
            />
          </div>

          {/* Math panel + artist chart */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <MathPanel mean={summary.formattedMean} stddev={summary.formattedStdDev} />
            <ArtistChart chart={summary.artistChart} />
          </div>

          {/* Most replayed */}
          {summary.mostPlayedCount > 1 && (
            <div className="rounded-2xl p-4 border border-[#374151] bg-[#1F2937] flex items-center gap-4">
              <span className="text-3xl">🔁</span>
              <div>
                <p className="text-[#6B7280] text-xs uppercase tracking-widest">Most Replayed</p>
                <p className="text-[#F9FAFB] font-semibold">
                  {summary.mostPlayedTitle}
                  <span className="text-[#7C3AED] ml-2 text-sm">×{summary.mostPlayedCount}</span>
                </p>
              </div>
            </div>
          )}

          {/* Now playing — like button */}
          {currentTrack && (
            <div className="rounded-2xl p-4 border border-[#374151] bg-[#1F2937] flex items-center justify-between gap-4">
              <div>
                <p className="text-[#6B7280] text-xs uppercase tracking-widest">Now Playing</p>
                <p className="text-[#F9FAFB] font-semibold">{currentTrack.title}</p>
                <p className="text-[#6B7280] text-sm">{currentTrack.artist}</p>
              </div>
              <button
                onClick={() => toggleLike(currentTrack as Track)}
                className={`text-3xl transition-transform hover:scale-125 ${
                  likedIds.has(currentTrack.id) ? 'text-[#EF4444]' : 'text-[#374151]'
                }`}
                title={likedIds.has(currentTrack.id) ? 'Unlike' : 'Like'}
              >
                ♥
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
