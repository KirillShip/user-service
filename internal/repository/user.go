package repository

import (
	"context"
	"errors"

	"github.com/KirillShip/user-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	const query = `
		INSERT INTO users (name, login, password, role)
		VALUES ($1, $2, $3)
		RETURNING id, name, login, role
	`
	err := r.db.QueryRow(
		ctx,
		query,
		user.Name,
		user.Login,
		user.Password,
		user.Role,
	).Scan(&user.ID, &user.Name, &user.Login, &user.Role)

	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	const query = `
		SELECT id, name, login, role
		FROM users
		WHERE id = $1 
	`
	var user model.User
	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(&user.ID, &user.Name, &user.Login, &user.Role)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	const query = `
		SELECT id, name, login, role, password 
		FROM users
		WHERE login = $1 
	`
	var user model.User
	err := r.db.QueryRow(
		ctx,
		query,
		login,
	).Scan(&user.ID, &user.Name, &user.Login, &user.Role, &user.Password)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *model.User) (*model.User, error) {
	const query = `
		UPDATE users 
		SET name = $1, password = $2
		WHERE id = $3
		RETURNING id, name, login
	`
	var updatedUser model.User
	err := r.db.QueryRow(
		ctx,
		query,
		user.Name,
		user.Password,
		user.ID,
	).Scan(
		&updatedUser.ID,
		&updatedUser.Name,
		&updatedUser.Login,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &updatedUser, nil
}

func (r *UserRepository) DeleteUserByID(ctx context.Context, id int64) error {
	const query = `
		DELETE FROM users 
		WHERE id = $1 
	`
	tag, err := r.db.Exec(ctx, query, id)

	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) DeleteUserByLogin(ctx context.Context, login string) error {
	const query = `
		DELETE FROM users 
		WHERE login = $1 
	`
	tag, err := r.db.Exec(ctx, query, login)

	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}
