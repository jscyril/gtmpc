/**
 * src/pages/LibraryPage.tsx
 * Music library browser with search/filter and playback.
 */
import { useState, useEffect, useMemo } from 'react';
import { getTracks } from '../api/library';
import type { Track } from '../api/types';
import { usePlayer } from '../context/PlayerContext';
import { SearchBar } from '../components/SearchBar';
import { TrackList } from '../components/TrackList';
import { LoadingSpinner } from '../components/LoadingSpinner';

export function LibraryPage() {
  const [tracks, setTracks] = useState<Track[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [search, setSearch] = useState('');
  const { currentTrack, isPlaying, play } = usePlayer();

  const fetchTracks = async () => {
    setLoading(true);
    setError('');
    try {
      const resp = await getTracks();
      setTracks(resp.tracks ?? []);
    } catch {
      setError('Failed to load tracks. Is the server running?');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchTracks(); }, []);

  const filtered = useMemo(() => {
    if (!search.trim()) return tracks;
    const q = search.toLowerCase();
    return tracks.filter(
      (t) =>
        t.title.toLowerCase().includes(q) ||
        t.artist.toLowerCase().includes(q) ||
        t.album.toLowerCase().includes(q)
    );
  }, [tracks, search]);

  const handlePlay = (track: Track, _index: number) => {
    play(track, filtered, filtered.indexOf(track));
  };

  return (
    <div className="h-full flex flex-col gap-4">
      <div className="flex items-center justify-between flex-wrap gap-3">
        <div>
          <h2 className="text-2xl font-bold text-[#F9FAFB]">Library</h2>
          <p className="text-[#6B7280] text-sm mt-0.5">
            {loading ? 'Loading…' : `${filtered.length} track${filtered.length !== 1 ? 's' : ''}`}
          </p>
        </div>
        <div className="flex gap-2 items-center">
          <SearchBar value={search} onChange={setSearch} placeholder="Search title, artist, album…" />
          <button
            onClick={fetchTracks}
            className="px-3 py-2 rounded-lg border border-[#4B5563] text-[#6B7280] hover:text-[#F9FAFB] hover:border-[#7C3AED] transition-colors text-sm"
            aria-label="Refresh"
          >
            ↻
          </button>
        </div>
      </div>

      <div className="flex-1 bg-[#1F2937] rounded-xl border border-[#374151] overflow-hidden">
        {loading ? (
          <div className="flex items-center justify-center h-48">
            <LoadingSpinner label="Loading tracks…" />
          </div>
        ) : error ? (
          <div className="flex flex-col items-center justify-center h-48 gap-3">
            <p className="text-[#EF4444]">{error}</p>
            <button
              onClick={fetchTracks}
              className="px-4 py-1.5 rounded bg-[#374151] text-sm hover:bg-[#4B5563]"
            >
              Retry
            </button>
          </div>
        ) : (
          <TrackList
            tracks={filtered}
            currentTrackId={currentTrack?.id}
            isPlaying={isPlaying}
            onPlay={handlePlay}
          />
        )}
      </div>
    </div>
  );
}
