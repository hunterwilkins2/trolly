package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/hunterwilkins2/trolly/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserService struct {
	repository *models.UserRepository
}

func NewUserService(repository *models.UserRepository) *UserService {
	return &UserService{
		repository: repository,
	}
}

func (s *UserService) Register(ctx context.Context, name, email, password string) (*models.User, error) {
	user := &models.User{
		ID:       uuid.New(),
		Name:     name,
		Email:    email,
		Password: password,
	}
	err := user.Validate()
	if err != nil {
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("unable to hash password: %v", err)
	}
	user.HashedPassword = hashed
	err = s.repository.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserService) Login(ctx context.Context, email, password string) (*models.User, error) {
	user, err := r.repository.Get(ctx, email)
	if err == models.ErrUserNotFound {
		return nil, ErrInvalidCredentials
	} else if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.repository.GetById(ctx, id)
	if err != nil {
		return nil, models.ErrUserNotFound
	}
	return user, err
}
