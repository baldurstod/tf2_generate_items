package main

import (
	"encoding/json"
	_ "fmt"
	_ "reflect"
	"strconv"
	"strings"

	"github.com/baldurstod/vdf"
)

var ITEM_FIELDS = [...]string{"image_inventory" /*"item_class",*/, "item_slot", "item_type_name" /*, "item_quality"*/, "holiday_restriction", "anim_slot", "particle_suffix", "extra_wearable"}

type item struct {
	ig                   *itemsGame
	Id                   string
	ImageInventory       string
	ItemClass            string
	ItemName             string
	ItemQuality          string
	ItemSlot             string
	ItemTypeName         string
	ModelPlayer          string
	ModelPlayerPerClass  itemStringMap
	Name                 string
	Prefab               string
	prefabs              []*item
	isPrefabsInitialized bool
	UsedByClasses        map[string]int
	kv                   *vdf.KeyValue
	fakeStyle            bool
	isCanteen            bool
	hasOnlyExtraWearable bool
}

func (item *item) toJSON(styleId string) itemGameMap {
	item.initPrefabs()
	ret := make(itemGameMap)

	isAustralium := false
	styleIdNum, _ := strconv.Atoi(styleId)

	if s, _ := item.getStringAttribute("item_quality"); s == "paintkitweapon" {
		if s, ok := item.getStringSubAttribute("visuals.styles." + styleId + ".name"); ok {
			ret["name"] = getStringToken(s)
		} else {
			if s, ok := item.getStringAttribute("name"); ok {
				ret["name"] = getStringToken(s)
			}
		}
		ret["hide"] = 1
	} else {
		if s, ok := item.getStringAttribute("item_name"); ok {
			s = getStringToken(s)
			if s2, ok := item.getStringSubAttribute("visuals.styles." + styleId + ".name"); ok {
				ret["realname"] = s
				s += " (" + getStringToken(s2) + ")"
			} else {
				if styleId == "1" && !item.fakeStyle {
					ret["realname"] = s
					s += " (" + getStringToken("ItemNameAustralium") + ")"
					isAustralium = true
				}
			}
			ret["name"] = s
		} else {
			if s, ok := item.getStringAttribute("name"); ok {
				ret["name"] = getStringToken(s)
			}
		}
	}

	if s, ok := item.getStringAttribute("name"); ok {
		ret["game_name"] = s
	}

	if s, ok := item.getStringAttribute("item_class"); ok {
		ret["item_class"] = s
	}

	for _, val := range ITEM_FIELDS {
		if s, ok := item.getStringAttribute(val); ok {
			if s != "" { //TODO: remove
				s = strings.ReplaceAll(s, "\\", "/")
				ret[val] = getStringToken(s)
			}
		}
	}

	usedByClasses := make(itemStringMap)
	item.getStringMapAttribute("used_by_classes", &usedByClasses)
	if len(usedByClasses) > 0 {
		usedByClassesLower := make(itemStringMap)
		for key, val := range usedByClasses {
			usedByClassesLower[strings.ToLower(key)] = val
		}
		ret["used_by_classes"] = usedByClassesLower
	}

	// model_player
	if modelPlayer, ok := item.getStringSubAttribute("visuals.styles." + styleId + ".model_player"); ok {
		ret["model_player"] = modelPlayer
	} else {
		if modelPlayer, ok := item.getStringAttribute("model_player"); ok {
			if modelWorld, ok := item.getStringAttribute("model_world"); ok {
				modelPlayer = modelWorld
			}
			ret["model_player"] = modelPlayer
		} else {
			ret["model_player"] = "" //TODO: remove me
		}
	}

	if item.fakeStyle && !item.isCanteen && styleId == "1" {
		ret["model_player"] = "" //TODO: remove me
	}

	// model_player_per_class
	modelPlayerPerClass := make(itemStringMap)
	item.getStringMapSubAttribute([]string{"visuals", "styles", styleId, "model_player_per_class"}, &modelPlayerPerClass)
	if len(modelPlayerPerClass) > 0 {
		ret["model_player_per_class"] = modelPlayerPerClass
	} else {
		item.getStringMapAttribute("model_player_per_class", &modelPlayerPerClass)
		if len(modelPlayerPerClass) > 0 {
			ret["model_player_per_class"] = modelPlayerPerClass
		}
	}

	// model_player_per_class_red
	modelPlayerPerClassRed := make(itemStringMap)
	item.getStringMapSubAttribute([]string{"visuals", "styles", styleId, "model_player_per_class_red"}, &modelPlayerPerClassRed)
	if len(modelPlayerPerClassRed) > 0 {
		ret["model_player_per_class_red"] = modelPlayerPerClassRed
	} else {
		item.getStringMapAttribute("model_player_per_class_red", &modelPlayerPerClassRed)
		if len(modelPlayerPerClassRed) > 0 {
			ret["model_player_per_class_red"] = modelPlayerPerClassRed
		}
	}

	// model_player_per_class_blue
	modelPlayerPerClassBlue := make(itemStringMap)
	item.getStringMapSubAttribute([]string{"visuals", "styles", styleId, "model_player_per_class_blue"}, &modelPlayerPerClassBlue)
	if len(modelPlayerPerClassBlue) > 0 {
		ret["model_player_per_class_blue"] = modelPlayerPerClassBlue
	} else {
		item.getStringMapAttribute("model_player_per_class_blue", &modelPlayerPerClassBlue)
		if len(modelPlayerPerClassBlue) > 0 {
			ret["model_player_per_class_blue"] = modelPlayerPerClassBlue
		}
	}

	// equip_regions
	var equipRegions = make(itemStringMap)
	var equipRegionsBug = make(itemStringMap)
	if sm, ok := item.kv.GetStringMap("equip_regions"); ok {
		for key, val := range *sm {
			if val == "1" {
				equipRegions[key] = "1"
			}
		}
	}

	// equip_region
	// sometimes equip_region is an array. this is a bug. In this case,
	// the game will not recognize any region, allowing to equip the item with anything
	equipRegion := make(itemStringMap)
	item.getStringMapAttribute("equip_region", &equipRegion)
	if len(equipRegion) > 0 {
		for key, val := range equipRegion {
			if val == "1" {
				equipRegionsBug[key] = val
			}
		}
	}

	if s, ok := item.getStringAttribute("equip_region"); ok {
		equipRegions[s] = "1"
	}

	if len(equipRegions) > 0 {
		equip := []string{}
		for key := range equipRegions {
			equip = append(equip, key)
		}
		ret["equip_regions"] = equip
	}

	// we export the key equip_region as an array in another key to keep track of them
	if len(equipRegionsBug) > 0 {
		equip := []string{}
		for key := range equipRegionsBug {
			equip = append(equip, key)
		}
		ret["equip_regions_bug"] = equip
		// we also remove the original equip region, as the bug removes everything, prefab regions included
		delete(ret, "equip_regions")
	}

	// paintable
	if paintable, ok := item.getStringSubAttribute("capabilities.paintable"); ok {
		ret["paintable"] = paintable
	}

	// can_customize_texture
	if canCustomizeTtexture, ok := item.getStringSubAttribute("capabilities.can_customize_texture"); ok {
		ret["can_customize_texture"] = canCustomizeTtexture
	}

	//player_bodygroups
	playerBodygroups := make(itemStringMap)
	item.getStringMapSubAttribute([]string{"visuals", "player_bodygroups"}, &playerBodygroups)
	item.getStringMapSubAttribute([]string{"visuals", "styles", styleId, "additional_hidden_bodygroups"}, &playerBodygroups)
	if len(playerBodygroups) > 0 {
		for _, val := range playerBodygroups {
			if val == "1" {
			}
		}
		ret["player_bodygroups"] = playerBodygroups
	}

	//wm_bodygroup_override
	if wmBodygroupOverride, ok := item.getStringSubAttribute("visuals.wm_bodygroup_override"); ok {
		if wmBodygroupStateOverride, ok := item.getStringSubAttribute("visuals.wm_bodygroup_state_override"); ok {
			ret["wm_bodygroup_override"] = itemStringMap{wmBodygroupOverride: wmBodygroupStateOverride}
		}
	}

	// is_taunt_item
	if isTaunt, ok := item.getStringSubAttribute("tags.is_taunt_item"); ok {
		ret["is_taunt_item"] = isTaunt
	}

	// weapon_stattrak_module_scale
	if moduleScale, ok := item.getStringSubAttribute("static_attrs.weapon_stattrak_module_scale"); ok {
		ret["weapon_stattrak_module_scale"] = moduleScale
	}

	// weapon_uses_stattrak_module
	if stattrakModule, ok := item.getStringSubAttribute("static_attrs.weapon_uses_stattrak_module"); ok {
		ret["weapon_uses_stattrak_module"] = stattrakModule
	}

	// custom_taunt_scene_per_class
	customTauntScenePerClass := make(itemGameMap)
	item.getSubAttribute([]string{"taunt", "custom_taunt_scene_per_class"}, &customTauntScenePerClass)
	if len(customTauntScenePerClass) > 0 {
		ret["custom_taunt_scene_per_class"] = customTauntScenePerClass
	}

	// custom_taunt_prop_scene_per_class
	customTauntPropScenePerClass := make(itemStringMap)
	item.getStringMapSubAttribute([]string{"taunt", "custom_taunt_prop_scene_per_class"}, &customTauntPropScenePerClass)
	if len(customTauntPropScenePerClass) > 0 {
		ret["custom_taunt_prop_scene_per_class"] = customTauntPropScenePerClass
	}

	// custom_taunt_outro_scene_per_class
	customTauntOutroScenePerClass := make(itemStringMap)
	item.getStringMapSubAttribute([]string{"taunt", "custom_taunt_outro_scene_per_class"}, &customTauntOutroScenePerClass)
	if len(customTauntOutroScenePerClass) > 0 {
		ret["custom_taunt_outro_scene_per_class"] = customTauntOutroScenePerClass
	}

	// custom_taunt_prop_outro_scene_per_class
	customTauntPropOutroScenePerClass := make(itemStringMap)
	item.getStringMapSubAttribute([]string{"taunt", "custom_taunt_prop_outro_scene_per_class"}, &customTauntPropOutroScenePerClass)
	if len(customTauntPropOutroScenePerClass) > 0 {
		ret["custom_taunt_prop_outro_scene_per_class"] = customTauntPropOutroScenePerClass
	}

	// custom_taunt_prop_per_class
	customTauntPropPerClass := make(itemStringMap)
	item.getStringMapSubAttribute([]string{"taunt", "custom_taunt_prop_per_class"}, &customTauntPropPerClass)
	if len(customTauntPropPerClass) > 0 {
		ret["custom_taunt_prop_per_class"] = customTauntPropPerClass
	}

	// taunt attack name
	if tauntAttackName, ok := item.getStringSubAttribute("attributes.taunt attack name"); ok {
		ret["taunt_attack_name"] = tauntAttackName
	} else {
		if tauntAttackName, ok := item.getStringSubAttribute("attributes.taunt attack name.value"); ok {
			ret["taunt_attack_name"] = tauntAttackName
		} else {
			if tauntAttackName, ok := item.getStringSubAttribute("static_attrs.taunt attack name"); ok {
				ret["taunt_attack_name"] = tauntAttackName
			}
		}
	}

	// taunt force weapon slot
	if tauntForceWeaponSlot, ok := item.getStringSubAttribute("attributes.taunt force weapon slot.value"); ok {
		ret["taunt_force_weapon_slot"] = tauntForceWeaponSlot
	}

	// attached_models_festive
	if !item.hasOnlyExtraWearable || styleId == "0" {
		if attachedModelsFestive, ok := item.getStringSubAttribute("visuals.attached_models_festive.0.model"); ok {
			ret["attached_models_festive"] = attachedModelsFestive
		} else {
			if attachedModelsFestive, ok := item.getStringSubAttribute("visuals_red.attached_models_festive.0.model"); ok {
				ret["attached_models_festive"] = attachedModelsFestive
			}
		}
	}

	// attached_models
	if attachedModels, ok := item.getStringSubAttribute("visuals.attached_models.0.model"); ok {
		ret["attached_models"] = attachedModels
	} else {
		if attachedModels, ok := item.getStringSubAttribute("visuals_red.attached_models.0.model"); ok {
			ret["attached_models"] = attachedModels
		}
	}

	// paintkit_proto_def_index
	if paintkitProtoDefIndex, ok := item.getStringSubAttribute("static_attrs.paintkit_proto_def_index"); ok {
		ret["paintkit_proto_def_index"] = paintkitProtoDefIndex
	}

	// set_item_tint_rgb
	if setItemTintRgb, ok := item.getStringSubAttribute("attributes.set item tint RGB.value"); ok {
		ret["set_item_tint_rgb"] = setItemTintRgb
	}
	if setItemTintRgb2, ok := item.getStringSubAttribute("attributes.set item tint RGB 2.value"); ok {
		ret["set_item_tint_rgb_2"] = setItemTintRgb2
	}

	if collection, grade, ok := item.ig.getItemCollection(item); ok {
		ret["grade"] = getStringToken(grade)
		ret["collection"] = getStringToken(collection)
	}

	// material_override
	if materialOverride, ok := item.getStringSubAttribute("visuals.material_override"); ok {
		ret["material_override"] = materialOverride
	}

	// use_per_class_bodygroups
	if usePerClassBodygroups, ok := item.getStringSubAttribute("visuals.use_per_class_bodygroups"); ok {
		ret["use_per_class_bodygroups"] = usePerClassBodygroups
	}

	// paintkit_base
	if _, ok := item.getStringAttribute("paintkit_base"); ok {
		ret["paintkit_base"] = 1
	}
	if prefab, ok := item.getStringAttribute("prefab"); ok {
		if !isAustralium && strings.Contains(prefab, "paintkit_base") {
			ret["paintkit_base"] = 1
		}
	}

	// set_attached_particle_static
	if setAttachedParticleStatic, ok := item.getStringSubAttribute("attributes.attach particle effect static.value"); ok {
		ret["set_attached_particle_static"] = setAttachedParticleStatic
	}

	// use_smoke_particle_effect
	if useSmokeParticleEffect, ok := item.getStringSubAttribute("visuals.styles." + styleId + ".use_smoke_particle_effect"); ok {
		ret["use_smoke_particle_effect"] = useSmokeParticleEffect
	}

	// taunt_success_sound_loop
	if tauntSuccessSoundLoop, ok := item.getStringSubAttribute("attributes.taunt success sound loop.value"); ok {
		ret["taunt_success_sound_loop"] = tauntSuccessSoundLoop
	}

	// taunt_success_sound_loop_offset
	if tauntSuccessSoundLoopEffect, ok := item.getStringSubAttribute("attributes.taunt success sound loop offset.value"); ok {
		ret["taunt_success_sound_loop_offset"] = tauntSuccessSoundLoopEffect
	}

	// skin_blu, skin_red
	if s, ok := item.getStringSubAttribute("visuals.styles." + styleId + ".skin_red"); ok {
		ret["skin_red"] = s
	} else {
		if s, ok := item.getStringSubAttribute("visuals.styles." + styleId + ".skin"); ok {
			ret["skin_red"] = s
		} else {
			if s, ok := item.getStringAttribute("skin_red"); ok {
				ret["skin_red"] = s
			} else {
				if s, ok := item.getStringSubAttribute("visuals_red.skin"); ok {
					ret["skin_red"] = s
				} else {
					if s, ok := item.getStringAttribute("skin"); ok {
						ret["skin_red"] = s
					} else {
						if s, ok := item.getStringSubAttribute("visuals.skin"); ok {
							ret["skin_red"] = s
						}
					}
				}
			}
		}
	}

	if s, ok := ret["skin_red"]; ok && (s == "0") {
		delete(ret, "skin_red")
	}

	if s, ok := item.getStringSubAttribute("visuals.styles." + styleId + ".skin_blu"); ok {
		ret["skin_blu"] = s
	} else {
		if s, ok := item.getStringSubAttribute("visuals.styles." + styleId + ".skin"); ok {
			ret["skin_blu"] = s
		} else {
			if s, ok := item.getStringAttribute("skin_blu"); ok {
				ret["skin_blu"] = s
			} else {
				if s, ok := item.getStringSubAttribute("visuals_blu.skin"); ok {
					ret["skin_blu"] = s
				} else {
					if s, ok := item.getStringAttribute("skin"); ok {
						ret["skin_blu"] = s
					} else {
						if s, ok := item.getStringSubAttribute("visuals.skin"); ok {
							ret["skin_blu"] = s
						}
					}
				}
			}
		}
	}

	if item.isCanteen {
		ret["skin_red"] = styleIdNum
		ret["skin_blu"] = styleIdNum
		if styleId == "1" {
			ret["name"] = ret["name"].(string) + " (Activated)"
		}
	}

	// attached_particlesystems
	attachedParticlesystems := make(itemGameMap)
	item.getSubAttribute([]string{"visuals", "attached_particlesystems"}, &attachedParticlesystems)
	if len(attachedParticlesystems) > 0 {
		var attached []interface{}
		for _, val := range attachedParticlesystems {
			attached = append(attached, val)
		}
		ret["attached_particlesystems"] = attached
	}

	return ret
}

