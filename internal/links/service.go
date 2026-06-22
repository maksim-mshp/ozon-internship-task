package links

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

var (
	ErrInvalidURL     = errors.New("invalid url")
	ErrCodeGeneration = errors.New("failed to generate unique code")
)

type CodeGenerator interface {
	Generate() (string, error)
}

type Service struct {
	storage     Storage
	generator   CodeGenerator
	maxAttempts int
}

func NewService(storage Storage, generator CodeGenerator, maxAttempts int) *Service {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	return &Service{
		storage:     storage,
		generator:   generator,
		maxAttempts: maxAttempts,
	}
}

func (s *Service) Shorten(ctx context.Context, rawURL string) (Link, error) {
	original, err := normalizeURL(rawURL)
	if err != nil {
		return Link{}, err
	}

	existing, err := s.storage.GetByOriginal(ctx, original)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return Link{}, err
	}

	for attempt := 0; attempt < s.maxAttempts; attempt++ {
		code, genErr := s.generator.Generate()
		if genErr != nil {
			return Link{}, genErr
		}

		link := Link{Code: code, OriginalURL: original}
		switch saveErr := s.storage.Save(ctx, link); {
		case saveErr == nil:
			return link, nil
		case errors.Is(saveErr, ErrOriginalExists):
			return s.storage.GetByOriginal(ctx, original)
		case errors.Is(saveErr, ErrCodeExists):
			continue
		default:
			return Link{}, saveErr
		}
	}

	return Link{}, ErrCodeGeneration
}

func (s *Service) Resolve(ctx context.Context, code string) (Link, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return Link{}, ErrNotFound
	}
	return s.storage.GetByCode(ctx, code)
}

func normalizeURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrInvalidURL
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || !parsed.IsAbs() || parsed.Host == "" {
		return "", ErrInvalidURL
	}
	return trimmed, nil
}
