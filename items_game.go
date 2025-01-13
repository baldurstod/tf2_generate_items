package main

import (
	"encoding/json"
	"strconv"
	"strings"

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
	medals         bool
	itemsVDF       *vdf.KeyValue
	Prefabs        itemMap
	Items          itemMap
	itemCollection collectionMap
}

func (itemsGame *itemsGame) MarshalJSON() ([]byte, error) {
	ret := make(itemGameMap)

	ret["items"] = *itemsGame.MarshalItems()
	ret["systems"] = *itemsGame.MarshalSystems()

	return json.Marshal(ret)
}

func (itemsGame *itemsGame) MarshalItems() *itemStyleMap {
	items := make(itemStyleMap)
	for itemId, item := range *itemsGame.getItems() {
		styles := item.getStyles()
		//fmt.Println(len(styles))
		if len(styles) > 1 {
			for _, styleId := range styles {
				items[itemId+"~"+styleId] = &itemStyle{it: item, styleId: styleId}
			}
		} else {
			items[itemId] = &itemStyle{it: item, styleId: "0"}
		}
	}

	return &items
}

func (itemsGame *itemsGame) MarshalSystems() *itemGameMap {
	systems := make(itemGameMap)

	if particlesList, ok := itemsGame.itemsVDF.Get("attribute_controlled_attached_particles"); ok {
		for _, particlesGroups := range particlesList.GetChilds() {
			for _, particle := range particlesGroups.GetChilds() {
				particleId := particle.Key
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

func (itemsGame *itemsGame) getItems() *itemMap {
	items := make(itemMap)

	for itemId, item := range itemsGame.Items {
		if ok, _ := itemsGame.filterOut(item, !itemsGame.medals); !ok {
			items[itemId] = item
		}
	}
	return &items
}

func (itemsGame *itemsGame) filterOut(it *item, filterMedals bool) (bool, string) {
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
			if itemId != 26 && itemId != 27 && itemId != 1152 && itemId != 1155 { //destruction PDA, disguise kit, grappling hook, passtime jack
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

func (itemsGame *itemsGame) init(dat []byte, staticDat []byte) {
	v := vdf.VDF{}
	root := v.Parse(dat)
	itemsGame.itemsVDF, _ = root.Get("items_game")
	itemsGame.Prefabs = make(itemMap)
	itemsGame.Items = make(itemMap)
	itemsGame.itemCollection = make(collectionMap)

	if prefabs, ok := itemsGame.itemsVDF.Get("prefabs"); ok {
		for _, val := range prefabs.GetChilds() {
			var it = item{}
			if it.init(itemsGame, val) {
				itemsGame.Prefabs[it.Id] = &it
			}
		}
	}

	if items, ok := itemsGame.itemsVDF.Get("items"); ok {
		for _, val := range items.GetChilds() {
			var it = item{}
			if it.init(itemsGame, val) {
				itemsGame.Items[it.Id] = &it
			}
		}
	}

	// Static file is loaded after to overwrite items_game
	if !itemsGame.medals {
		v := vdf.VDF{}
		staticRoot := v.Parse(staticDat)
		staticItemsVDF, _ := staticRoot.Get("items_game")

		if items, ok := staticItemsVDF.Get("items"); ok {
			for _, val := range items.GetChilds() {
				var it = item{}
				if it.init(itemsGame, val) {
					itemsGame.Items[it.Id] = &it
				}
			}
		}
	}

	if itemCollections, ok := itemsGame.itemsVDF.Get("item_collections"); ok {
		for _, collection := range itemCollections.GetChilds() {

			collectionName, _ := collection.GetString("name")
			if collectionItems, ok := collection.Get("items"); ok {
				for _, grade := range collectionItems.GetChilds() {
					if gradeItems, ok := grade.ToStringMap(); ok {
						for itemName := range *gradeItems {
							itemsGame.itemCollection[itemName] = stringPair{s1: collectionName, s2: grade.Key}
						}
					}
				}
			}
		}
	}
}

func (itemsGame *itemsGame) getPrefab(prefabName string) *item {
	return itemsGame.Prefabs[prefabName]
}

func (itemsGame *itemsGame) getItemCollection(it *item) (string, string, bool) {
	if s, ok := it.getStringAttribute("name"); ok {
		if itemCollection, exists := itemsGame.itemCollection[s]; exists {
			return itemCollection.s1, itemCollection.s2, true
		}
	}
	return "", "", false
}
