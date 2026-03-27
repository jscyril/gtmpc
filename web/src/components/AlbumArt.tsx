/**
 * src/components/AlbumArt.tsx
 * Displays album cover art; falls back to a styled placeholder.
 */
import { useState } from 'react';
import { coverURL } from '../api/library';

interface AlbumArtProps {
  trackId: string;
  title?: string;
  size?: number;
  className?: string;
}

export function AlbumArt({ trackId, title = 'Album Art', size = 56, className = '' }: AlbumArtProps) {
  const [failed, setFailed] = useState(false);
  const url = coverURL(trackId);

  if (failed || !trackId) {
    return (
      <div
        className={`flex items-center justify-center rounded-lg bg-[#374151] text-[#6B7280] text-xs font-medium ${className}`}
        style={{ width: size, height: size, minWidth: size }}
      >
        🎵
      </div>
    );
  }

  return (
    <img
      src={url}
      alt={title}
      onError={() => setFailed(true)}
      className={`rounded-lg object-cover ${className}`}
      style={{ width: size, height: size, minWidth: size }}
    />
  );
}
