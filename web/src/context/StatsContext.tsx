/**
 * src/context/StatsContext.tsx
 * Client-side session statistics tracker.
 *
 * Tracks: play count per track, total listening seconds, liked tracks,
 * and artist play frequency. Computes mean duration and standard deviation
 * to mirror the statistical computations in pkg/stats/stats.go.
 *
 * Data is stored in React state only (in-memory, resets on page reload).
 * No backend changes are required.
 */

import React, {
  createContext, useContext, useState, useCallback,
} from 'react';
import type { Track } from '../api/types';

// ── Types ──────────────────────────────────────────────────────────────────────

export interface ArtistBar {
  artist: string;
  count: number;
  bar: string; // pre-rendered CSS width percentage string
  pct: number;
}

export interface StatsSummary {
  tracksPlayed: number;
  tracksLiked: number;
  totalSeconds: number;
  formattedTime: string;
  topArtist: string;
  artistCounts: Record<string, number>;
  artistChart: ArtistBar[];
  meanDurationSec: number;
  stdDevSec: number;
  formattedMean: string;
  formattedStdDev: string;
  mostPlayedTitle: string;
  mostPlayedCount: number;
}

interface StatsContextValue {
  summary: StatsSummary;
  likedIds: Set<string>;
  isLiked: (id: string) => boolean;
  toggleLike: (track: Track) => void;
  recordPlay: (track: Track) => void;
  clearStats: () => void;
}

// ── Helpers (mirrors pkg/stats) ────────────────────────────────────────────────

/** Formats total seconds as human-readable e.g. "1h 02m 05s" */
export function formatListenTime(totalSecs: number): string {
  if (totalSecs <= 0) return '0s';
  const h = Math.floor(totalSecs / 3600);
  const m = Math.floor((totalSecs % 3600) / 60);
  const s = totalSecs % 60;
  if (h > 0) return `${h}h ${String(m).padStart(2, '0')}m ${String(s).padStart(2, '0')}s`;
  if (m > 0) return `${m}m ${String(s).padStart(2, '0')}s`;
  return `${s}s`;
}

/** Formats seconds as m:ss */
function formatMmSs(secs: number): string {
  if (secs <= 0) return '0:00';
  const m = Math.floor(secs / 60);
  const s = Math.round(secs % 60);
  return `${m}:${String(s).padStart(2, '0')}`;
}

// ── Play event store ───────────────────────────────────────────────────────────

interface PlayEvent {
  trackId: string;
  title: string;
  artist: string;
  durationSecs: number;
}

function computeSummary(events: PlayEvent[], likedIds: Set<string>): StatsSummary {
  const n = events.length;
  if (n === 0) {
    return {
      tracksPlayed: 0,
      tracksLiked: likedIds.size,
      totalSeconds: 0,
      formattedTime: '0s',
      topArtist: '',
      artistCounts: {},
      artistChart: [],
      meanDurationSec: 0,
      stdDevSec: 0,
      formattedMean: '—',
      formattedStdDev: '—',
      mostPlayedTitle: '',
      mostPlayedCount: 0,
    };
  }

  // Total seconds
  const totalSeconds = events.reduce((acc, e) => acc + e.durationSecs, 0);

  // Artist frequency map
  const artistCounts: Record<string, number> = {};
  for (const e of events) {
    const artist = e.artist.trim() || 'Unknown';
    artistCounts[artist] = (artistCounts[artist] ?? 0) + 1;
  }

  // Top artist
  let topArtist = '';
  let topCount = 0;
  for (const [artist, count] of Object.entries(artistCounts)) {
    if (count > topCount || (count === topCount && artist < topArtist)) {
      topArtist = artist;
      topCount = count;
    }
  }

  // Most-played track
  const trackCounts: Record<string, number> = {};
  const trackTitles: Record<string, string> = {};
  for (const e of events) {
    trackCounts[e.trackId] = (trackCounts[e.trackId] ?? 0) + 1;
    trackTitles[e.trackId] = e.title;
  }
  let mostPlayedTitle = '';
  let mostPlayedCount = 0;
  for (const [id, count] of Object.entries(trackCounts)) {
    if (count > mostPlayedCount) {
      mostPlayedCount = count;
      mostPlayedTitle = trackTitles[id] ?? '';
    }
  }

  // Mean duration (arithmetic mean)
  const mean = totalSeconds / n;

  // Population standard deviation: σ = √(Σ(xi - μ)² / n)
  const variance = events.reduce((acc, e) => {
    const diff = e.durationSecs - mean;
    return acc + diff * diff;
  }, 0) / n;
  const stdDev = Math.sqrt(variance);

  // Artist bar chart (sorted desc by count, max 8 entries)
  const maxCount = Math.max(...Object.values(artistCounts), 1);
  const artistChart: ArtistBar[] = Object.entries(artistCounts)
    .sort(([a, ca], [b, cb]) => cb - ca || a.localeCompare(b))
    .slice(0, 8)
    .map(([artist, count]) => ({
      artist,
      count,
      pct: Math.max(4, Math.round((count / maxCount) * 100)),
      bar: '',
    }));

  return {
    tracksPlayed: n,
    tracksLiked: likedIds.size,
    totalSeconds,
    formattedTime: formatListenTime(totalSeconds),
    topArtist,
    artistCounts,
    artistChart,
    meanDurationSec: mean,
    stdDevSec: stdDev,
    formattedMean: formatMmSs(mean),
    formattedStdDev: `±${formatMmSs(stdDev)}`,
    mostPlayedTitle,
    mostPlayedCount,
  };
}

// ── Context ────────────────────────────────────────────────────────────────────

const StatsContext = createContext<StatsContextValue | null>(null);

export function StatsProvider({ children }: { children: React.ReactNode }) {
  const [events, setEvents] = useState<PlayEvent[]>([]);
  const [likedIds, setLikedIds] = useState<Set<string>>(new Set());

  const recordPlay = useCallback((track: Track) => {
    setEvents((prev) => [
      ...prev,
      {
        trackId: track.id,
        title: track.title,
        artist: track.artist,
        durationSecs: track.duration_seconds,
      },
    ]);
  }, []);

  const toggleLike = useCallback((track: Track) => {
    setLikedIds((prev) => {
      const next = new Set(prev);
      if (next.has(track.id)) {
        next.delete(track.id);
      } else {
        next.add(track.id);
      }
      return next;
    });
  }, []);

  const isLiked = useCallback(
    (id: string) => likedIds.has(id),
    [likedIds],
  );

  const clearStats = useCallback(() => {
    setEvents([]);
    setLikedIds(new Set());
  }, []);

  const summary = computeSummary(events, likedIds);

  return (
    <StatsContext.Provider value={{ summary, likedIds, isLiked, toggleLike, recordPlay, clearStats }}>
      {children}
    </StatsContext.Provider>
  );
}

export function useStats(): StatsContextValue {
  const ctx = useContext(StatsContext);
  if (!ctx) throw new Error('useStats must be used within StatsProvider');
  return ctx;
}
