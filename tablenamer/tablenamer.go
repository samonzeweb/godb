package tablenamer

import (
	"unicode"

	"github.com/samonzeweb/godb/tablenamer/plural"
)

// TableNamer function type is used set table naming function to build table name from struct's name
type TableNamerFn func(string, bool) string

// SetTableNamerPlural builds table name as plural form of struct's name
func TableNamerPlural() TableNamerFn {
	fn := plural.EnglishPluralization().Plural
	return func(name string, done bool) string {
		if done {
			return name
		}
		return fn(name)
	}
}

// SetTableNamerSame builds table name same as struct's name
func TableNamerSame() TableNamerFn {
	return func(name string, done bool) string {
		if done {
			return name
		}
		return name
	}
}

// SetTableNamerSnake builds table name from struct's name in snake format
func TableNamerSnake() TableNamerFn {
	return func(name string, done bool) string {
		if done {
			return name
		}
		return ToSnakeCase(name)
	}
}

// SetTableNamerSnake builds table name from struct's name in plural snake format
func TableNamerSnakePlural() TableNamerFn {
	fn := plural.EnglishPluralization().Plural
	return func(name string, done bool) string {
		if done {
			return name
		}
		return fn(ToSnakeCase(name))
	}
}

// Converts a string to snake case, used for converting struct name to snake_case
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
