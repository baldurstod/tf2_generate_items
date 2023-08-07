package main

import (
	_ "fmt"
	"os"
	"strings"
	"strconv"
	"encoding/json"
	//"github.com/baldurstod/vdf"
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
	itemsVDF *KeyValue
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
		//fmt.Println(len(styles))
		if len(styles) > 1 {
			for _, styleId := range styles {
				items[itemId + "~" + styleId] = &itemStyle{it: item, styleId: styleId}
			}
		} else {
			items[itemId] = &itemStyle{it: item, styleId: "0"}
		}
	}

	/*for itemId, itemData := range this.staticData {
		it := item{}
		if it.init(this, itemId, getMap(itemData)) {
			items[itemId] = &itemStyle{it: &it, styleId: "0"}
		}
	}*/

	return &items
}

func (this *itemsGame) MarshalSystems() *itemGameMap {
	systems := make(itemGameMap)

	if particlesList, ok := this.itemsVDF.Get("attribute_controlled_attached_particles"); ok {
		for _, particlesGroups := range particlesList.GetChilds() {
			for _, particle := range particlesGroups.GetChilds() {
				particleId := particle.key
				particleSystem, _ := particle.ToStringMap()
				systems[particleId] = particleSystem

				if s, ok := getStringTokenRaw("Attrib_Particle" + particleId); ok {
					(*particleSystem)["name"] = s
				} else {
					if s, ok := getStringTokenRaw("Attrib_KillStreakEffect" + particleId); ok {
						(*particleSystem)["name"] = s
					}
				}
			}
		}
	}

	return &systems
}

func (this *itemsGame) getItems() (*itemMap) {
	items := make(itemMap)

	for itemId, item := range this.Items {

		/*if itemId == "462" {
			fmt.Println(item.kv)
			fmt.Println(this.filterOut(item, !this.medals))
		}*/

		if ok, _ := this.filterOut(item, !this.medals); !ok {
			items[itemId] = item
		}
	}
	return &items
}

func (this *itemsGame) filterOut(it *item, filterMedals bool) (bool, string) {
	it.initPrefabs()

	itemId, _ := strconv.Atoi(it.Id)

	if itemId == 5838 { //Winter 2015 Mystery Box
		return true, "item is id 5838"
	}

	// Filter medals
	if s, ok := it.getStringAttribute("item_type_name"); ok {
		if filterMedals {
			if s == "#TF_Wearable_TournamentMedal" {
				return true, "item is tournament medal"
			}
		} else {
			if s != "#TF_Wearable_TournamentMedal" {
				return true, "item is not tournament medal"
			}
		}
	}

	if s, ok := it.getStringAttribute("item_name"); ok {
		if s == "#TF_Item_Zombie_Armory" {
			return true, "item is zombie armory"
		}
	}

	if s, ok := it.getStringAttribute("name"); ok {
		if strings.Contains(s, "Autogrant") {
			return true, "item is autogrant"
		}
	}

	if s, ok := it.getStringAttribute("item_class"); ok {
		if s == "tf_weapon_invis" {
			return true, "item is watch"
		}
	}

	if s, ok := it.getStringAttribute("baseitem"); ok {
		if s == "1" {
			if itemId != 26 && itemId != 27 {//destruction PDA and disguise kit
				return true, "item is base item"
			}
		}
	}

	// Filter show_in_armory
	if s, ok := it.getStringAttribute("show_in_armory"); ok {
		if s == "0" {
			if s, ok := it.getStringAttribute("name"); ok && s != "Duck Badge" {
				if itemId == 294 || (itemId >= 831 && itemId <= 838) || (itemId >= 30739 && itemId <= 30741) {
				} else {
					return true, "item doesn't have show_in_armory flag"
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
						return true, "item doesn't have a model and item_slot is action"
					}
				} else {
					return true, "item doesn't have a model nor an item_slot"
				}
			}
		}
	}

	usedByClasses := it.getUsedByClasses()
	if len(usedByClasses) == 0 {
		return true, "item is used by no one"
	}

	return false, ""
}

func (this *itemsGame) init(dat []byte, staticPath string) {
	this.staticData = make(itemGameMap)
	//dat, _ := os.ReadFile(path)
	if !this.medals {
		staticContent, _ := os.ReadFile(staticPath)
		_ = json.Unmarshal(staticContent, &this.staticData)
	}

	vdf := VDF{}
	root := vdf.Parse(dat)
	this.itemsVDF, _ = root.Get("items_game")
	this.Prefabs = make(itemMap)
	this.Items = make(itemMap)
	this.itemCollection = make(collectionMap)

	if prefabs, ok := this.itemsVDF.Get("prefabs"); ok {
		for _, val := range prefabs.GetChilds() {
			var it = item{}
			if it.init(this, val) {
				this.Prefabs[it.Id] = &it
			}
		}
	}

	if items, ok := this.itemsVDF.Get("items"); ok {
		for _, val := range items.GetChilds() {
			var it = item{}
			if it.init(this, val) {
				this.Items[it.Id] = &it
			}
		}
	}

	if itemCollections, ok := this.itemsVDF.Get("item_collections"); ok {
		for _, collection := range itemCollections.GetChilds() {

			collectionName := collection.key
			if collectionItems, ok := collection.Get("items"); ok {
				for _, grade := range collectionItems.GetChilds() {
					if gradeItems, ok := grade.ToStringMap(); ok {
						for itemName, _ := range *gradeItems {
							this.itemCollection[itemName] = stringPair{s1: collectionName, s2: grade.key}
						}
					}
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
