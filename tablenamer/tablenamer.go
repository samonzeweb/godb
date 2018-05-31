package tablenamer

import (
	"unicode"

	"github.com/samonzeweb/godb/tablenamer/plural"
	"sync"
)

// TableNamerFn function type is used set table naming function to build table name from struct's name
type TableNamerFn func(string, bool) string

// sCache is cache for ToSnakeCase function
var sCache = sync.Map{}

// Plural builds table name as plural form of struct's name
func Plural() TableNamerFn {
	return func(name string, done bool) string {
		if done {
			return name
		}
		return plural.EnglishPluralization().Plural(name)
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
	sCache = sync.Map{}
	return func(name string, done bool) string {
		if done {
			return name
		}
		return ToSnakeCase(name)
	}
}

// SnakePlural builds table name from struct's name in plural snake format
func SnakePlural() TableNamerFn {
	sCache = sync.Map{}
	return func(name string, done bool) string {
		if done {
			return name
		}
		return plural.EnglishPluralization().Plural(ToSnakeCase(name))
	}
}

// ToSnakeCase converts a string to snake case, used for converting struct name to snake_case
func ToSnakeCase(s string) string {
	if res, ok := sCache.Load(s); ok {
		return res.(string)
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
	sCache.Store(s, res)
	return res
}
