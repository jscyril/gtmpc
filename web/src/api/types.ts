/**
 * src/api/types.ts
 * TypeScript interfaces matching the gtmpc REST API contract.
 */

export interface Track {
  id: string;
  title: string;
  artist: string;
  album: string;
  duration_seconds: number;
  format: string;
  cover_url: string;
}

export interface Playlist {
  id: string;
  name: string;
  track_ids: string[];
  created_at: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  username: string;
  expires_at: string;
}

export interface RegisterRequest {
  username: string;
  password: string;
}

export interface RegisterResponse {
  message: string;
  user_id: string;
}

export interface TrackListResponse {
  tracks: Track[];
}

export interface PlaylistListResponse {
  playlists: Playlist[];
}

export interface CreatePlaylistRequest {
  name: string;
  track_ids: string[];
}

export interface CreatePlaylistResponse {
  id: string;
  name: string;
}

export interface HealthResponse {
  status: string;
  version: string;
}

export interface APIError {
  error: string;
}
