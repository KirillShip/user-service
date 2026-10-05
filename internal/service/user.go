package service

import (
	"context"
	"errors"
	"strings"

	"github.com/KirillShip/user-service/internal/model"
	"github.com/KirillShip/user-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrLoginInvalid       = errors.New("invalid login")
	ErrNameInvalid        = errors.New("invalid user name")
	ErrPasswordInvalid    = errors.New("invalid password")
	ErrIDInvalid          = errors.New("invalid id")
	ErrUserNotFound       = errors.New("user not found")
	ErrLoginAlreadyUsed   = errors.New("login already used")
	ErrInvalidCredentials = errors.New("invalid login or password")
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) (*model.User, error)
	GetUserByID(ctx context.Context, id int64) (*model.User, error)
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) (*model.User, error)
	DeleteUserByID(ctx context.Context, id int64) error
	DeleteUserByLogin(ctx context.Context, login string) error
}

type UserService struct {
	repository UserRepository
}

func NewUserService(repository UserRepository) *UserService {
	return &UserService{
		repository: repository,
	}
}

func (s *UserService) CreateUser(ctx context.Context, name string, login string, password string) (*model.User, error) {
	name = strings.TrimSpace(name)
	login = strings.TrimSpace(login)

	if name == "" {
		return nil, ErrNameInvalid
	}
	if login == "" {
		return nil, ErrLoginInvalid
	}

	if _, err := s.repository.GetUserByLogin(ctx, login); err == nil {
		return nil, ErrLoginAlreadyUsed
	}

	passwordLength := len([]byte(password))

	if passwordLength < 8 || passwordLength > 72 {
		return nil, ErrPasswordInvalid
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Name:     name,
		Login:    login,
		Role:     model.RoleUser,
		Password: string(hashedPassword),
	}
	createdUser, err := s.repository.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

func (s *UserService) CheckUser(ctx context.Context, login string, password string) (*model.User, error) {
	login = strings.TrimSpace(login)
	if login == "" {
		return nil, ErrInvalidCredentials
	}
	user, err := s.repository.GetUserByLogin(ctx, login)
	if errors.Is(err, repository.ErrUserNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	} else {
		return user, nil
	}
}

func (s *UserService) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	if id <= 0 {
		return nil, ErrIDInvalid
	}
	user, err := s.repository.GetUserByID(ctx, id)
	if errors.Is(err, repository.ErrUserNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	login = strings.TrimSpace(login)
	if login == "" {
		return nil, ErrLoginInvalid
	}
	return s.repository.GetUserByLogin(ctx, login)
}

func (s *UserService) UpdateUser(ctx context.Context, id int64, name string, password string) (*model.User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrNameInvalid
	}
	if id <= 0 {
		return nil, ErrIDInvalid
	}

	passwordLength := len([]byte(password))
	if passwordLength < 8 || passwordLength > 72 {
		return nil, ErrPasswordInvalid
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		ID:       id,
		Name:     name,
		Password: string(hashedPassword),
	}
	updatedUser, err := s.repository.UpdateUser(ctx, user)
	if errors.Is(err, repository.ErrUserNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return updatedUser, nil
}

func (s *UserService) DeleteUserByID(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrIDInvalid
	}
	err := s.repository.DeleteUserByID(ctx, id)
	if errors.Is(err, repository.ErrUserNotFound) {
		return ErrUserNotFound
	}
	if err != nil {
		return err
	}
	return nil
}

func (s *UserService) DeleteUserByLogin(ctx context.Context, login string) error {
	login = strings.TrimSpace(login)
	if login == "" {
		return ErrLoginInvalid
	}
	return s.repository.DeleteUserByLogin(ctx, login)
}
