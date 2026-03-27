/**
 * src/context/PlayerContext.tsx
 * Manages audio playback state using the HTML5 Audio API.
 *
 * Audio authorization strategy:
 * The HTML5 Audio API doesn't natively support custom headers. Two approaches exist:
 *
 * Approach A (implemented): Fetch-blob
 *   Use fetch() with Authorization header, convert response to Blob URL, set as audio src.
 *   Pro: Token never exposed in URL. Con: Entire file buffered before playback starts.
 *
 * Approach B (commented fallback): Token as query param
 *   Set src = `/api/stream/${id}?token=${token}` if the backend supports it.
 *   Pro: Native seek support. Con: Token in URL (visible in logs/history).
 *
 * We default to Approach A (blob). For large files you may prefer Approach B.
 */

import React, {
  createContext, useContext, useState, useRef, useCallback, useEffect,
} from 'react';
import type { Track } from '../api/types';
import { streamURL } from '../api/library';
import { getToken } from '../utils/storage';
import { useStats } from './StatsContext';

interface PlayerState {
  currentTrack: Track | null;
  queue: Track[];
  queueIndex: number;
  isPlaying: boolean;
  progress: number;    // seconds elapsed
  duration: number;   // seconds total
  volume: number;     // 0–100
}

interface PlayerContextValue extends PlayerState {
  play: (track: Track, queue?: Track[], index?: number) => void;
  pause: () => void;
  resume: () => void;
  next: () => void;
  previous: () => void;
  seek: (seconds: number) => void;
  setVolume: (pct: number) => void;
}

const PlayerContext = createContext<PlayerContextValue | null>(null);

export function PlayerProvider({ children }: { children: React.ReactNode }) {
  const audioRef = useRef(new Audio());
  const { recordPlay } = useStats();
  const [state, setState] = useState<PlayerState>({
    currentTrack: null,
    queue: [],
    queueIndex: 0,
    isPlaying: false,
    progress: 0,
    duration: 0,
    volume: 80,
  });
  const blobUrlRef = useRef<string | null>(null);

  // Sync volume to audio element
  useEffect(() => {
    audioRef.current.volume = state.volume / 100;
  }, [state.volume]);

  // Set up audio event listeners once
  useEffect(() => {
    const audio = audioRef.current;

    const onTimeUpdate = () => {
      setState((s) => ({ ...s, progress: audio.currentTime }));
    };
    const onLoadedMetadata = () => {
      setState((s) => ({ ...s, duration: audio.duration || 0 }));
    };
    const onEnded = () => {
      setState((s) => {
        if (s.queueIndex < s.queue.length - 1) {
          // Auto-advance to next track
          return s; // handled via play() call below
        }
        return { ...s, isPlaying: false };
      });
      // Auto-advance
      setState((s) => {
        const nextIdx = s.queueIndex + 1;
        if (nextIdx < s.queue.length) {
          playTrackInternal(s.queue[nextIdx], s.queue, nextIdx);
        }
        return s;
      });
    };
    const onPlay = () => setState((s) => ({ ...s, isPlaying: true }));
    const onPause = () => setState((s) => ({ ...s, isPlaying: false }));

    audio.addEventListener('timeupdate', onTimeUpdate);
    audio.addEventListener('loadedmetadata', onLoadedMetadata);
    audio.addEventListener('ended', onEnded);
    audio.addEventListener('play', onPlay);
    audio.addEventListener('pause', onPause);

    return () => {
      audio.removeEventListener('timeupdate', onTimeUpdate);
      audio.removeEventListener('loadedmetadata', onLoadedMetadata);
      audio.removeEventListener('ended', onEnded);
      audio.removeEventListener('play', onPlay);
      audio.removeEventListener('pause', onPause);
    };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  // Approach A: fetch blob for auth
  const playTrackInternal = useCallback(async (track: Track, queue: Track[], index: number) => {
    const audio = audioRef.current;

    // Revoke previous blob URL to free memory
    if (blobUrlRef.current) {
      URL.revokeObjectURL(blobUrlRef.current);
      blobUrlRef.current = null;
    }

    const token = getToken();

    try {
      if (token) {
        // Approach A: fetch with Authorization header
        const resp = await fetch(streamURL(track.id), {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (!resp.ok) throw new Error(`Stream fetch failed: ${resp.status}`);
        const blob = await resp.blob();
        const objUrl = URL.createObjectURL(blob);
        blobUrlRef.current = objUrl;
        audio.src = objUrl;
      } else {
        // Approach B fallback (no token — dev mode)
        audio.src = streamURL(track.id);
      }

      audio.load();
      await audio.play();

      // Record this play event in the stats tracker
      recordPlay(track);

      setState((s) => ({
        ...s,
        currentTrack: track,
        queue,
        queueIndex: index,
        isPlaying: true,
        progress: 0,
        duration: track.duration_seconds,
      }));
    } catch (err) {
      console.error('Playback error:', err);
      // Fallback: try Approach B (token as query param) if backend supports it
      audio.src = `${streamURL(track.id)}?token=${token ?? ''}`;
      audio.load();
      audio.play().catch(console.error);
      setState((s) => ({
        ...s,
        currentTrack: track,
        queue,
        queueIndex: index,
        isPlaying: true,
      }));
    }
  }, [recordPlay]);

  const play = useCallback((track: Track, queue: Track[] = [track], index = 0) => {
    playTrackInternal(track, queue, index);
  }, [playTrackInternal]);

  const pause = useCallback(() => {
    audioRef.current.pause();
  }, []);

  const resume = useCallback(() => {
    audioRef.current.play().catch(console.error);
  }, []);

  const next = useCallback(() => {
    setState((s) => {
      const nextIdx = s.queueIndex + 1;
      if (nextIdx < s.queue.length) {
        playTrackInternal(s.queue[nextIdx], s.queue, nextIdx);
      }
      return s;
    });
  }, [playTrackInternal]);

  const previous = useCallback(() => {
    setState((s) => {
      if (audioRef.current.currentTime > 3) {
        // Restart current track if >3s in
        audioRef.current.currentTime = 0;
        return s;
      }
      const prevIdx = s.queueIndex - 1;
      if (prevIdx >= 0) {
        playTrackInternal(s.queue[prevIdx], s.queue, prevIdx);
      }
      return s;
    });
  }, [playTrackInternal]);

  const seek = useCallback((seconds: number) => {
    audioRef.current.currentTime = seconds;
    setState((s) => ({ ...s, progress: seconds }));
  }, []);

  const setVolume = useCallback((pct: number) => {
    const clamped = Math.max(0, Math.min(100, pct));
    audioRef.current.volume = clamped / 100;
    setState((s) => ({ ...s, volume: clamped }));
  }, []);

  return (
    <PlayerContext.Provider value={{ ...state, play, pause, resume, next, previous, seek, setVolume }}>
      {children}
    </PlayerContext.Provider>
  );
}

export function usePlayer(): PlayerContextValue {
  const ctx = useContext(PlayerContext);
  if (!ctx) throw new Error('usePlayer must be used within PlayerProvider');
  return ctx;
}
