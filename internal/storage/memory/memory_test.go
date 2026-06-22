package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/maksim-mshp/ozon-internship-task/internal/links"
	"github.com/stretchr/testify/require"
)

func TestStorage_SaveAndGet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := New()

	link := links.Link{Code: "abc123", OriginalURL: "https://example.com"}
	require.NoError(t, s.Save(ctx, link))
	require.Equal(t, 1, s.Len())

	byCode, err := s.GetByCode(ctx, "abc123")
	require.NoError(t, err)
	require.Equal(t, link, byCode)

	byOriginal, err := s.GetByOriginal(ctx, "https://example.com")
	require.NoError(t, err)
	require.Equal(t, link, byOriginal)
}

func TestStorage_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := New()

	_, err := s.GetByCode(ctx, "missing")
	require.ErrorIs(t, err, links.ErrNotFound)

	_, err = s.GetByOriginal(ctx, "https://missing.example")
	require.ErrorIs(t, err, links.ErrNotFound)
}

func TestStorage_DuplicateOriginal(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := New()

	require.NoError(t, s.Save(ctx, links.Link{Code: "code1", OriginalURL: "https://example.com"}))

	err := s.Save(ctx, links.Link{Code: "code2", OriginalURL: "https://example.com"})
	require.ErrorIs(t, err, links.ErrOriginalExists)
	require.Equal(t, 1, s.Len())
}

func TestStorage_DuplicateCode(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := New()

	require.NoError(t, s.Save(ctx, links.Link{Code: "code1", OriginalURL: "https://a.example"}))

	err := s.Save(ctx, links.Link{Code: "code1", OriginalURL: "https://b.example"})
	require.ErrorIs(t, err, links.ErrCodeExists)
	require.Equal(t, 1, s.Len())
}

func TestStorage_ConcurrentDistinct(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := New()

	const n = 500
	var wg sync.WaitGroup
	errCh := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			link := links.Link{
				Code:        fmt.Sprintf("code%05d", i),
				OriginalURL: fmt.Sprintf("https://example.com/%d", i),
			}
			if err := s.Save(ctx, link); err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		require.NoError(t, err)
	}
	require.Equal(t, n, s.Len())
}

func TestStorage_ConcurrentSameOriginal(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := New()

	const n = 200
	var wg sync.WaitGroup
	var success int64
	badCh := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := s.Save(ctx, links.Link{
				Code:        fmt.Sprintf("code%05d", i),
				OriginalURL: "https://same.example",
			})
			switch {
			case err == nil:
				atomic.AddInt64(&success, 1)
			case errors.Is(err, links.ErrOriginalExists):
			default:
				badCh <- err
			}
		}()
	}
	wg.Wait()
	close(badCh)

	for err := range badCh {
		require.NoError(t, err)
	}
	require.Equal(t, int64(1), success)
	require.Equal(t, 1, s.Len())

	stored, err := s.GetByOriginal(ctx, "https://same.example")
	require.NoError(t, err)
	require.NotEmpty(t, stored.Code)
}