func (item *item) getStyles() []string {
	styles := []string{}

	if _, ok := item.getStringSubAttribute("static_attrs.paintkit_proto_def_index"); ok {
		// No style when we have a paintkit_proto_def_index
		return styles
	}

	stylesMap := make(itemGameMap)
	item.getSubAttribute([]string{"visuals", "styles"}, &stylesMap)

	for key := range stylesMap {
		styles = append(styles, key)
	}

	if len(styles) == 0 {
		if modelPlayer, ok := item.getStringAttribute("model_player"); ok {
			if extraWearable, ok := item.getStringAttribute("extra_wearable"); ok {
				if modelPlayer != extraWearable {
					if _, ok := item.getStringAttribute("extra_wearable_vm"); !ok {
						if itemClass, ok := item.getStringAttribute("item_class"); ok {
							if (itemClass == "tf_weapon_buff_item") || (itemClass == "tf_weapon_medigun") {
								styles = append(styles, "0", "1")
								item.fakeStyle = true
								item.hasOnlyExtraWearable = true
							}
						}
					}
				}
			} else {
				if itemClass, ok := item.getStringAttribute("item_class"); ok && itemClass == "tf_powerup_bottle" {
					styles = append(styles, "0", "1")
					item.fakeStyle = true
					item.isCanteen = true
				}
			}
		}
	}

	return styles
}

