/**
 * src/api/library.ts
 * Library API calls: tracks, playlists, cover art, stream URL.
 */

import apiClient from './client';
import type {
  TrackListResponse,
  PlaylistListResponse,
  CreatePlaylistRequest,
  CreatePlaylistResponse,
} from './types';

export async function getTracks(): Promise<TrackListResponse> {
  const { data } = await apiClient.get<TrackListResponse>('/api/library/tracks');
  return data;
}

export async function getPlaylists(): Promise<PlaylistListResponse> {
  const { data } = await apiClient.get<PlaylistListResponse>('/api/library/playlists');
  return data;
}

export async function createPlaylist(req: CreatePlaylistRequest): Promise<CreatePlaylistResponse> {
  const { data } = await apiClient.post<CreatePlaylistResponse>('/api/library/playlists', req);
  return data;
}

/** Returns the URL to stream audio for a track. Used as <audio src>. */
export function streamURL(trackId: string): string {
  return `/api/stream/${trackId}`;
}

/** Returns the URL to fetch cover art for a track. */
export function coverURL(trackId: string): string {
  return `/api/library/cover/${trackId}`;
}
