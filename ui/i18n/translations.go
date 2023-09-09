package i18n

import "fmt"

func TranslationsFromMap(defaultLanguage string, languages map[string]Entries) *Translations {
	result := NewTranslations()
	for code, entries := range languages {
		result.AddLanguage(code, entries)
	}
	result.SetDefaultLanguage(defaultLanguage)
	return result
}

func NewTranslations() *Translations {
	return &Translations{
		languages: make(map[string]Entries),
	}
}

type Translations struct {
	defaultLanguage Entries
	languages       map[string]Entries
}

func (b *Translations) AddLanguage(code string, entries Entries) {
	b.languages[code] = entries
	if b.defaultLanguage == nil {
		b.defaultLanguage = entries
	}
}

func (b *Translations) SetDefaultLanguage(code string) {
	b.defaultLanguage = b.languages[code]
}

func (b *Translations) Translate(code, key string, args ...any) string {
	language, ok := b.languages[code]
	if !ok {
		language = b.defaultLanguage
	}
	if language == nil {
		return "N/A"
	}
	template, ok := language[key]
	if !ok {
		return "N/A"
	}
	return fmt.Sprintf(template, args...)
}

type Entries map[string]string
