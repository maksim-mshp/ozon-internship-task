package links

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type mapStorage struct {
	byCode     map[string]string
	byOriginal map[string]string
}

func newMapStorage() *mapStorage {
	return &mapStorage{
		byCode:     make(map[string]string),
		byOriginal: make(map[string]string),
	}
}

func (m *mapStorage) Save(_ context.Context, link Link) error {
	if _, ok := m.byOriginal[link.OriginalURL]; ok {
		return ErrOriginalExists
	}
	if _, ok := m.byCode[link.Code]; ok {
		return ErrCodeExists
	}
	m.byCode[link.Code] = link.OriginalURL
	m.byOriginal[link.OriginalURL] = link.Code
	return nil
}

func (m *mapStorage) GetByCode(_ context.Context, code string) (Link, error) {
	original, ok := m.byCode[code]
	if !ok {
		return Link{}, ErrNotFound
	}
	return Link{Code: code, OriginalURL: original}, nil
}

func (m *mapStorage) GetByOriginal(_ context.Context, original string) (Link, error) {
	code, ok := m.byOriginal[original]
	if !ok {
		return Link{}, ErrNotFound
	}
	return Link{Code: code, OriginalURL: original}, nil
}

type fixedGenerator struct {
	code string
}

func (g fixedGenerator) Generate() (string, error) {
	return g.code, nil
}

type scriptedGenerator struct {
	codes []string
	idx   int
}

func (g *scriptedGenerator) Generate() (string, error) {
	if g.idx >= len(g.codes) {
		return "", errors.New("scripted generator exhausted")
	}
	code := g.codes[g.idx]
	g.idx++
	return code, nil
}

type errGenerator struct {
	err error
}

func (g errGenerator) Generate() (string, error) {
	return "", g.err
}

type raceStorage struct {
	calls    int
	existing Link
}

func (s *raceStorage) Save(_ context.Context, _ Link) error {
	return ErrOriginalExists
}

func (s *raceStorage) GetByCode(_ context.Context, _ string) (Link, error) {
	return Link{}, ErrNotFound
}

func (s *raceStorage) GetByOriginal(_ context.Context, _ string) (Link, error) {
	s.calls++
	if s.calls == 1 {
		return Link{}, ErrNotFound
	}
	return s.existing, nil
}

func TestService_ShortenNew(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := NewService(newMapStorage(), fixedGenerator{code: "ABCDEFGHIJ"}, 5)

	link, err := svc.Shorten(ctx, "https://example.com/page")
	require.NoError(t, err)
	require.Equal(t, "ABCDEFGHIJ", link.Code)
	require.Equal(t, "https://example.com/page", link.OriginalURL)

	resolved, err := svc.Resolve(ctx, "ABCDEFGHIJ")
	require.NoError(t, err)
	require.Equal(t, "https://example.com/page", resolved.OriginalURL)
}

func TestService_ShortenDeduplicatesSameURL(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := NewService(newMapStorage(), fixedGenerator{code: "ABCDEFGHIJ"}, 5)

	first, err := svc.Shorten(ctx, "https://example.com")
	require.NoError(t, err)
	second, err := svc.Shorten(ctx, "https://example.com")
	require.NoError(t, err)
	require.Equal(t, first.Code, second.Code)
}

func TestService_ShortenValidatesURL(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := NewService(newMapStorage(), fixedGenerator{code: "ABCDEFGHIJ"}, 5)

	cases := []string{"", "   ", "not a url", "/relative/path", "example.com"}
	for _, raw := range cases {
		_, err := svc.Shorten(ctx, raw)
		require.ErrorIsf(t, err, ErrInvalidURL, "input %q", raw)
	}
}

func TestService_ResolveNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	svc := NewService(newMapStorage(), fixedGenerator{code: "X"}, 5)

	_, err := svc.Resolve(ctx, "missing")
	require.ErrorIs(t, err, ErrNotFound)

	_, err = svc.Resolve(ctx, "   ")
	require.ErrorIs(t, err, ErrNotFound)
}

func TestService_RetriesOnCodeCollision(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	storage := newMapStorage()
	require.NoError(t, storage.Save(ctx, Link{Code: "TAKEN00000", OriginalURL: "https://taken.example"}))

	gen := &scriptedGenerator{codes: []string{"TAKEN00000", "FREE000001"}}
	svc := NewService(storage, gen, 5)

	link, err := svc.Shorten(ctx, "https://new.example")
	require.NoError(t, err)
	require.Equal(t, "FREE000001", link.Code)
	require.Equal(t, 2, gen.idx)
}

func TestService_ExhaustsAttempts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	storage := newMapStorage()
	require.NoError(t, storage.Save(ctx, Link{Code: "TAKEN00000", OriginalURL: "https://taken.example"}))

	svc := NewService(storage, fixedGenerator{code: "TAKEN00000"}, 3)

	_, err := svc.Shorten(ctx, "https://new.example")
	require.ErrorIs(t, err, ErrCodeGeneration)
}

func TestService_ReturnsExistingOnOriginalRace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	existing := Link{Code: "RACECODE00", OriginalURL: "https://race.example"}
	svc := NewService(&raceStorage{existing: existing}, fixedGenerator{code: "NEWCODE000"}, 5)

	link, err := svc.Shorten(ctx, "https://race.example")
	require.NoError(t, err)
	require.Equal(t, existing, link)
}

func TestService_PropagatesGeneratorError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sentinel := errors.New("boom")
	svc := NewService(newMapStorage(), errGenerator{err: sentinel}, 5)

	_, err := svc.Shorten(ctx, "https://example.com")
	require.ErrorIs(t, err, sentinel)
}
