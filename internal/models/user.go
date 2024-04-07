package models

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/hunterwilkins2/trolly/internal/validator"
)

var (
	ErrDuplicateEmail = errors.New("email address already exists")
	ErrUserNotFound   = errors.New("user not found")
)

type User struct {
	ID             uuid.UUID
	Name           string
	Email          string
	Password       string
	HashedPassword []byte
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
	stmt := `INSERT INTO users (id, name, email, hashed_password)
	VALUES(?, ?, ?, ?)`

	_, err := r.db.Exec(stmt, user.ID, user.Name, user.Email, string(user.HashedPassword))
	if err != nil {
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email") {
				return ErrDuplicateEmail
			}
		}
		return err
	}
	return nil
}

func (r *UserRepository) Get(ctx context.Context, email string) (*User, error) {
	stmt := `SELECT id, name, email, hashed_password
	FROM users
	WHERE email = ?`

	user := &User{}
	err := r.db.QueryRowContext(ctx, stmt, email).Scan(&user.ID, &user.Name, &user.Email, &user.HashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		} else {
			return nil, err
		}
	}
	return user, nil
}

func (r *UserRepository) GetById(ctx context.Context, id uuid.UUID) (*User, error) {
	stmt := `SELECT id, name, email, hashed_password
	FROM users
	WHERE id = ?`

	user := &User{}
	err := r.db.QueryRowContext(ctx, stmt, id).Scan(&user.ID, &user.Name, &user.Email, &user.HashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		} else {
			return nil, err
		}
	}
	return user, nil
}

func (u *User) Validate() error {
	v := validator.New()

	ValidateName(v, u.Name)
	ValidateEmail(v, u.Email)
	ValidatePassword(v, u.Password)
	if v.HasErrors() {
		return v
	}
	return nil
}

func ValidateName(v *validator.Validator, name string) {
	v.Check(len(name) == 0, "name", "Name cannot be empty")
	v.Check(len(name) > 255, "name", "Name cannot be more than 255 characters")
}

func ValidateEmail(v *validator.Validator, email string) {
	v.MatchesRX(validator.EmailRX, email, "email", "Must be a valid email")
}

func ValidatePassword(v *validator.Validator, password string) {
	v.Check(len(password) < 6, "password", "Password must be at least 6 characters")
	v.Check(len(password) > 72, "password", "Password must be less than 72 characters")
}
