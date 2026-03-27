/**
 * src/pages/PlaylistsPage.tsx
 * Playlist browser — list playlists, expand to see tracks, play tracks.
 */
import { useState, useEffect } from 'react';
import { getPlaylists, getTracks } from '../api/library';
import type { Playlist, Track } from '../api/types';
import { usePlayer } from '../context/PlayerContext';
import { TrackList } from '../components/TrackList';
import { LoadingSpinner } from '../components/LoadingSpinner';

export function PlaylistsPage() {
  const [playlists, setPlaylists] = useState<Playlist[]>([]);
  const [allTracks, setAllTracks] = useState<Track[]>([]);
  const [selected, setSelected] = useState<Playlist | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const { currentTrack, isPlaying, play } = usePlayer();

  useEffect(() => {
    Promise.all([getPlaylists(), getTracks()])
      .then(([plResp, trResp]) => {
        setPlaylists(plResp.playlists ?? []);
        setAllTracks(trResp.tracks ?? []);
      })
      .catch(() => setError('Failed to load playlists.'))
      .finally(() => setLoading(false));
  }, []);

  const playlistTracks = (playlist: Playlist): Track[] => {
    const trackMap = new Map(allTracks.map((t) => [t.id, t]));
    return (playlist.track_ids ?? []).flatMap((id) => {
      const t = trackMap.get(id);
      return t ? [t] : [];
    });
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-48">
        <LoadingSpinner label="Loading playlists…" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-48 text-[#EF4444]">
        {error}
      </div>
    );
  }

  return (
    <div className="flex h-full gap-4">
      {/* Playlist list */}
      <div className="w-64 bg-[#1F2937] rounded-xl border border-[#374151] flex flex-col shrink-0">
        <div className="p-4 border-b border-[#374151]">
          <h2 className="text-lg font-bold text-[#F9FAFB]">Playlists</h2>
          <p className="text-[#6B7280] text-xs mt-0.5">
            {playlists.length} playlist{playlists.length !== 1 ? 's' : ''}
          </p>
        </div>
        {playlists.length === 0 ? (
          <div className="flex-1 flex items-center justify-center p-4 text-[#6B7280] text-sm text-center">
            No playlists yet. Create one via the server API or CLI.
          </div>
        ) : (
          <ul className="overflow-y-auto flex-1 p-2 flex flex-col gap-1">
            {playlists.map((pl) => (
              <li key={pl.id}>
                <button
                  onClick={() => setSelected(pl)}
                  className={`w-full text-left px-3 py-2.5 rounded-lg text-sm transition-colors ${
                    selected?.id === pl.id
                      ? 'bg-[#7C3AED] text-white'
                      : 'text-[#6B7280] hover:bg-[#374151] hover:text-[#F9FAFB]'
                  }`}
                >
                  <p className="font-medium truncate">{pl.name}</p>
                  <p className="text-xs opacity-70 mt-0.5">
                    {(pl.track_ids ?? []).length} tracks
                  </p>
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>

      {/* Track list for selected playlist */}
      <div className="flex-1 bg-[#1F2937] rounded-xl border border-[#374151] overflow-hidden">
        {selected ? (
          <div className="h-full flex flex-col">
            <div className="p-4 border-b border-[#374151]">
              <h3 className="text-xl font-bold text-[#F9FAFB]">{selected.name}</h3>
              <p className="text-[#6B7280] text-sm mt-0.5">
                {(selected.track_ids ?? []).length} tracks
              </p>
            </div>
            <div className="flex-1 overflow-auto">
              <TrackList
                tracks={playlistTracks(selected)}
                currentTrackId={currentTrack?.id}
                isPlaying={isPlaying}
                onPlay={(track, idx) => play(track, playlistTracks(selected), idx)}
              />
            </div>
          </div>
        ) : (
          <div className="flex items-center justify-center h-full text-[#6B7280]">
            ← Select a playlist to view tracks
          </div>
        )}
      </div>
    </div>
  );
}
