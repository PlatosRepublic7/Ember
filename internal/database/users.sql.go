// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: users.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, username, password)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at, updated_at, username, password
`

type CreateUserParams struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Username  string
	Password  string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.Username,
		arg.Password,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Username,
		&i.Password,
	)
	return i, err
}

const getUserByUsername = `-- name: GetUserByUsername :one
SELECT id, created_at, updated_at, username, password FROM users WHERE username = $1
`

func (q *Queries) GetUserByUsername(ctx context.Context, username string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByUsername, username)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Username,
		&i.Password,
	)
	return i, err
}

const getUserLoginInfo = `-- name: GetUserLoginInfo :one
SELECT id, username, password FROM users WHERE username = $1
`

type GetUserLoginInfoRow struct {
	ID       uuid.UUID
	Username string
	Password string
}

func (q *Queries) GetUserLoginInfo(ctx context.Context, username string) (GetUserLoginInfoRow, error) {
	row := q.db.QueryRowContext(ctx, getUserLoginInfo, username)
	var i GetUserLoginInfoRow
	err := row.Scan(&i.ID, &i.Username, &i.Password)
	return i, err
}
