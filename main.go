package main

import (
	"encoding/json"
	"os"
	"log"
	"flag"
	"fmt"
)
var lg language

func main() {
	var lang string
	var outputFolder string
	var inputFolder string
	var medals bool

	flag.StringVar(&lang, "l", "english", "Language")
	flag.BoolVar(&medals, "m", false, "Tournament medals")
	flag.StringVar(&outputFolder, "o", "", "Output folder")
	flag.StringVar(&inputFolder, "i", "", "Input folder")
	flag.Parse()

	if inputFolder == "" {
		fmt.Println("No input folder provided. Use the flag -i")
		os.Exit(1)
	}
	if outputFolder == "" {
		fmt.Println("No output folder provided. Use the flag -o")
		os.Exit(1)
	}

	file, _ := os.OpenFile("var/log.log", os.O_WRONLY|os.O_CREATE, 0644)
	log.SetOutput(file)

	lg = language{}
	lg.init(inputFolder + "tf_" + lang + ".txt")

	ig := itemsGame{}
	ig.medals = medals
	ig.init(inputFolder + "items_game.txt", inputFolder + "static.json")
	j, _ := json.MarshalIndent(&ig, "", "\t")

	var prefix string
	if medals {
		prefix = "medals"
	} else {
		prefix = "items"
	}
	os.WriteFile(outputFolder + prefix + "_" + lang + ".json", j, 0666)
}

func getMap(i interface{}) itemGameMap {
	switch i.(type) {
	case itemGameMap: return i.(itemGameMap)
	case map[string]interface{}: return itemGameMap((i).(map[string]interface{}))
	default: panic("Unknown type")
	}
}

func getStringToken(token string) string {
	s, exist := lg.getToken(token)

	if (exist) {
		return s
	} else {
		return token
	}
}

func getStringTokenRaw(token string) (string, bool) {
	return lg.getToken(token)
}
