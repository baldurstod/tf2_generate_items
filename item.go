package main

import (
	"fmt"
	"strconv"
	"strings"
	//"runtime/debug"
	"encoding/json"
)

var ITEM_FIELDS = [...]string{"image_inventory", /*"item_class",*/ "item_slot"/*, /*"item_type_name"/*, "item_quality"*/, "holiday_restriction", "anim_slot", "particle_suffix", "extra_wearable"}

type item struct {
	ig *itemsGame
	Id string
	ImageInventory string
	ItemClass string
	ItemName string
	ItemQuality string
	ItemSlot string
	ItemTypeName string
	ModelPlayer string
	ModelPlayerPerClass itemStringMap
	Name string
	Prefab string
	prefabs []*item
	isPrefabsInitialized bool `default:false`
	UsedByClasses map[string]int
	m itemGameMap
	fakeStyle bool `default:false`
	isCanteen bool `default:false`
	hasOnlyExtraWearable bool `default:false`
}

//func (this *item) MarshalJSON() ([]byte, error) {
func (this *item) toJSON(styleId string) itemGameMap {
	this.initPrefabs()
	ret := make(itemGameMap)

	isAustralium := false
	styleIdNum, _ := strconv.Atoi(styleId)

	if s, _ := this.getStringAttribute("item_quality"); s == "paintkitweapon" {
		if s, ok := this.getStringSubAttribute("visuals.styles." + styleId + ".name"); ok {
			ret["name"] = getStringToken(s)
		} else {
			if s, ok := this.getStringAttribute("name"); ok {
				ret["name"] = getStringToken(s)
			}
		}
		ret["hide"] = 1
	} else {
		if s, ok := this.getStringAttribute("item_name"); ok {
			s = getStringToken(s)
			if s2, ok := this.getStringSubAttribute("visuals.styles." + styleId + ".name"); ok {
				ret["realname"] = s
				s += " (" + getStringToken(s2) + ")"
			} else {
				if styleId == "1" && !this.fakeStyle {
					ret["realname"] = s
					s += " (" + getStringToken("ItemNameAustralium") + ")"
					isAustralium = true
				}
			}
			ret["name"] = s
		} else {
			if s, ok := this.getStringAttribute("name"); ok {
				ret["name"] = getStringToken(s)
			}
		}
	}

	for _, val := range ITEM_FIELDS {
		if s, ok := this.getStringAttribute(val); ok {
			if s != "" { //TODO: remove
				s = strings.ReplaceAll(s, "\\", "/")
				ret[val] = getStringToken(s)
			}
		}
	}

	usedByClasses := make(itemStringMap)
	usedByClassesLower := make(itemStringMap)
	this.getStringMapAttribute("used_by_classes", &usedByClasses)
	if len(usedByClasses) > 0 {
		for key, val := range usedByClasses {
			usedByClassesLower[strings.ToLower(key)] = val
		}
		ret["used_by_classes"] = usedByClassesLower
	}

	// model_player
	if modelPlayer, ok := this.getStringSubAttribute("visuals.styles." + styleId + ".model_player"); ok {
		ret["model_player"] = modelPlayer
	} else {
		if modelPlayer, ok := this.getStringAttribute("model_player"); ok {
			if modelWorld, ok := this.getStringAttribute("model_world"); ok {
				modelPlayer = modelWorld
			}
			ret["model_player"] = modelPlayer
		} else {
			ret["model_player"] = ""//TODO: remove me
		}
	}

	if this.fakeStyle && !this.isCanteen && styleId == "1" {
		ret["model_player"] = ""//TODO: remove me
	}

	// model_player_per_class
	modelPlayerPerClass := make(itemStringMap)
	this.getStringMapSubAttribute("visuals.styles." + styleId + ".model_player_per_class", &modelPlayerPerClass)
	if len(modelPlayerPerClass) > 0 {
		ret["model_player_per_class"] = modelPlayerPerClass
	} else {
		this.getStringMapAttribute("model_player_per_class", &modelPlayerPerClass)
		if len(modelPlayerPerClass) > 0 {
			ret["model_player_per_class"] = modelPlayerPerClass
		}
	}

	// model_player_per_class_red
	modelPlayerPerClassRed := make(itemStringMap)
	this.getStringMapSubAttribute("visuals.styles." + styleId + ".model_player_per_class_red", &modelPlayerPerClassRed)
	if len(modelPlayerPerClassRed) > 0 {
		ret["model_player_per_class_red"] = modelPlayerPerClassRed
	} else {
		this.getStringMapAttribute("model_player_per_class_red", &modelPlayerPerClassRed)
		if len(modelPlayerPerClassRed) > 0 {
			ret["model_player_per_class_red"] = modelPlayerPerClassRed
		}
	}

	// model_player_per_class_blue
	modelPlayerPerClassBlue := make(itemStringMap)
	this.getStringMapSubAttribute("visuals.styles." + styleId + ".model_player_per_class_blue", &modelPlayerPerClassBlue)
	if len(modelPlayerPerClassBlue) > 0 {
		ret["model_player_per_class_blue"] = modelPlayerPerClassBlue
	} else {
		this.getStringMapAttribute("model_player_per_class_blue", &modelPlayerPerClassBlue)
		if len(modelPlayerPerClassBlue) > 0 {
			ret["model_player_per_class_blue"] = modelPlayerPerClassBlue
		}
	}

	// equip_regions
	var equipRegions = make(itemStringMap)
	if v := this.m["equip_regions"]; v != nil {
		mapValue := itemGameMap((v).(map[string]interface{}))
		for key, val := range mapValue {
			if val == "1" {
				equipRegions[key] = "1"
			}
		}
	}

	// equip_region
	// sometimes equip_region is an array
	equipRegion := make(itemStringMap)
	this.getStringMapAttribute("equip_region", &equipRegion)
	if len(equipRegion) > 0 {
		for key, val := range equipRegion {
			if val == "1" {
				equipRegions[key] = val
			}
		}
	}

	if s, ok := this.getStringAttribute("equip_region"); ok {
		equipRegions[s] = "1"
	}

	if len(equipRegions) > 0 {
		equip := []string{}
		for key, _ := range equipRegions {
			equip = append(equip, key)
		}
		ret["equip_regions"] = equip
	}

	// paintable
	if paintable, ok := this.getStringSubAttribute("capabilities.paintable"); ok {
		ret["paintable"] = paintable
	}

	// can_customize_texture
	if canCustomizeTtexture, ok := this.getStringSubAttribute("capabilities.can_customize_texture"); ok {
		ret["can_customize_texture"] = canCustomizeTtexture
	}

	//player_bodygroups
	playerBodygroups := make(itemStringMap)
	this.getStringMapSubAttribute("visuals.player_bodygroups", &playerBodygroups)
	this.getStringMapSubAttribute("visuals.styles." + styleId + ".additional_hidden_bodygroups", &playerBodygroups)
	if len(playerBodygroups) > 0 {
		for _, val := range playerBodygroups {
			if val == "1" {
			}
		}
		ret["player_bodygroups"] = playerBodygroups
	}

	//wm_bodygroup_override
	if wmBodygroupOverride, ok := this.getStringSubAttribute("visuals.wm_bodygroup_override"); ok {
		if wmBodygroupStateOverride, ok := this.getStringSubAttribute("visuals.wm_bodygroup_state_override"); ok {
			ret["wm_bodygroup_override"] = itemStringMap{wmBodygroupOverride: wmBodygroupStateOverride}
		}
	}

	// is_taunt_item
	if isTaunt, ok := this.getStringSubAttribute("tags.is_taunt_item"); ok {
		ret["is_taunt_item"] = isTaunt
	}

	// weapon_stattrak_module_scale
	if moduleScale, ok := this.getStringSubAttribute("static_attrs.weapon_stattrak_module_scale"); ok {
		ret["weapon_stattrak_module_scale"] = moduleScale
	}

	// weapon_uses_stattrak_module
	if stattrakModule, ok := this.getStringSubAttribute("static_attrs.weapon_uses_stattrak_module"); ok {
		ret["weapon_uses_stattrak_module"] = stattrakModule
	}

	// custom_taunt_scene_per_class
	customTauntScenePerClass := make(itemGameMap)
	this.getSubAttribute("taunt.custom_taunt_scene_per_class", &customTauntScenePerClass)
	if len(customTauntScenePerClass) > 0 {
		ret["custom_taunt_scene_per_class"] = customTauntScenePerClass
	}

	// custom_taunt_prop_scene_per_class
	customTauntPropScenePerClass := make(itemStringMap)
	this.getStringMapSubAttribute("taunt.custom_taunt_prop_scene_per_class", &customTauntPropScenePerClass)
	if len(customTauntPropScenePerClass) > 0 {
		ret["custom_taunt_prop_scene_per_class"] = customTauntPropScenePerClass
	}

	// custom_taunt_outro_scene_per_class
	customTauntOutroScenePerClass := make(itemStringMap)
	this.getStringMapSubAttribute("taunt.custom_taunt_outro_scene_per_class", &customTauntOutroScenePerClass)
	if len(customTauntOutroScenePerClass) > 0 {
		ret["custom_taunt_outro_scene_per_class"] = customTauntOutroScenePerClass
	}

	// custom_taunt_prop_outro_scene_per_class
	customTauntPropOutroScenePerClass := make(itemStringMap)
	this.getStringMapSubAttribute("taunt.custom_taunt_prop_outro_scene_per_class", &customTauntPropOutroScenePerClass)
	if len(customTauntPropOutroScenePerClass) > 0 {
		ret["custom_taunt_prop_outro_scene_per_class"] = customTauntPropOutroScenePerClass
	}

	// custom_taunt_prop_per_class
	customTauntPropPerClass := make(itemStringMap)
	this.getStringMapSubAttribute("taunt.custom_taunt_prop_per_class", &customTauntPropPerClass)
	if len(customTauntPropPerClass) > 0 {
		ret["custom_taunt_prop_per_class"] = customTauntPropPerClass
	}

	// taunt attack name
	if tauntAttackName, ok := this.getStringSubAttribute("attributes.taunt attack name"); ok {
		ret["taunt_attack_name"] = tauntAttackName
	} else {
		if tauntAttackName, ok := this.getStringSubAttribute("attributes.taunt attack name.value"); ok {
			ret["taunt_attack_name"] = tauntAttackName
		} else {
			if tauntAttackName, ok := this.getStringSubAttribute("static_attrs.taunt attack name"); ok {
				ret["taunt_attack_name"] = tauntAttackName
			}
		}
	}

	// attached_models_festive
	if !this.hasOnlyExtraWearable || styleId == "0" {
		if attachedModelsFestive, ok := this.getStringSubAttribute("visuals.attached_models_festive.0.model"); ok {
			ret["attached_models_festive"] = attachedModelsFestive
		} else {
			if attachedModelsFestive, ok := this.getStringSubAttribute("visuals_red.attached_models_festive.0.model"); ok {
				ret["attached_models_festive"] = attachedModelsFestive
			}
		}
	}

	// attached_models
	if attachedModels, ok := this.getStringSubAttribute("visuals.attached_models.0.model"); ok {
		ret["attached_models"] = attachedModels
	} else {
		if attachedModels, ok := this.getStringSubAttribute("visuals_red.attached_models.0.model"); ok {
			ret["attached_models"] = attachedModels
		}
	}

	// paintkit_proto_def_index
	if paintkitProtoDefIndex, ok := this.getStringSubAttribute("static_attrs.paintkit_proto_def_index"); ok {
		ret["paintkit_proto_def_index"] = paintkitProtoDefIndex
	}

	// set_item_tint_rgb
	if setItemTintRgb, ok := this.getStringSubAttribute("attributes.set item tint RGB.value"); ok {
		ret["set_item_tint_rgb"] = setItemTintRgb
	}
	if setItemTintRgb2, ok := this.getStringSubAttribute("attributes.set item tint RGB 2.value"); ok {
		ret["set_item_tint_rgb_2"] = setItemTintRgb2
	}

	if collection, grade, ok := this.ig.getItemCollection(this); ok {
		ret["grade"] = getStringToken(grade)
		ret["collection"] = getStringToken(collection)
	}

	// material_override
	if materialOverride, ok := this.getStringSubAttribute("visuals.material_override"); ok {
		ret["material_override"] = materialOverride
	}

	// use_per_class_bodygroups
	if usePerClassBodygroups, ok := this.getStringSubAttribute("visuals.use_per_class_bodygroups"); ok {
		ret["use_per_class_bodygroups"] = usePerClassBodygroups
	}

	// paintkit_base
	if _, ok := this.getStringAttribute("paintkit_base"); ok {
		ret["paintkit_base"] = 1
	}
	if prefab, ok := this.getStringAttribute("prefab"); ok {
		if !isAustralium && strings.Contains(prefab, "paintkit_base") {
			ret["paintkit_base"] = 1
		}
	}

	// set_attached_particle_static
	if setAttachedParticleStatic, ok := this.getStringSubAttribute("attributes.attach particle effect static.value"); ok {
		ret["set_attached_particle_static"] = setAttachedParticleStatic
	}

	// use_smoke_particle_effect
	if useSmokeParticleEffect, ok := this.getStringSubAttribute("visuals.styles." + styleId + ".use_smoke_particle_effect"); ok {
		ret["use_smoke_particle_effect"] = useSmokeParticleEffect
	}

	// taunt_success_sound_loop
	if tauntSuccessSoundLoop, ok := this.getStringSubAttribute("attributes.taunt success sound loop.value"); ok {
		ret["taunt_success_sound_loop"] = tauntSuccessSoundLoop
	}

	// taunt_success_sound_loop_offset
	if tauntSuccessSoundLoopEffect, ok := this.getStringSubAttribute("attributes.taunt success sound loop offset.value"); ok {
		ret["taunt_success_sound_loop_offset"] = tauntSuccessSoundLoopEffect
	}

	// skin_blu, skin_red
	if s, ok := this.getStringSubAttribute("visuals.styles." + styleId + ".skin_red"); ok {
		ret["skin_red"] = s
	} else {
		if s, ok := this.getStringSubAttribute("visuals.styles." + styleId + ".skin"); ok {
			ret["skin_red"] = s
		} else {
			if s, ok := this.getStringAttribute("skin_red"); ok {
				ret["skin_red"] = s
			} else {
				if s, ok := this.getStringSubAttribute("visuals_red.skin"); ok {
					ret["skin_red"] = s
				} else {
					if s, ok := this.getStringAttribute("skin"); ok {
						ret["skin_red"] = s
					} else {
						if s, ok := this.getStringSubAttribute("visuals.skin"); ok {
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

	if s, ok := this.getStringSubAttribute("visuals.styles." + styleId + ".skin_blu"); ok {
		ret["skin_blu"] = s
	} else {
		if s, ok := this.getStringSubAttribute("visuals.styles." + styleId + ".skin"); ok {
			ret["skin_blu"] = s
		} else {
			if s, ok := this.getStringAttribute("skin_blu"); ok {
				ret["skin_blu"] = s
			} else {
				if s, ok := this.getStringSubAttribute("visuals_blu.skin"); ok {
					ret["skin_blu"] = s
				} else {
					if s, ok := this.getStringAttribute("skin"); ok {
						ret["skin_blu"] = s
					} else {
						if s, ok := this.getStringSubAttribute("visuals.skin"); ok {
							ret["skin_blu"] = s
						}
					}
				}
			}
		}
	}

	if this.isCanteen {
		ret["skin_red"] = styleIdNum
		ret["skin_blu"] = styleIdNum
		if styleId == "1" {
			ret["name"] = ret["name"].(string) + " (Activated)"
		}
	}

	// attached_particlesystems
	attachedParticlesystems := make(itemGameMap)
	this.getSubAttribute("visuals.attached_particlesystems", &attachedParticlesystems);
	if len(attachedParticlesystems) > 0 {
		var attached []interface{}
		for _, val := range attachedParticlesystems {
			attached = append(attached, val)
		}
		ret["attached_particlesystems"] = attached
	}


	return ret
}

func (this *item) getStyles() []string {
	styles := []string{}

	if _, ok := this.getStringSubAttribute("static_attrs.paintkit_proto_def_index"); ok {
		// No style when we have a paintkit_proto_def_index
		return styles
	}

	stylesMap := make(itemGameMap)
	this.getSubAttribute("visuals.styles", &stylesMap);

	for key, _ := range stylesMap {
		styles = append(styles, key)
	}

	if len(styles) == 0 {
		if modelPlayer, ok := this.getStringAttribute("model_player"); ok {
			if extraWearable, ok := this.getStringAttribute("extra_wearable"); ok {
				if modelPlayer != extraWearable {
					if _, ok := this.getStringAttribute("extra_wearable_vm"); !ok {
						if itemClass, ok := this.getStringAttribute("item_class"); ok {
							if (itemClass == "tf_weapon_buff_item") || (itemClass == "tf_weapon_medigun") {
								styles = append(styles, "0", "1")
								this.fakeStyle = true
								this.hasOnlyExtraWearable = true
								fmt.Println(this)
							}
						}
					}
				}
			} else {
				if itemClass, ok := this.getStringAttribute("item_class"); ok && itemClass == "tf_powerup_bottle" {
					styles = append(styles, "0", "1")
					this.fakeStyle = true
					this.isCanteen = true
				}
			}
		}
	}

	return styles
}

func (this *item) init(ig *itemsGame, key string, val interface{}) bool {
	this.ig = ig
	this.Id = key

	m := getMap(val)
	this.m = m
	return true
}

func (this *item) initPrefabs() {
	if !this.isPrefabsInitialized {
		this.isPrefabsInitialized = true
		if s, ok := this.m.getMapStringValue("prefab"); ok {
			prefabs := strings.Split(s, " ")
			for _, prefabName := range prefabs {
				prefab := this.ig.getPrefab(prefabName)
				prefab.initPrefabs() //Ensure prefab is initialized
				this.prefabs = append(this.prefabs, prefab)
			}
		}
	}
}

func (this *item) getStringMapAttribute(attributeName string, i *itemStringMap) {
	if v := this.m[attributeName]; v != nil {
		switch v.(type) {
		case map[string]interface{}:
			mapValue := itemGameMap((v).(map[string]interface{}))
			for key, val := range mapValue {
				switch val.(type) {
				case string:
					(*i)[key] = val.(string)
				}
			}
		}
	}

	for _, prefab := range this.prefabs {
		prefab.getStringMapAttribute(attributeName, i)
	}
}

func (this *item) getStringMapSubAttribute(attributePath string, i *itemStringMap) {
	path := strings.Split(attributePath, ".")

	current := this.m

ForLoop:
	for _, p := range path {
		//fmt.Println(p)
		next := current[p]
		switch next.(type) {
			case nil:
				current = nil
				break ForLoop
			case map[string]interface{}:
				current = itemGameMap((next).(map[string]interface{}))
			case string:
				panic("Found a string " + attributePath + this.Id)
			default:
				fmt.Println(next)
				panic("Unknown type")
		}
		current = itemGameMap((next).(map[string]interface{}))
	}

	for _, prefab := range this.prefabs {
		prefab.getStringMapSubAttribute(attributePath, i)
	}

	if current != nil {
		for key, val := range current {
			switch val.(type) {
			case string:
				(*i)[key] = val.(string)
			}
		}
	}

	return
}

func (this *item) getSubAttribute(attributePath string, i *itemGameMap) {
	path := strings.Split(attributePath, ".")

	current := this.m

ForLoop:
	for _, p := range path {
		next := current[p]
		switch next.(type) {
			case nil:
				current = nil
				break ForLoop
			case map[string]interface{}:
				current = itemGameMap((next).(map[string]interface{}))
			case string:
				//return next.(string), true
				panic("Found a string " + attributePath + this.Id)
			default:
				fmt.Println(next)
				panic("Unknown type")
		}
		current = itemGameMap((next).(map[string]interface{}))
	}
	for _, prefab := range this.prefabs {
		prefab.getSubAttribute(attributePath, i)
	}

	if current != nil {
		for key, val := range current {
			switch val.(type) {
			case map[string]interface{}: (*i)[key] = itemGameMap((val).(map[string]interface{}))
			case string: (*i)[key] = val.(string)
			}
		}
	}

	return
}


func (this *item) getStringAttribute(attributeName string) (string, bool) {
	if s, ok := this.m.getMapStringValue(attributeName); ok {
		return s, true
	}

	for _, prefab := range this.prefabs {
		if s, ok := prefab.getStringAttribute(attributeName); ok && s != "0" {//TODO: remove s != "0"
			return s, true
		}
	}
	return "", false
}

func (this *item) getStringSubAttribute(attributePath string) (string, bool) {
	path := strings.Split(attributePath, ".")

	current := this.m
ForLoop:
	for _, p := range path {
		//fmt.Println(p)
		next := current[p]
		switch next.(type) {
			case nil:
				break ForLoop
			case map[string]interface{}:
				current = itemGameMap((next).(map[string]interface{}))
			case string:
				return next.(string), true
			default:
				fmt.Println(next)
				panic("Unknown type")
		}
		current = itemGameMap((next).(map[string]interface{}))
	}

	for _, prefab := range this.prefabs {
		if s, ok := prefab.getStringSubAttribute(attributePath); ok {
			return s, true
		}
	}

	return "", false
}

type itemStyle struct {
	it *item
	styleId string
}

func (this *itemStyle) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.it.toJSON(this.styleId))
}

func (this itemGameMap) getMapStringValue(key string) (string, bool) {
	if mapValue := this[key]; mapValue != nil {
		switch mapValue.(type) {
			case string:
				return mapValue.(string), true
			default:
				return "", false
				//fmt.Println(this)
				//panic("Wrong type " + key )
		}
	}
	return "", false
}

func (this *itemGameMap) getMapIntValue(key string) (int, bool) {
	if mapValue := (*this)[key]; mapValue != nil {
		if i, err := strconv.Atoi(mapValue.(string)); err == nil {
			return i, true
		}
	}
	return 0, false
}
