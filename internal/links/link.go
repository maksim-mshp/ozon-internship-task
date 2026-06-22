package links

import (
	"context"
	"errors"
)

type Link struct {
	Code        string
	OriginalURL string
}

var (
	ErrNotFound       = errors.New("link not found")
	ErrCodeExists     = errors.New("code already exists")
	ErrOriginalExists = errors.New("original url already exists")
)

type Storage interface {
	Save(ctx context.Context, link Link) error
	GetByCode(ctx context.Context, code string) (Link, error)
	GetByOriginal(ctx context.Context, original string) (Link, error)
}
