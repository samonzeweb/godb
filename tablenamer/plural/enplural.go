package plural

import (
	"regexp"
	"strings"
	"sync"
)

// regCache holds a compiled regex to match pluralized word
type regCache struct {
	Regexp   *regexp.Regexp
	Replacer string
}

// irregular holds irregular words
type irregular struct {
	Singular string
	Plural   string
}

// EnglishPlural defines methods and rules for getting plural form of a word for English
type EnglishPlural struct {
	plurals      []regCache
	uncountables map[string]bool
	irregulars   map[string]*irregular
	cache        map[string]string
	mu           sync.RWMutex // protects this struct's properties
}

// AddPluralRegex add regex rule for holding plural rule
func (ep *EnglishPlural) AddPluralRegex(matcherReg, replacerWord string) {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	ep.plurals = append(ep.plurals, regCache{
		Regexp:   regexp.MustCompile(matcherReg),
		Replacer: replacerWord,
	})
	ep.cache = map[string]string{}
}

// getIfPlural returns a non-empty string if the str is found in the Plurals, else empty string is returned
func (ep *EnglishPlural) getIfPlural(word string) string {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	for _, p := range ep.plurals {
		if p.Regexp.MatchString(word) {
			return p.Regexp.ReplaceAllString(word, p.Replacer)
		}
	}
	return ""
}

// AddPluralRegex add regex rule for holding plural rule
func (ep *EnglishPlural) AddIrregular(singularWord, pluralWord string) {
	irr := &irregular{
		Singular: singularWord,
		Plural:   pluralWord,
	}
	ep.mu.Lock()
	defer ep.mu.Unlock()

	ep.irregulars[strings.ToLower(singularWord)] = irr
	ep.irregulars[strings.ToLower(pluralWord)] = irr
	ep.cache = map[string]string{}
}

// getIfIrregular returns a non-empty string if the str is found in the Irregulars, else empty string is returned
func (ep *EnglishPlural) getIfIrregular(word string) string {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	if p, ok := ep.irregulars[strings.ToLower(word)]; ok {
		return p.Plural
	}
	return ""
}

// AddUncountable adds uncountable word
func (ep *EnglishPlural) AddUncountable(uncountable string) {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	ep.uncountables[uncountable] = true
	ep.cache = map[string]string{}
}

// isUncountable returns a true if the str is found in the Uncountables.
func (ep *EnglishPlural) isUncountable(word string) bool {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	if _, ok := ep.uncountables[word]; ok {
		return true
	}
	return false
}

func (ep *EnglishPlural) pluralFromCache(word string) string {
	ep.mu.RLock()
	defer ep.mu.RUnlock()
	if res, ok := ep.cache[word]; ok {
		return res
	}
	return ""
}

func (ep *EnglishPlural) plural(word string) string {
	if ep.isUncountable(word) {
		return word
	} else if irr := ep.getIfIrregular(word); irr != "" {
		return irr
	} else if p := ep.getIfPlural(word); p != "" {
		return p
	}
	return word
}

func (ep *EnglishPlural) pluralToCache(word, res string) {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	ep.cache[word] = res
}

// Plural returns plural form of a word
func (ep *EnglishPlural) Plural(word string) string {
	var res string
	if res = ep.pluralFromCache(word); res == "" {
		res = ep.plural(word)
		ep.pluralToCache(word, res)
	}
	return res
}

func EnglishPluralization() *EnglishPlural {
	ep := &EnglishPlural{
		plurals:    []regCache{},
		irregulars: map[string]*irregular{},
		cache:      map[string]string{},
		uncountables: map[string]bool{
			`advice`:      true,
			`art`:         true,
			`buffalo`:     true,
			`butter`:      true,
			`currency`:    true,
			`deer`:        true,
			`electricity`: true,
			`equipment`:   true,
			`fish`:        true,
			`furniture`:   true,
			`gas`:         true,
			`happiness`:   true,
			`information`: true,
			`jeans`:       true,
			`love`:        true,
			`luggage`:     true,
			`money`:       true,
			`music`:       true,
			`news`:        true,
			`police`:      true,
			`power`:       true,
			`rice`:        true,
			`salmon`:      true,
			`scenery`:     true,
			`series`:      true,
			`sheep`:       true,
			`species`:     true,
			`sugar`:       true,
			`trout`:       true,
			`tuna`:        true,
			`water`:       true,
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
