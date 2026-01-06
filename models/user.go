package models

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type User struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Password   string    `json:"-"`
	CreatedAt  time.Time `json:"created_at"`
	IsVerified bool      `json:"is_verified"`
}

type UserDetails struct {
	ID         string `json:"auth_user_id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	IsVerified bool   `json:"is_verified"`
}

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{
		db: db,
	}
}

func InitDB(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			is_verified BOOLEAN NOT NULL DEFAULT FALSE
		);
		
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`

	_, err := db.Exec(query)
	return err
}

func (s *UserStore) CreateUser(user *User) error {
	query := `INSERT INTO users (username, email, password, created_at, is_verified) VALUES ($1, $2, $3, $4, $5)`

	_, err := s.db.Exec(query, user.Username, user.Email, user.Password, user.CreatedAt, false)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return ErrUserExists
			}
		}
		return err
	}

	return nil
}

func (s *UserStore) GetUserByEmail(email string) (*User, error) {
	query := `SELECT id, username, email, password, created_at, is_verified FROM users WHERE email = $1`

	user := &User{}
	err := s.db.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.IsVerified)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (s *UserStore) GetUserDetailsByEmail(email string) (*UserDetails, error) {
	query := `SELECT id, username, email, is_verified FROM users WHERE email = $1`

	user := &UserDetails{}
	err := s.db.QueryRow(query, email).Scan(&user.ID, &user.Username, &user.Email, &user.IsVerified)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserStore) UpdateUserVerification(email string) error {
	query := `UPDATE users SET is_verified = TRUE WHERE email = $1`

	result, err := s.db.Exec(query, email)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

var (
	ErrUserExists   = &Error{Message: "User already exists", Code: "USER_EXISTS"}
	ErrUserNotFound = &Error{Message: "User not found", Code: "USER_NOT_FOUND"}
)

type Error struct {
	Message string
	Code    string
}

func (e *Error) Error() string {
	return e.Message
}
