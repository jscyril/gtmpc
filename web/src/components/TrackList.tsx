/**
 * src/components/TrackList.tsx
 * Scrollable, sortable table of tracks.
 */
import { useState, useMemo } from 'react';
import type { Track } from '../api/types';
import { TrackRow } from './TrackRow';

type SortKey = 'title' | 'artist' | 'album' | 'duration_seconds';
type SortDir = 'asc' | 'desc';

interface TrackListProps {
  tracks: Track[];
  currentTrackId?: string;
  isPlaying?: boolean;
  onPlay: (track: Track, index: number) => void;
}

export function TrackList({ tracks, currentTrackId, isPlaying, onPlay }: TrackListProps) {
  const [sortKey, setSortKey] = useState<SortKey>('title');
  const [sortDir, setSortDir] = useState<SortDir>('asc');

  const sorted = useMemo(() => {
    return [...tracks].sort((a, b) => {
      const av = a[sortKey];
      const bv = b[sortKey];
      const cmp = typeof av === 'number' ? av - (bv as number) : String(av).localeCompare(String(bv));
      return sortDir === 'asc' ? cmp : -cmp;
    });
  }, [tracks, sortKey, sortDir]);

  const handleSort = (key: SortKey) => {
    if (key === sortKey) {
      setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortKey(key);
      setSortDir('asc');
    }
  };

  const SortIcon = ({ k }: { k: SortKey }) =>
    sortKey === k ? (sortDir === 'asc' ? ' ↑' : ' ↓') : '';

  if (tracks.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-48 text-[#6B7280]">
        <p className="text-lg">🎵</p>
        <p className="mt-2">No tracks found. Add music to the server.</p>
      </div>
    );
  }

  return (
    <div className="overflow-auto">
      <table className="w-full text-left text-[#F9FAFB]">
        <thead>
          <tr className="text-xs uppercase tracking-wider text-[#6B7280] border-b border-[#374151]">
            <th className="px-4 py-2 w-12">#</th>
            <th
              className="px-4 py-2 cursor-pointer hover:text-[#F9FAFB]"
              onClick={() => handleSort('title')}
            >
              Title<SortIcon k="title" />
            </th>
            <th
              className="px-4 py-2 cursor-pointer hover:text-[#F9FAFB] hidden md:table-cell"
              onClick={() => handleSort('album')}
            >
              Album<SortIcon k="album" />
            </th>
            <th
              className="px-4 py-2 cursor-pointer hover:text-[#F9FAFB] text-right"
              onClick={() => handleSort('duration_seconds')}
            >
              Time<SortIcon k="duration_seconds" />
            </th>
          </tr>
        </thead>
        <tbody>
          {sorted.map((track, i) => (
            <TrackRow
              key={track.id}
              track={track}
              index={i}
              isPlaying={track.id === currentTrackId && (isPlaying ?? false)}
              isSelected={track.id === currentTrackId}
              onClick={() => onPlay(track, i)}
            />
          ))}
        </tbody>
      </table>
    </div>
  );
}
