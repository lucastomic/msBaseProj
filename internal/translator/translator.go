package translator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucastomic/msBaseProj/internal/contextypes"
)

var (
	translations map[string]map[string]string
	defaultLang  = "en"
)

func init() {
	translations = make(map[string]map[string]string)
	loadTranslations("en")
	loadTranslations("es")
}

func Translate(lang string, key string) string {
	if trans, ok := translations[lang]; ok {
		if val, ok := trans[key]; ok {
			return val
		}
	}
	return key
}

func TranslateGivenCtx(ctx context.Context, key string) string {
	lang := ctx.Value(contextypes.ContextLangKey{}).(string)
	return Translate(lang, key)
}

func loadTranslations(lang string) {
	filePath := filepath.Join("locales", fmt.Sprintf("%s.json", lang))
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("error loading translation %s", lang))
	}
	var trans map[string]string
	if err := json.Unmarshal(bytes, &trans); err != nil {
		panic(fmt.Sprintf("error loading translation %s", lang))
	}
	translations[lang] = trans
}
