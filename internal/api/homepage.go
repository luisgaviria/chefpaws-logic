package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/luisgaviria/chefpaws-logic/internal/models"
)

// FetchPageData grabs a specific landing page and flattens all nested paragraphs.
// Optimized for Drupal 11 on Heroku/Railway and local DDEV.
func FetchPageData(baseURL string, slug string) (models.HomepageResponse, error) {
	// 1. CORRECTED INCLUDE PARAMS: 
	// Removed field_sections.field_cta_button because it is a Link field, not a relationship.
	includeParams := "field_sections,field_sections.field_feature_items,field_sections.field_media,field_sections.field_trust_items,field_sections.field_trust_items.field_icon,field_sections.field_teardown_image,field_sections.field_hotspot,field_sections.field_comp_items"

	var url string
	
	// 2. DYNAMIC ROUTING LOGIC (No Hard-coded IDs)
	if slug == "/" || slug == "" || slug == "home" {
		homepageUUID := os.Getenv("HOMEPAGE_UUID")
		homepageNID := os.Getenv("HOMEPAGE_NID")

		if homepageUUID != "" {
			// Best for Production: Fetch by Direct UUID
			url = fmt.Sprintf("%s/jsonapi/node/landing_page/%s?include=%s", baseURL, homepageUUID, includeParams)
		} else if homepageNID != "" {
			// Best for Local: Fetch by Internal NID
			url = fmt.Sprintf("%s/jsonapi/node/landing_page?filter[drupal_internal__nid]=%s&include=%s", baseURL, homepageNID, includeParams)
		} else {
			// Safety Fallback: Use slug filter if no IDs are provided in env
			url = fmt.Sprintf("%s/jsonapi/node/landing_page?filter[path.alias]=/&include=%s", baseURL, includeParams)
		}
	} else {
		// Standard filter for all other library pages
		url = fmt.Sprintf("%s/jsonapi/node/landing_page?filter[path.alias]=%s&include=%s", baseURL, slug, includeParams)
	}

	// DEBUG LOG: Monitor the exact request in your terminal
	fmt.Printf("🔍 DEBUG: Fetching from Drupal: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return models.HomepageResponse{}, err
	}
	defer resp.Body.Close()

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return models.HomepageResponse{}, err
	}

	// 3. DATA EXTRACTION
	var firstNode map[string]interface{}
	data := raw["data"]

	if nodeMap, ok := data.(map[string]interface{}); ok {
		firstNode = nodeMap
	} else if nodeList, ok := data.([]interface{}); ok && len(nodeList) > 0 {
		firstNode = nodeList[0].(map[string]interface{})
	} else {
		// Return fallback if Drupal returns no data (e.g., node is unpublished)
		return models.HomepageResponse{Title: "ChefPaws", Sections: []models.Section{}}, nil
	}

	// 4. LOOKUP MAP (Flattened 'included' data for fast access)
	includedMap := make(map[string]map[string]interface{})
	if included, ok := raw["included"].([]interface{}); ok {
		for _, item := range included {
			itemData := item.(map[string]interface{})
			includedMap[itemData["id"].(string)] = itemData
		}
	}

	nodeAttrs := firstNode["attributes"].(map[string]interface{})
	nodeRels := firstNode["relationships"].(map[string]interface{})

	sectionLinks, ok := nodeRels["field_sections"].(map[string]interface{})["data"].([]interface{})
	if !ok {
		return models.HomepageResponse{Title: fmt.Sprint(nodeAttrs["title"])}, nil
	}

	// 5. COMPONENT MAPPING
	var finalSections []models.Section
	for _, link := range sectionLinks {
		id := link.(map[string]interface{})["id"].(string)

		if details, found := includedMap[id]; found {
			attrs := details["attributes"].(map[string]interface{})
			sectionType := details["type"].(string)
			rels, hasRels := details["relationships"].(map[string]interface{})

			// HERO
			if sectionType == "paragraph--hero" && hasRels {
				if mediaRel, ok := rels["field_media"].(map[string]interface{}); ok && mediaRel["data"] != nil {
					mediaData := mediaRel["data"].(map[string]interface{})
					if fileEntity, found := includedMap[mediaData["id"].(string)]; found {
						if uri, ok := fileEntity["attributes"].(map[string]interface{})["uri"].(map[string]interface{}); ok {
							attrs["field_media"] = resolveS3URL(baseURL, uri["url"].(string))
						}
					}
				}
			}

			// FEATURES GRID
			if sectionType == "paragraph--features_grid" && hasRels {
				if itemsRel, ok := rels["field_feature_items"].(map[string]interface{}); ok && itemsRel["data"] != nil {
					itemLinks := itemsRel["data"].([]interface{})
					var nestedFeatures []map[string]interface{}
					for _, itemLink := range itemLinks {
						childID := itemLink.(map[string]interface{})["id"].(string)
						if child, ok := includedMap[childID]; ok {
							nestedFeatures = append(nestedFeatures, child["attributes"].(map[string]interface{}))
						}
					}
					attrs["features"] = nestedFeatures
				}
			}

			// TRUST BAR
			if sectionType == "paragraph--trust_bar" && hasRels {
				if itemsRel, ok := rels["field_trust_items"].(map[string]interface{}); ok && itemsRel["data"] != nil {
					itemLinks := itemsRel["data"].([]interface{})
					var nestedTrustItems []map[string]interface{}
					for _, itemLink := range itemLinks {
						childID := itemLink.(map[string]interface{})["id"].(string)
						if child, ok := includedMap[childID]; ok {
							childAttrs := child["attributes"].(map[string]interface{})
							if childRels, ok := child["relationships"].(map[string]interface{}); ok {
								if iconRel, ok := childRels["field_icon"].(map[string]interface{}); ok && iconRel["data"] != nil {
									iconData := iconRel["data"].(map[string]interface{})
									if iconEntity, found := includedMap[iconData["id"].(string)]; found {
										if uri, ok := iconEntity["attributes"].(map[string]interface{})["uri"].(map[string]interface{}); ok {
											childAttrs["field_icon"] = resolveS3URL(baseURL, uri["url"].(string))
										}
									}
								}
							}
							nestedTrustItems = append(nestedTrustItems, childAttrs)
						}
					}
					attrs["field_trust_items"] = nestedTrustItems
				}
			}

			// VISUAL TEARDOWN
			if sectionType == "paragraph--visual_teardown" && hasRels {
				if imgRel, ok := rels["field_teardown_image"].(map[string]interface{}); ok && imgRel["data"] != nil {
					imgData := imgRel["data"].(map[string]interface{})
					if fileEntity, found := includedMap[imgData["id"].(string)]; found {
						if uri, ok := fileEntity["attributes"].(map[string]interface{})["uri"].(map[string]interface{}); ok {
							attrs["field_teardown_image"] = resolveS3URL(baseURL, uri["url"].(string))
						}
					}
				}
				if hotspotRel, ok := rels["field_hotspot"].(map[string]interface{}); ok && hotspotRel["data"] != nil {
					hotspotLinks := hotspotRel["data"].([]interface{})
					var nestedHotspots []map[string]interface{}
					for _, hLink := range hotspotLinks {
						hID := hLink.(map[string]interface{})["id"].(string)
						if hDetail, ok := includedMap[hID]; ok {
							nestedHotspots = append(nestedHotspots, hDetail["attributes"].(map[string]interface{}))
						}
					}
					attrs["field_hotspots"] = nestedHotspots
				}
			}

			// COMPARISON TABLE
			if sectionType == "paragraph--comparison_table" && hasRels {
				if itemsRel, ok := rels["field_comp_items"].(map[string]interface{}); ok && itemsRel["data"] != nil {
					itemLinks := itemsRel["data"].([]interface{})
					var nestedCompRows []map[string]interface{}
					for _, itemLink := range itemLinks {
						childID := itemLink.(map[string]interface{})["id"].(string)
						if child, ok := includedMap[childID]; ok {
							nestedCompRows = append(nestedCompRows, child["attributes"].(map[string]interface{}))
						}
					}
					attrs["field_comparison_items"] = nestedCompRows
				}
			}

			// FINAL CTA (Maps Title and URL from the Link field)
			if sectionType == "paragraph--cta" {
				if btnField, ok := attrs["field_cta_button"].(map[string]interface{}); ok {
					attrs["button_text"] = btnField["title"]
					uri := btnField["uri"].(string)
					attrs["button_url"] = strings.Replace(uri, "internal:", "", 1)
				}
			}

			finalSections = append(finalSections, models.Section{
				Type:    sectionType,
				Content: attrs,
			})
		}
	}

	return models.HomepageResponse{
		Title:    fmt.Sprint(nodeAttrs["title"]),
		Sections: finalSections,
	}, nil
}

func resolveS3URL(baseURL string, rawURL string) string {
	var finalURL string
	if len(rawURL) > 4 && rawURL[:4] == "http" {
		finalURL = rawURL
	} else {
		finalURL = baseURL + rawURL
	}
	if os.Getenv("IS_DDEV_PROJECT") == "true" {
		if strings.Contains(finalURL, ".r2.dev/") && !strings.Contains(finalURL, "/s3fs-public/") {
			finalURL = strings.Replace(finalURL, ".r2.dev/", ".r2.dev/s3fs-public/", 1)
		}
	}
	return finalURL
}