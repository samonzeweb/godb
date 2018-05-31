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
	uncountables sync.Map
	irregulars   sync.Map
	cache        sync.Map
	plurals      []regCache
	mu           sync.RWMutex // protects plurals
}

// AddPluralRegex add regex rule for holding plural rule
func (ep *EnglishPlural) AddPluralRegex(matcherReg, replacerWord string) {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	ep.plurals = append(ep.plurals, regCache{
		Regexp:   regexp.MustCompile(matcherReg),
 		Replacer: replacerWord,
	})
	ep.cache = sync.Map{}
}

// getIfPlural returns a non-empty string if the str is found in the Plurals, else empty string is returned
func (ep *EnglishPlural) getIfPlural(word string) (res string) {
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
	irr := irregular{
		Singular: singularWord,
		Plural:   pluralWord,
	}
	ep.irregulars.Store(strings.ToLower(singularWord), irr)
	ep.irregulars.Store(strings.ToLower(pluralWord), irr)
	ep.cache = sync.Map{}
}

// getIfIrregular returns a non-empty string if the str is found in the Irregulars, else empty string is returned
func (ep *EnglishPlural) getIfIrregular(word string) string {
	if p, ok := ep.irregulars.Load(strings.ToLower(word)); ok {
		return p.(irregular).Plural
	}
	return ""
}

// AddUncountable adds uncountable word
func (ep *EnglishPlural) AddUncountable(uncountable string) {
	ep.uncountables.Store(uncountable, true)
	ep.cache = sync.Map{}
}

// isUncountable returns a true if the str is found in the Uncountables.
func (ep *EnglishPlural) isUncountable(word string) bool {
	if _, ok := ep.uncountables.Load(word); ok {
		return true
	}
	return false
}

// pluralFromCache gets plural form of word from cache if exists else returns empty string
func (ep *EnglishPlural) pluralFromCache(word string) string {
	if res, ok := ep.cache.Load(word); ok {
		return res.(string)
	}
	return ""
}

// plural returns plural form of a word if matched any rule, if not returns word itself
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

// pluralToCache saves plural value of a word to cache
func (ep *EnglishPlural) pluralToCache(key, val string) {
	ep.cache.Store(key, val)
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

var ep *EnglishPlural
var once sync.Once

// EnglishPluralization returns EnglishPlural instance
func EnglishPluralization() *EnglishPlural {
	once.Do(func() {
		ep = &EnglishPlural{
			plurals:    []regCache{},
			irregulars: sync.Map{},
			cache:      sync.Map{},
			uncountables: sync.Map{},
		}
		ep.AddUncountable(`advice`)
		ep.AddUncountable(`art`)
		ep.AddUncountable(`buffalo`)
		ep.AddUncountable(`butter`)
		ep.AddUncountable(`currency`)
		ep.AddUncountable(`deer`)
		ep.AddUncountable(`electricity`)
		ep.AddUncountable(`equipment`)
		ep.AddUncountable(`fish`)
		ep.AddUncountable(`furniture`)
		ep.AddUncountable(`gas`)
		ep.AddUncountable(`happiness`)
		ep.AddUncountable(`information`)
		ep.AddUncountable(`jeans`)
		ep.AddUncountable(`love`)
		ep.AddUncountable(`luggage`)
		ep.AddUncountable(`money`)
		ep.AddUncountable(`music`)
		ep.AddUncountable(`news`)
		ep.AddUncountable(`police`)
		ep.AddUncountable(`power`)
		ep.AddUncountable(`rice`)
		ep.AddUncountable(`salmon`)
		ep.AddUncountable(`scenery`)
		ep.AddUncountable(`series`)
		ep.AddUncountable(`sheep`)
		ep.AddUncountable(`species`)
		ep.AddUncountable(`sugar`)
		ep.AddUncountable(`trout`)
		ep.AddUncountable(`tuna`)
		ep.AddUncountable(`water`)
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
	})
	return ep
}