func (item *item) init(ig *itemsGame, kv *vdf.KeyValue) bool {
	item.ig = ig
	item.Id = kv.Key
	item.kv = kv

	return true
}

func (item *item) initPrefabs() {
	if !item.isPrefabsInitialized {
		item.isPrefabsInitialized = true
		if s, ok := item.kv.GetString("prefab"); ok {
			prefabs := strings.Split(s, " ")
			for _, prefabName := range prefabs {
				prefab := item.ig.getPrefab(prefabName)
				prefab.initPrefabs() //Ensure prefab is initialized
				item.prefabs = append(item.prefabs, prefab)
			}
		}
	}
}

func (item *item) getStringMapAttribute(attributeName string, i *itemStringMap) {
	for _, prefab := range item.prefabs {
		prefab.getStringMapAttribute(attributeName, i)
	}

	if sm, ok := item.kv.GetStringMap(attributeName); ok {
		for key, val := range *sm {
			(*i)[key] = val
		}
	}
}

func (item *item) getStringMapSubAttribute(path []string, i *itemStringMap) {
	for _, prefab := range item.prefabs {
		prefab.getStringMapSubAttribute(path, i)
	}

	if sm, ok := item.kv.GetSubElementStringMap(path); ok {
		for key, val := range *sm {
			(*i)[key] = val
		}
	}
}

