package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
)

var lg language

func main() {
	var lang string
	var outputFolder string
	var itemsFolder string
	var resourceFolder string
	var staticFile string
	var medals bool

	flag.StringVar(&lang, "l", "english", "Language")
	flag.BoolVar(&medals, "m", false, "Tournament medals")
	flag.StringVar(&outputFolder, "o", "", "Output folder")
	flag.StringVar(&itemsFolder, "i", "", "Items folder")
	flag.StringVar(&resourceFolder, "r", "", "Resource folder")
	flag.StringVar(&staticFile, "s", "", "Static file")
	flag.Parse()

	if itemsFolder == "" {
		fmt.Println("No items folder provided. Use the flag -i")
		os.Exit(1)
	}
	if resourceFolder == "" {
		fmt.Println("No resource folder provided. Use the flag -r")
		os.Exit(1)
	}
	if outputFolder == "" {
		fmt.Println("No output folder provided. Use the flag -o")
		os.Exit(1)
	}

	var err error
	file, err := os.OpenFile("var/log.log", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	log.SetOutput(file)

	lg = language{}
	if err := lg.init(path.Join(resourceFolder, "tf_"+lang+".txt")); err != nil {
		log.Println(err)
		return
	}

	ig := itemsGame{}
	ig.medals = medals
	dat, err := os.ReadFile(path.Join(itemsFolder, "items_game.txt"))
	if err != nil {
		log.Println(err)
		return
	}

	dat2 := []byte{}
	if staticFile != "" {
		dat2, err = os.ReadFile(staticFile)
		if err != nil {
			log.Println(err)
			return
		}
	}

	ig.init(dat, dat2)
	j, _ := json.MarshalIndent(&ig, "", "\t")

	var prefix string
	if medals {
		prefix = "medals"
	} else {
		prefix = "items"
	}
	os.WriteFile(path.Join(outputFolder, prefix+"_"+lang+".json"), j, 0666)
}

/*func getMap(i interface{}) itemGameMap {
	switch i.(type) {
	case itemGameMap: return i.(itemGameMap)
	case map[string]interface{}: return itemGameMap((i).(map[string]interface{}))
	default: panic("Unknown type")
	}
}*/

func getStringToken(token string) string {
	s, exist := lg.getToken(token)

	if exist {
		return s
	} else {
		return token
	}
}

func getStringTokenRaw(token string) (string, bool) {
	return lg.getToken(token)
}
