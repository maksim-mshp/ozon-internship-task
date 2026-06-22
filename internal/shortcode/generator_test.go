package shortcode

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

const testAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_"

func TestNewGenerator_Validation(t *testing.T) {
	t.Parallel()

	_, err := NewGenerator("", 10)
	require.ErrorIs(t, err, ErrEmptyAlphabet)

	_, err = NewGenerator(testAlphabet, 0)
	require.ErrorIs(t, err, ErrInvalidLength)

	_, err = NewGenerator(testAlphabet, -5)
	require.ErrorIs(t, err, ErrInvalidLength)

	g, err := NewGenerator(testAlphabet, 10)
	require.NoError(t, err)
	require.Equal(t, 10, g.Length())
}

func TestGenerate_LengthAndAlphabet(t *testing.T) {
	t.Parallel()

	const length = 10
	g, err := NewGenerator(testAlphabet, length)
	require.NoError(t, err)

	for i := 0; i < 1000; i++ {
		code, err := g.Generate()
		require.NoError(t, err)
		require.Len(t, code, length)
		for _, r := range code {
			require.Truef(t, strings.ContainsRune(testAlphabet, r), "недопустимый символ %q в коде %q", r, code)
		}
	}
}

func TestGenerate_Uniqueness(t *testing.T) {
	t.Parallel()

	g, err := NewGenerator(testAlphabet, 10)
	require.NoError(t, err)

	const n = 10000
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		code, err := g.Generate()
		require.NoError(t, err)
		_, dup := seen[code]
		require.Falsef(t, dup, "коллизия на коде %q", code)
		seen[code] = struct{}{}
	}
}

func TestGenerate_CoversWholeAlphabet(t *testing.T) {
	t.Parallel()

	g, err := NewGenerator(testAlphabet, 16)
	require.NoError(t, err)

	used := make(map[rune]struct{})
	for i := 0; i < 5000; i++ {
		code, err := g.Generate()
		require.NoError(t, err)
		for _, r := range code {
			used[r] = struct{}{}
		}
	}
	require.Len(t, used, len([]rune(testAlphabet)), "не все символы алфавита встречаются в выборке")
}

func TestGenerate_ConcurrentSafe(t *testing.T) {
	t.Parallel()

	const length = 10
	g, err := NewGenerator(testAlphabet, length)
	require.NoError(t, err)

	const goroutines = 100
	const perGoroutine = 100
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				code, err := g.Generate()
				if err != nil {
					errs <- err
					return
				}
				if len(code) != length {
					errs <- fmt.Errorf("неожиданная длина кода: %d", len(code))
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
}
