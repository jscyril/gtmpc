/**
 * src/components/TrackRow.tsx
 * A single row in the track list table.
 */
import type { Track } from '../api/types';
import { AlbumArt } from './AlbumArt';
import { formatTime } from '../utils/formatTime';

interface TrackRowProps {
  track: Track;
  index: number;
  isPlaying: boolean;
  isSelected: boolean;
  onClick: () => void;
}

export function TrackRow({ track, index, isPlaying, isSelected, onClick }: TrackRowProps) {
  const rowClass = isPlaying
    ? 'bg-[#7C3AED]/20 text-[#10B981]'
    : isSelected
    ? 'bg-[#374151]'
    : 'hover:bg-[#1F2937]';

  return (
    <tr
      onClick={onClick}
      className={`cursor-pointer transition-colors border-b border-[#374151] ${rowClass}`}
    >
      <td className="px-4 py-3 text-[#6B7280] text-sm w-12">
        {isPlaying ? (
          <span className="inline-block w-4 text-[#10B981] animate-pulse">♫</span>
        ) : (
          index + 1
        )}
      </td>
      <td className="px-4 py-3">
        <div className="flex items-center gap-3">
          <AlbumArt trackId={track.id} title={track.title} size={36} />
          <div className="min-w-0">
            <p className="truncate font-medium text-sm">{track.title}</p>
            <p className="truncate text-xs text-[#6B7280]">{track.artist}</p>
          </div>
        </div>
      </td>
      <td className="px-4 py-3 text-sm text-[#6B7280] hidden md:table-cell truncate max-w-[12rem]">
        {track.album}
      </td>
      <td className="px-4 py-3 text-sm text-[#6B7280] text-right">
        {formatTime(track.duration_seconds)}
      </td>
    </tr>
  );
}
