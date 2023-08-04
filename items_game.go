package main

import (
	"os"
	"strings"
	"strconv"
	"encoding/json"
	"github.com/baldurstod/vdf"
)

type itemMap map[string]*item
type itemStyleMap map[string]*itemStyle
type itemGameMap map[string]interface{}
type itemStringMap map[string]string
type stringPair struct {
	s1 string
	s2 string
}
type collectionMap map[string]stringPair


type itemsGame struct {
	medals bool `default:false`
	itemsVDF itemGameMap
	Prefabs itemMap
	Items itemMap
	itemCollection collectionMap
	staticData itemGameMap
}

func (this *itemsGame) MarshalJSON() ([]byte, error) {
	ret := make(itemGameMap)

	ret["items"] = *this.MarshalItems()
	ret["systems"] = *this.MarshalSystems()

	return json.Marshal(ret)
}

func (this *itemsGame) MarshalItems() *itemStyleMap {
	items := make(itemStyleMap)
	for itemId, item := range *this.getItems() {
		styles := item.getStyles()
		if len(styles) > 1 {
			for _, styleId := range styles {
				items[itemId + "~" + styleId] = &itemStyle{it: item, styleId: styleId}
			}
		} else {
			items[itemId] = &itemStyle{it: item, styleId: "0"}
		}
	}

	for itemId, itemData := range this.staticData {
		it := item{}
		if it.init(this, itemId, getMap(itemData)) {
			items[itemId] = &itemStyle{it: &it, styleId: "0"}
		}
	}

	return &items
}

func (this *itemsGame) MarshalSystems() *itemGameMap {
	systems := make(itemGameMap)

	particles := getMap(getMap(this.itemsVDF)["attribute_controlled_attached_particles"])
	for _, val := range particles {
		subParticles := getMap(val)
		for particleId, val := range subParticles {
			systems[particleId] = getMap(val)

			if s, ok := getStringTokenRaw("Attrib_Particle" + particleId); ok {
				getMap(systems[particleId])["name"] = s
			} else {
				if s, ok := getStringTokenRaw("Attrib_KillStreakEffect" + particleId); ok {
					getMap(systems[particleId])["name"] = s
				}
			}
		}
	}


	return &systems
}

func (this *itemsGame) getItems() (*itemMap) {
	items := make(itemMap)

	for itemId, item := range this.Items {
		if !this.filterOut(item, !this.medals) {
			items[itemId] = item
		}
	}
	return &items
}

func (this *itemsGame) filterOut(it *item, filterMedals bool) (bool) {
	it.initPrefabs()

	itemId, _ := strconv.Atoi(it.Id)

	if itemId == 5838 { //Winter 2015 Mystery Box
		return true
	}

	// Filter medals
	if s, ok := it.getStringAttribute("item_type_name"); ok {
		if filterMedals {
			if s == "#TF_Wearable_TournamentMedal" {
				return true
			}
		} else {
			if s != "#TF_Wearable_TournamentMedal" {
				return true
			}
		}
	}

	if s, ok := it.getStringAttribute("item_name"); ok {
		if s == "#TF_Item_Zombie_Armory" {
			return true
		}
	}

	if s, ok := it.getStringAttribute("name"); ok {
		if strings.Contains(s, "Autogrant") {
			return true
		}
	}

	if s, ok := it.getStringAttribute("item_class"); ok {
		if s == "tf_weapon_invis" {
			return true
		}
	}

	if s, ok := it.getStringAttribute("baseitem"); ok {
		if s == "1" {
			if itemId != 26 && itemId != 27 {//destruction PDA and disguise kit
				return true;
			}
		}
	}

	// Filter show_in_armory
	if s, ok := it.getStringAttribute("show_in_armory"); ok {
		if s == "0" {
			if s, ok := it.getStringAttribute("name"); ok && s != "Duck Badge" {
				if itemId == 294 || (itemId >= 831 && itemId <= 838) || (itemId >= 30739 && itemId <= 30741) {
				} else {
					return true
				}
			}
		}
	}

	// Filter items whitout models excepts a few (hatless...)
	if s, ok := it.getStringAttribute("model_player"); !ok || s == "" {
		if _, ok := it.getStringAttribute("model_player_per_class"); !ok {
			if _, ok := it.getStringAttribute("extra_wearable"); !ok {
				if itemSlot, ok := it.getStringAttribute("item_slot"); ok {
					if strings.Contains(itemSlot, "action") {
						return true
					}
				} else {
					return true
				}
			}
		}
	}

	usedByClasses := make(itemStringMap)
	it.getStringMapAttribute("used_by_classes", &usedByClasses)
	if len(usedByClasses) == 0 {
		return true
	}



	return false
}

func (this *itemsGame) init(path string, staticPath string) {
	this.staticData = make(itemGameMap)
	dat, _ := os.ReadFile(path)
	if !this.medals {
		staticContent, _ := os.ReadFile(staticPath)
		_ = json.Unmarshal(staticContent, &this.staticData)
	}

	vdf := vdf.VDF{}
	itemsVdf := vdf.Parse(dat)
	this.itemsVDF = getMap(itemsVdf["items_game"]);
	this.Prefabs = make(itemMap)
	this.Items = make(itemMap)
	this.itemCollection = make(collectionMap)

	prefabs := getMap(getMap(this.itemsVDF)["prefabs"])
	for key, val := range prefabs {
		var it = item{}
		if it.init(this, key, val) {
			this.Prefabs[it.Id] = &it
		}
	}

	items := getMap(getMap(this.itemsVDF)["items"])
	for key, val := range items {
		var it = item{}
		if it.init(this, key, val) {
			this.Items[it.Id] = &it
		}
	}

	itemCollections := getMap(getMap(this.itemsVDF)["item_collections"])
	for _, val := range itemCollections {
		collection := getMap(val)
		collectionItems := getMap(collection["items"])
		collectionName := collection["name"].(string)

		for gradeName, val := range collectionItems {
			switch val.(type) {
			case map[string]interface{}:
				gradeItems := itemGameMap((val).(map[string]interface{}))

				for itemName, _ := range gradeItems {
					this.itemCollection[itemName] = stringPair{s1: collectionName, s2: gradeName}
				}

			}
		}
	}
}

func (this *itemsGame) getPrefab(prefabName string) *item {
	return this.Prefabs[prefabName]
}

func (this *itemsGame) getItemCollection(it *item) (string, string, bool) {
	if s, ok := it.getStringAttribute("name"); ok {
		if itemCollection, exists := this.itemCollection[s]; exists {
			return itemCollection.s1, itemCollection.s2, true
		}
	}
	return "", "", false
}
