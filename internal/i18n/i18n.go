package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed lang.json
var langFile embed.FS

var translations map[string]map[string]string
var currentLang = "en"

func init() {
	data, err := langFile.ReadFile("lang.json")
	if err != nil {
		panic(fmt.Sprintf("Language file missing (embed): %v", err))
	}
	if err := json.Unmarshal(data, &translations); err != nil {
		panic(fmt.Sprintf("Failed to parse language JSON: %v", err))
	}
}

func SetLanguage(lang string) {
	if _, ok := translations[lang]; ok {
		currentLang = lang
	}
}

func T(key string, args ...interface{}) string {
	if val, ok := translations[currentLang][key]; ok {
		return fmt.Sprintf(val, args...)
	}
	if val, ok := translations["en"][key]; ok {
		return fmt.Sprintf(val, args...)
	}
	return key
}
