// Package i18n provides a shared internationalization layer for all
// harns-platform services. Translation messages live in locales/*.yaml and are
// embedded into the binary at build time, so no external files are needed at
// runtime.
package i18n

import (
	"embed"
	"fmt"
	"path"
	"strings"
	"sync"

	"golang.org/x/text/language"
	"sigs.k8s.io/yaml"
)

//go:embed locales/*.yaml
var localeFS embed.FS

const defaultLang = "en-US"

var (
	mu       sync.RWMutex
	catalogs = map[string]map[string]string{}
	langs    []string
	matcher  language.Matcher
)

func init() {
	entries, err := localeFS.ReadDir("locales")
	if err != nil {
		panic(fmt.Sprintf("i18n: failed to read locales dir: %v", err))
	}

	tags := make([]language.Tag, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		lang := strings.TrimSuffix(entry.Name(), ".yaml")
		data, err := localeFS.ReadFile(path.Join("locales", entry.Name()))
		if err != nil {
			panic(fmt.Sprintf("i18n: failed to read locale %q: %v", lang, err))
		}
		messages := map[string]string{}
		if err := yaml.Unmarshal(data, &messages); err != nil {
			panic(fmt.Sprintf("i18n: failed to parse locale %q: %v", lang, err))
		}
		catalogs[lang] = messages
		langs = append(langs, lang)
		tags = append(tags, language.Make(lang))
	}

	if len(tags) == 0 {
		panic("i18n: no locale files found")
	}
	matcher = language.NewMatcher(tags)
}

// DefaultLang returns the fallback language used when no match is found.
func DefaultLang() string {
	return defaultLang
}

// Supported returns the list of supported language codes.
func Supported() []string {
	mu.RLock()
	defer mu.RUnlock()
	return append([]string(nil), langs...)
}

// Match resolves an Accept-Language header (or a bare language code) to one of
// the supported languages, falling back to the default language.
func Match(acceptLanguage string) string {
	if strings.TrimSpace(acceptLanguage) == "" {
		return defaultLang
	}
	prefs, _, err := language.ParseAcceptLanguage(acceptLanguage)
	if err != nil || len(prefs) == 0 {
		return defaultLang
	}
	_, idx, _ := matcher.Match(prefs...)
	if idx < 0 || idx >= len(langs) {
		return defaultLang
	}
	return langs[idx]
}

// Translate looks up key for lang and formats it with args. If the key is
// missing in lang it falls back to the default language, and finally to the key
// itself.
func Translate(lang, key string, args ...interface{}) string {
	tmpl := lookup(lang, key)
	if tmpl == "" {
		return key
	}
	if len(args) == 0 {
		return tmpl
	}
	return fmt.Sprintf(tmpl, args...)
}

func lookup(lang, key string) string {
	mu.RLock()
	defer mu.RUnlock()

	if catalog, ok := catalogs[lang]; ok {
		if msg, ok := catalog[key]; ok {
			return msg
		}
	}
	if lang != defaultLang {
		if catalog, ok := catalogs[defaultLang]; ok {
			return catalog[key]
		}
	}
	return ""
}
