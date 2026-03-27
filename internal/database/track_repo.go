package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jscyril/golang_music_player/api"
)

// TrackRepo handles all track-related database operations.
type TrackRepo struct {
	db *DB
}

// NewTrackRepo creates a new track repository backed by PostgreSQL.
func NewTrackRepo(db *DB) *TrackRepo {
	return &TrackRepo{db: db}
}

// Upsert inserts a track or updates it if the ID already exists.
// Used by the library scanner to avoid duplicates on re-scans.
func (r *TrackRepo) Upsert(ctx context.Context, track *api.Track) error {
	query := `
		INSERT INTO tracks (id, title, artist, album, duration, file_path, genre, year, track_num, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			artist = EXCLUDED.artist,
			album = EXCLUDED.album,
			duration = EXCLUDED.duration,
			genre = EXCLUDED.genre,
			year = EXCLUDED.year,
			track_num = EXCLUDED.track_num`

	_, err := r.db.Pool.Exec(ctx, query,
		track.ID, track.Title, track.Artist, track.Album,
		int64(track.Duration), track.FilePath, track.Genre,
		track.Year, track.TrackNum, track.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("track_repo: upsert failed: %w", err)
	}
	return nil
}

// GetByID retrieves a single track by its ID.
func (r *TrackRepo) GetByID(ctx context.Context, id string) (*api.Track, error) {
	query := `SELECT id, title, artist, album, duration, file_path, genre, year, track_num, created_at
		FROM tracks WHERE id = $1`

	t := &api.Track{}
	var dur int64
	err := r.db.Pool.QueryRow(ctx, query, id).
		Scan(&t.ID, &t.Title, &t.Artist, &t.Album, &dur, &t.FilePath,
			&t.Genre, &t.Year, &t.TrackNum, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("track_repo: not found: %w", err)
	}
	t.Duration = time.Duration(dur)
	return t, nil
}

// GetAll retrieves all tracks, ordered by artist then title.
func (r *TrackRepo) GetAll(ctx context.Context) ([]*api.Track, error) {
	query := `SELECT id, title, artist, album, duration, file_path, genre, year, track_num, created_at
		FROM tracks ORDER BY artist, album, track_num, title`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("track_repo: list failed: %w", err)
	}
	defer rows.Close()

	var tracks []*api.Track
	for rows.Next() {
		t := &api.Track{}
		var dur int64
		if err := rows.Scan(&t.ID, &t.Title, &t.Artist, &t.Album, &dur, &t.FilePath,
			&t.Genre, &t.Year, &t.TrackNum, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("track_repo: scan failed: %w", err)
		}
		t.Duration = time.Duration(dur)
		tracks = append(tracks, t)
	}
	return tracks, rows.Err()
}

// Search performs a case-insensitive search across title, artist, and album.
// Uses PostgreSQL's ILIKE for efficient pattern matching.
func (r *TrackRepo) Search(ctx context.Context, query string) ([]*api.Track, error) {
	sql := `SELECT id, title, artist, album, duration, file_path, genre, year, track_num, created_at
		FROM tracks
		WHERE title ILIKE $1 OR artist ILIKE $1 OR album ILIKE $1
		ORDER BY artist, title
		LIMIT 100`

	pattern := "%" + strings.ReplaceAll(query, "%", "\\%") + "%"
	rows, err := r.db.Pool.Query(ctx, sql, pattern)
	if err != nil {
		return nil, fmt.Errorf("track_repo: search failed: %w", err)
	}
	defer rows.Close()

	var tracks []*api.Track
	for rows.Next() {
		t := &api.Track{}
		var dur int64
		if err := rows.Scan(&t.ID, &t.Title, &t.Artist, &t.Album, &dur, &t.FilePath,
			&t.Genre, &t.Year, &t.TrackNum, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("track_repo: scan failed: %w", err)
		}
		t.Duration = time.Duration(dur)
		tracks = append(tracks, t)
	}
	return tracks, rows.Err()
}

// GetByArtist returns all tracks by a given artist.
func (r *TrackRepo) GetByArtist(ctx context.Context, artist string) ([]*api.Track, error) {
	query := `SELECT id, title, artist, album, duration, file_path, genre, year, track_num, created_at
		FROM tracks WHERE artist = $1 ORDER BY album, track_num`

	rows, err := r.db.Pool.Query(ctx, query, artist)
	if err != nil {
		return nil, fmt.Errorf("track_repo: artist query failed: %w", err)
	}
	defer rows.Close()

	var tracks []*api.Track
	for rows.Next() {
		t := &api.Track{}
		var dur int64
		if err := rows.Scan(&t.ID, &t.Title, &t.Artist, &t.Album, &dur, &t.FilePath,
			&t.Genre, &t.Year, &t.TrackNum, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("track_repo: scan failed: %w", err)
		}
		t.Duration = time.Duration(dur)
		tracks = append(tracks, t)
	}
	return tracks, rows.Err()
}

// Count returns the total number of tracks in the database.
func (r *TrackRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM tracks`).Scan(&count)
	return count, err
}

// Delete removes a track by its ID.
func (r *TrackRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM tracks WHERE id = $1`, id)
	return err
}
