package main

import (
	//"os"
	"fmt"
	"strings"
	"github.com/baldurstod/vdf"
)

type language struct {
	lang string
	tokens itemStringMap
}

func (this *language) init(path string) {
	//dat, _ := os.ReadFile(path)
	dat, _ := ReadFileUTF16(path)
	vdf := vdf.VDF{}
	languageVdf := itemGameMap(vdf.Parse(dat))

	lang := getMap(getMap(languageVdf)["lang"])
	this.lang = (lang["Language"]).(string)
	this.tokens = make(itemStringMap)//getMap(lang["Tokens"])

	for key, val := range getMap(lang["Tokens"]) {
		this.tokens[strings.ToLower(key)] = val.(string)
	}

	fmt.Println(this.lang)
}


func (this *language) getToken(token string) (string, bool) {
	token = strings.TrimPrefix(token, "#")
	s, ok := this.tokens[strings.ToLower(token)]
	return s, ok
}