package tablenamer

import (
	"unicode"

	"github.com/samonzeweb/godb/tablenamer/plural"
)

// TableNamer function type is used set table naming function to build table name from struct's name
type TableNamerFn func(string, bool) string

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
	return func(name string, done bool) string {
		if done {
			return name
		}
		return ToSnakeCase(name)
	}
}

// SnakePlural builds table name from struct's name in plural snake format
func SnakePlural() TableNamerFn {
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

	return string(out)
}
