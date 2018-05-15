package plural

import (
	"regexp"
	"strings"
)

// RegCache holds a compiled regex to match pluralized word
type RegCache struct {
	Regexp   *regexp.Regexp
	Replacer string
}

// Irregular holds irregular words
type Irregular struct {
	Singular string
	Plural   string
}

// EnglishPlural defines methods and rules for getting plural form of a word for English
type EnglishPlural struct {
	Plurals      []RegCache
	Uncountables []string
	Irregulars   []Irregular
}

// AddPluralRegex add regex rule for holding plural rule
func (ep *EnglishPlural) AddPluralRegex(matcherReg, replacerWord string) {
	ep.Plurals = append(ep.Plurals, RegCache{
		Regexp:   regexp.MustCompile(matcherReg),
		Replacer: replacerWord,
	})
}

// GetPlural returns a non-empty string if the str is found in the Plurals, else empty string is returned
func (ep *EnglishPlural) GetIfPlural(word string) string {
	for _, p := range ep.Plurals {
		if p.Regexp.MatchString(word) {
			return p.Regexp.ReplaceAllString(word, p.Replacer)
		}
	}
	return ""
}

// AddPluralRegex add regex rule for holding plural rule
func (ep *EnglishPlural) AddIrregular(singularWord, pluralWord string) {
	ep.Irregulars = append(ep.Irregulars, Irregular{
		Singular: singularWord,
		Plural:   pluralWord,
	})
}

// GetIfIrregular returns a non-empty string if the str is found in the Irregulars, else empty string is returned
func (ep *EnglishPlural) GetIfIrregular(word string) string {
	word = strings.ToLower(word)
	for _, irr := range ep.Irregulars {
		if strings.ToLower(irr.Singular) == word || strings.ToLower(irr.Plural) == word {
			return irr.Plural
		}
	}
	return ""
}

// AddUncountable adds uncountable word
func (ep *EnglishPlural) AddUncountable(uncountable string) {
	ep.Uncountables = append(ep.Uncountables, uncountable)
}

// IsUncountable returns a true if the str is found in the Uncountables.
func (ep *EnglishPlural) IsUncountable(word string) bool {
	for _, w := range ep.Uncountables {
		if w == word {
			return true
		}
	}
	return false
}

// Plural returns plural form of a word
func (ep *EnglishPlural) Plural(word string) string {
	if ep.IsUncountable(word) {
		return word
	} else if irr := ep.GetIfIrregular(word); irr != "" {
		return irr
	} else if p := ep.GetIfPlural(word); p != "" {
		return p
	}
	return word
}

func EnglishPluralization() *EnglishPlural {
	ep := &EnglishPlural{
		Plurals:    []RegCache{},
		Irregulars: []Irregular{},
		Uncountables: []string{
			`advice`,
			`art`,
			`buffalo`,
			`butter`,
			`currency`,
			`deer`,
			`electricity`,
			`equipment`,
			`fish`,
			`furniture`,
			`gas`,
			`happiness`,
			`information`,
			`jeans`,
			`love`,
			`luggage`,
			`money`,
			`music`,
			`news`,
			`police`,
			`power`,
			`rice`,
			`salmon`,
			`scenery`,
			`series`,
			`sheep`,
			`species`,
			`sugar`,
			`trout`,
			`tuna`,
			`water`,
		},
	}

	ep.AddIrregular(`child`, `children`)
	ep.AddIrregular(`person`, `people`)
	ep.AddIrregular(`man`, `men`)
	ep.AddIrregular(`staff`, `staves`)
	ep.AddIrregular(`turf`, `trueves`)
	ep.AddIrregular(`goose`, `geese`)
	ep.AddPluralRegex(`(auto)$`, `${1}s`)
	ep.AddPluralRegex(`(quiz)$`, `${1}zes`)
	ep.AddPluralRegex(`(dwarf)$`, `${1}es`)
	ep.AddPluralRegex(`(matr|vert|ind)(ix|ex)$`, `${1}ices`)
	ep.AddPluralRegex(`(matr|vert|ind)(ix|ex)$`, `${1}ices`)
	ep.AddPluralRegex(`(alumn|bacill|cact|foc|fung|nucle|radi|stimul|syllab|termin|vir)us$`, `${1}i`)
	ep.AddPluralRegex(`(s|ss|sh|ch|x|to|ro|ho|jo|no)$`, `${1}es`)
	ep.AddPluralRegex(`(i)fe$`, `${1}ves`)
	ep.AddPluralRegex(`(t|f|g)oo(th|se|t)$`, `${1}ee${2}`)
	ep.AddPluralRegex(`(a|e|i|o|u)y$`, `${1}ys`)
	ep.AddPluralRegex(`(m|l)ouse$`, `${1}ice`)
	ep.AddPluralRegex(`(al|ie|l)f$`, `${1}ves`)
	ep.AddPluralRegex(`(d)ice`, `${1}ie`)
	ep.AddPluralRegex(`y$`, `ies`)
	ep.AddPluralRegex(`$`, `s`)
	return ep
}
