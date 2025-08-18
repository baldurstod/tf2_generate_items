package main

import (
	"fmt"
	"strings"

	"github.com/baldurstod/vdf"
)

type language struct {
	lang   string
	tokens map[string]string
}

func (lg *language) init(path string) error {
	dat, err := ReadFileUTF16(path)
	if err != nil {
		return fmt.Errorf("unable to read file %s: %w", path, err)
	}

	v := vdf.VDF{}
	languageVdf := v.Parse(dat)

	lang, ok := languageVdf.Get("lang")
	if !ok {
		panic("lang key not found")
	}
	language, ok := lang.GetString("Language")
	if !ok {
		panic("Language key not found")
	}

	tokens, ok := lang.Get("Tokens")
	if !ok {
		panic("Tokens key not found")
	}

	lg.lang = language
	lg.tokens = make(map[string]string)
	for _, val := range tokens.Value.([]*vdf.KeyValue) {
		lg.tokens[strings.ToLower(val.Key)] = val.Value.(string)
	}
	return nil
}

func (lg *language) getToken(token string) (string, bool) {
	token = strings.TrimPrefix(token, "#")
	s, ok := lg.tokens[strings.ToLower(token)]
	return s, ok
}
