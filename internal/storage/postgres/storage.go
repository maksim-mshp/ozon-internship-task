package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maksim-mshp/ozon-internship-task/internal/links"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Save(ctx context.Context, link links.Link) error {
	if _, err := s.db.Exec(ctx, saveLinkQuery, link.Code, link.OriginalURL); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "links_original_url_key":
				return links.ErrOriginalExists
			case "links_pkey":
				return links.ErrCodeExists
			}
		}

		return fmt.Errorf("save link: %w", err)
	}

	return nil
}

func (s *Storage) GetByCode(ctx context.Context, code string) (links.Link, error) {
	return s.getOne(ctx, getByCodeQuery, code)
}

func (s *Storage) GetByOriginal(ctx context.Context, original string) (links.Link, error) {
	return s.getOne(ctx, getByOriginalQuery, original)
}

func (s *Storage) getOne(ctx context.Context, query, arg string) (links.Link, error) {
	var link links.Link
	if err := s.db.QueryRow(ctx, query, arg).Scan(&link.Code, &link.OriginalURL); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return links.Link{}, links.ErrNotFound
		}

		return links.Link{}, fmt.Errorf("get link: %w", err)
	}

	return link, nil
}
