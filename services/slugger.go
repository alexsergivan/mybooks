package services

import (
	"regexp"
	"strings"

	"github.com/alexsergivan/transliterator"
)

type Text string

func NormalizeForUrl(text, langcode string) string {
	s := Text(text)
	s.toLowerCase().transliterate(langcode).replaceAllNonCharactersSymbolsWithDashes()
	return string(s)
}

func (s *Text) toLowerCase() *Text {
	*s = Text(strings.ToLower(string(*s)))
	return s
}

func (s *Text) replaceAllNonCharactersSymbolsWithDashes() *Text {
	re := regexp.MustCompile(`[^a-z0-9/\p{L}]`)
	text := re.ReplaceAllString(string(*s), `-`)
	*s = Text(text)
	return s
}

func (s *Text) transliterate(langcode string) *Text {
	trans := transliterator.NewTransliterator(nil)
	*s = Text(trans.Transliterate(string(*s), langcode))
	return s
}