func (item *item) getSubAttribute(path []string, i *itemGameMap) {
	for _, prefab := range item.prefabs {
		prefab.getSubAttribute(path, i)
	}

	if kv, ok := item.kv.GetSubElement(path); ok {
		for _, val := range kv.GetChilds() {
			(*i)[val.Key] = val
		}
	}
}

func (item *item) getStringAttribute(attributeName string) (string, bool) {
	if s, ok := item.kv.GetString(attributeName); ok {
		return s, true
	}

	for _, prefab := range item.prefabs {
		if s, ok := prefab.getStringAttribute(attributeName); ok && s != "0" { //TODO: remove s != "0"
			return s, true
		}
	}
	return "", false
}

func (item *item) getStringSubAttribute(attributePath string) (string, bool) {
	path := strings.Split(attributePath, ".")

	if kv, ok := item.kv.GetSubElement(path); ok {
		if s, ok := kv.ToString(); ok {
			return s, true
		}
	}

	for _, prefab := range item.prefabs {
		if s, ok := prefab.getStringSubAttribute(attributePath); ok {
			return s, true
		}
	}
	return "", false
}

func (item *item) getUsedByClasses() []string {
	ret := []string{}

	usedByClasses := make(itemStringMap)
	item.getStringMapAttribute("used_by_classes", &usedByClasses)
	if len(usedByClasses) > 0 {
		for key := range usedByClasses {
			ret = append(ret, strings.ToLower(key))
		}
	}
	return ret
}

type itemStyle struct {
	it      *item
	styleId string
}

func (item *itemStyle) MarshalJSON() ([]byte, error) {
	return json.Marshal(item.it.toJSON(item.styleId))
}

/*
func (item itemGameMap) getMapStringValue(key string) (string, bool) {
	if mapValue := item[key]; mapValue != nil {
		switch mapValue.(type) {
		case string:
			return mapValue.(string), true
		default:
			return "", false
		}
	}
	return "", false
}

func (item *itemGameMap) getMapIntValue(key string) (int, bool) {
	if mapValue := (*item)[key]; mapValue != nil {
		if i, err := strconv.Atoi(mapValue.(string)); err == nil {
			return i, true
		}
	}
	return 0, false
}
*/
