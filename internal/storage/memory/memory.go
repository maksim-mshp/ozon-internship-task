package memory

import (
	"context"
	"sync"

	"github.com/maksim-mshp/ozon-internship-task/internal/links"
)

type Storage struct {
	mu         sync.RWMutex
	byCode     map[string]string
	byOriginal map[string]string
}

func New() *Storage {
	return &Storage{
		byCode:     make(map[string]string),
		byOriginal: make(map[string]string),
	}
}

func (s *Storage) Save(_ context.Context, link links.Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.byOriginal[link.OriginalURL]; ok {
		return links.ErrOriginalExists
	}
	if _, ok := s.byCode[link.Code]; ok {
		return links.ErrCodeExists
	}

	s.byCode[link.Code] = link.OriginalURL
	s.byOriginal[link.OriginalURL] = link.Code
	return nil
}

func (s *Storage) GetByCode(_ context.Context, code string) (links.Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	original, ok := s.byCode[code]
	if !ok {
		return links.Link{}, links.ErrNotFound
	}
	return links.Link{Code: code, OriginalURL: original}, nil
}

func (s *Storage) GetByOriginal(_ context.Context, original string) (links.Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	code, ok := s.byOriginal[original]
	if !ok {
		return links.Link{}, links.ErrNotFound
	}
	return links.Link{Code: code, OriginalURL: original}, nil
}

func (s *Storage) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.byCode)
}
