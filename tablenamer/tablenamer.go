package tablenamer

import (
	"unicode"

	"github.com/samonzeweb/godb/tablenamer/plural"
	"sync"
)

// TableNamer function type is used set table naming function to build table name from struct's name
type TableNamerFn func(string, bool) string

// snakeCache is cacher for ToSnakeCase function
type snakeCache struct {
	sync.RWMutex
	cache map[string]string
}

// sCache is cache for ToSnakeCase function
var sCache = snakeCache{cache: map[string]string{} }

// Get gets a value from cache if found, else returns empty string
func (s *snakeCache) Get(word string) string {
	s.RLock()
	defer s.RUnlock()
	if s, ok := s.cache[word]; ok {
		return s
	}
	return ""
}

// Set sets a new value into cache
func (s *snakeCache) Set(key, newVal string) {
	s.Lock()
	defer s.Unlock()
	s.cache[key] = newVal
}


// Clear clears cache
func (s *snakeCache) Clear() {
	s.Lock()
	defer s.Unlock()
	s.cache = map[string]string{}
}

// Plural builds table name as plural form of struct's name
func Plural() TableNamerFn {
	fn := plural.EnglishPluralization().Plural
	return func(name string, done bool) string {
		if done {
			return name
		}
		return fn(name)
	}
}

// Same builds table name same as struct's name
func Same() TableNamerFn {
	return func(name string, done bool) string {
		if done {
			return name
		}
		return name
	}
}

// Snake builds table name from struct's name in snake format
func Snake() TableNamerFn {
	sCache.Clear()
	return func(name string, done bool) string {
		if done {
			return name
		}
		return ToSnakeCase(name)
	}
}

// SnakePlural builds table name from struct's name in plural snake format
func SnakePlural() TableNamerFn {
	sCache.Clear()
	fn := plural.EnglishPluralization().Plural
	return func(name string, done bool) string {
		if done {
			return name
		}
		return fn(ToSnakeCase(name))
	}
}

// ToSnakeCase converts a string to snake case, used for converting struct name to snake_case
func ToSnakeCase(s string) string {
	if res := sCache.Get(s); res != "" {
		return s
	}
	in := []rune(s)
	isLower := func(idx int) bool {
		return idx >= 0 && idx < len(in) && unicode.IsLower(in[idx])
	}

	out := make([]rune, 0, len(in) + len(in) / 2)
	for i, r := range in {
		if unicode.IsUpper(r) {
			r = unicode.ToLower(r)
			if i > 0 && in[i - 1] != '_' && (isLower(i - 1) || isLower(i + 1)) {
				out = append(out, '_')
			}
		}
		out = append(out, r)
	}

	res := string(out)
	sCache.Set(s, res)
	return res
}
