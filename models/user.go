package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

type User struct {
	ID              string    `json:"id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	Password        string    `json:"-"`
	CreatedAt       time.Time `json:"created_at"`
	IsVerified      bool      `json:"is_verified"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	IsAnonymousUser bool      `json:"is_anonymous_user"`
}

type UserDetails struct {
	ID         string `json:"auth_user_id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	IsVerified bool   `json:"is_verified"`
}

type AnonymousUser struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{
		db: db,
	}
}

func InitUserSchema(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR(255),
			email VARCHAR(255) UNIQUE,
			password VARCHAR(255),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			is_verified BOOLEAN NOT NULL DEFAULT FALSE,
			first_name VARCHAR(225),
			last_name VARCHAR(225),
			is_anonymous_user BOOLEAN DEFAULT FALSE
		);
		
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`

	_, err := db.Exec(query)
	return err
}

func (s *UserStore) CreateUser(user *User) (string, error) {
	query := `INSERT INTO users (username, email, password, created_at) VALUES ($1, $2, $3, $4) RETURNING id`

	row := s.db.QueryRow(query, user.Username, user.Email, user.Password, user.CreatedAt)
	var id string
	err := row.Scan(&id)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return "", ErrUserExists
			}
		}
		return "", err
	}

	return id, nil
}

func (s *UserStore) CreateAnonymousUsers(users []AnonymousUser) error {
	if len(users) == 0 {
		return nil
	}

	query := `INSERT INTO users (first_name, last_name, email, is_anonymous_user) VALUES `
	values := make([]interface{}, 0, len(users)*4)

	for i, user := range users {
		base := i * 4

		query += fmt.Sprintf("($%d, $%d, $%d, $%d),", (base + 1), (base + 2), (base + 3), (base + 4))
		values = append(values, user.FirstName, user.LastName, user.Email, true)
	}
	query = strings.TrimSuffix(query, ",")

	_, err := s.db.Exec(query, values)
	return err
}

func (s *UserStore) AnonymousUserIdExists(userId string) bool {
	query := `SELECT id FROM users WHERE id = $1 and is_anonymous_user = TRUE`
	var id string
	err := s.db.QueryRow(query, userId).Scan(&id)

	if err != nil {
		return false
	}
	return id != ""
}

func (s *UserStore) GetAnonymousUserByEmail(email string) (string, error) {
	query := `SELECT id FROM users WHERE email = $1 and is_anonymous_user = TRUE LIMIT 1`

	var id string
	if err := s.db.QueryRow(query, email).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return "", ErrUserNotFound
		}
	}

	return id, nil
}

func (s *UserStore) UpdateUserSignupDetails(user *User, anonymousUserId string) error {
	query := `UPDATE users SET username = $1, email = $2, password = $3, is_verified = $4, is_anonymous_user = $5 WHERE id = $6`
	_, err := s.db.Exec(query, user.Username, user.Email, user.Password, false, false, anonymousUserId)
	return err
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

func (s *UserStore) GetUserDetailsById(userId string) (*UserDetails, error) {
	query := `SELECT id, username, email, is_verified FROM users WHERE id = $1`

	user := &UserDetails{}
	err := s.db.QueryRow(query, userId).Scan(&user.ID, &user.Username, &user.Email, &user.IsVerified)
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

func (s *UserStore) DeleteUserById(userId string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := s.db.Exec(query, userId)
	return err
}

func (s *UserStore) CreateSessionUser() (string, error) {
	query := `INSERT INTO users(is_anonymous_user) VALUES (TRUE) RETURNING id`
	row := s.db.QueryRow(query)
	var id string
	if err := row.Scan(&id); err != nil {
		return "", err
	}

	return id, nil
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
