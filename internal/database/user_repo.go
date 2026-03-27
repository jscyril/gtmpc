package database

import (
	"context"
	"fmt"

	"github.com/jscyril/golang_music_player/api"
	"golang.org/x/crypto/bcrypt"
)

// UserRepo handles all user-related database operations.
// Passwords are always stored as bcrypt hashes — never in plaintext.
type UserRepo struct {
	db *DB
}

// NewUserRepo creates a new user repository backed by PostgreSQL.
func NewUserRepo(db *DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create inserts a new user with a bcrypt-hashed password.
// Returns ErrUserAlreadyExists if the username is taken (unique constraint).
func (r *UserRepo) Create(ctx context.Context, id, username, passwordHash, role string) (*api.User, error) {
	query := `
		INSERT INTO users (id, username, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, username, role, created_at`

	user := &api.User{}
	err := r.db.Pool.QueryRow(ctx, query, id, username, passwordHash, role).
		Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt)
	if err != nil {
		// Check for unique constraint violation on username
		if isDuplicateKeyError(err) {
			return nil, fmt.Errorf("username already taken")
		}
		return nil, fmt.Errorf("user_repo: create failed: %w", err)
	}
	user.PasswordHash = passwordHash
	return user, nil
}

// GetByUsername retrieves a user by username for authentication.
// The returned user includes the PasswordHash for bcrypt verification.
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*api.User, error) {
	query := `SELECT id, username, password_hash, role, created_at FROM users WHERE username = $1`

	user := &api.User{}
	err := r.db.Pool.QueryRow(ctx, query, username).
		Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("user_repo: user not found: %w", err)
	}
	return user, nil
}

// VerifyPassword compares a plaintext password with the user's bcrypt hash.
func (r *UserRepo) VerifyPassword(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}

// GetAll returns all users (without password hashes - for admin listing).
func (r *UserRepo) GetAll(ctx context.Context) ([]*api.User, error) {
	query := `SELECT id, username, role, created_at FROM users ORDER BY created_at DESC`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("user_repo: list failed: %w", err)
	}
	defer rows.Close()

	var users []*api.User
	for rows.Next() {
		u := &api.User{}
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("user_repo: scan failed: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// isDuplicateKeyError checks for PostgreSQL unique constraint violation (23505)
func isDuplicateKeyError(err error) bool {
	return err != nil && (contains(err.Error(), "23505") || contains(err.Error(), "duplicate key"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
