package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/luisgaviria/chefpaws-logic/internal/models"
)

// FetchPageData grabs a specific landing page and flattens all nested paragraphs.
// Uses Decoupled Router to resolve slugs to UUIDs for reliable fetching.
func FetchPageData(baseURL string, slug string) (models.HomepageResponse, error) {
	// NOTE: field_sections.field_process_steps is intentionally omitted here.
	// Drupal's JSON:API schema validates include paths against ALL paragraph types
	// that field_sections can hold — not just the ones on this page. Since only
	// paragraph--process_steps has field_process_steps, including it globally
	// causes a 400. Child steps are fetched via a targeted secondary request instead.
	includeParams := "field_sections,field_sections.field_feature_items,field_sections.field_media,field_sections.field_trust_items,field_sections.field_trust_items.field_icon,field_sections.field_teardown_image,field_sections.field_hotspot,field_sections.field_comp_items"

	var url string

	// 1. DYNAMIC ROUTING & RESOLUTION LOGIC
	if slug == "/" || slug == "" || slug == "home" {
		homepageUUID := os.Getenv("HOMEPAGE_UUID")
		homepageNID := os.Getenv("HOMEPAGE_NID")

		if homepageUUID != "" {
			url = fmt.Sprintf("%s/jsonapi/node/landing_page/%s?include=%s", baseURL, homepageUUID, includeParams)
		} else if homepageNID != "" {
			url = fmt.Sprintf("%s/jsonapi/node/landing_page?filter[drupal_internal__nid]=%s&include=%s", baseURL, homepageNID, includeParams)
		} else {
			url = fmt.Sprintf("%s/jsonapi/node/landing_page?filter[path.alias]=/&include=%s", baseURL, includeParams)
		}
	} else {
		// Ensure leading slash for the router
		cleanSlug := slug
		if !strings.HasPrefix(cleanSlug, "/") {
			cleanSlug = "/" + cleanSlug
		}

		// Use Decoupled Router to find the UUID of the node at this path
		routerURL := fmt.Sprintf("%s/router/translate-path?path=%s", baseURL, cleanSlug)
		fmt.Printf("🌐 ROUTER: Resolving path: %s\n", routerURL)

		routerResp, err := http.Get(routerURL)
		if err != nil {
			fmt.Printf("❌ ROUTER ERROR: %v\n", err)
			return models.HomepageResponse{Title: "Not Found", Sections: []models.Section{}}, nil
		}
		fmt.Printf("🌐 ROUTER STATUS: %d\n", routerResp.StatusCode)
		if routerResp.StatusCode != 200 {
			body, _ := io.ReadAll(routerResp.Body)
			routerResp.Body.Close()
			fmt.Printf("❌ ROUTER NON-200 BODY: %s\n", string(body))
			return models.HomepageResponse{Title: "Not Found", Sections: []models.Section{}}, nil
		}
		defer routerResp.Body.Close()

		var routerResult struct {
			Entity struct {
				UUID string `json:"uuid"`
			} `json:"entity"`
		}
		if err := json.NewDecoder(routerResp.Body).Decode(&routerResult); err != nil || routerResult.Entity.UUID == "" {
			fmt.Printf("❌ ROUTER: failed to decode or UUID empty (err=%v, uuid=%q)\n", err, routerResult.Entity.UUID)
			return models.HomepageResponse{Title: "Not Found", Sections: []models.Section{}}, nil
		}
		fmt.Printf("✅ ROUTER: resolved UUID = %s\n", routerResult.Entity.UUID)

		url = fmt.Sprintf("%s/jsonapi/node/landing_page/%s?include=%s", baseURL, routerResult.Entity.UUID, includeParams)
	}

	fmt.Printf("🔍 Fetching from Drupal: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return models.HomepageResponse{}, err
	}
	defer resp.Body.Close()

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return models.HomepageResponse{}, err
	}

	// 2. DATA EXTRACTION
	var firstNode map[string]interface{}
	data := raw["data"]

	if nodeMap, ok := data.(map[string]interface{}); ok {
		firstNode = nodeMap
	} else if nodeList, ok := data.([]interface{}); ok && len(nodeList) > 0 {
		firstNode = nodeList[0].(map[string]interface{})
	} else {
		return models.HomepageResponse{Title: "ChefPaws", Sections: []models.Section{}}, nil
	}

	// 3. LOOKUP MAP (Flattened 'included' data)
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
		fmt.Printf("❌ field_sections not found or empty in relationships\n")
		return models.HomepageResponse{Title: fmt.Sprint(nodeAttrs["title"])}, nil
	}

	// 4. COMPONENT MAPPING
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

			// PROCESS STEPS
			// Cannot be included via the global includeParams (Drupal validates field_process_steps
			// against ALL paragraph types in field_sections and rejects it with a 400 because only
			// paragraph--process_steps has that field). Instead, fetch the paragraph directly.
			if sectionType == "paragraph--process_steps" {
				stepsURL := fmt.Sprintf("%s/jsonapi/paragraph/process_steps/%s?include=field_process_steps", baseURL, id)
				stepsResp, err := http.Get(stepsURL)
				if err == nil && stepsResp.StatusCode == 200 {
					var stepsRaw map[string]interface{}
					if json.NewDecoder(stepsResp.Body).Decode(&stepsRaw) == nil {
						// Build lookup from secondary included
						stepsIncluded := make(map[string]map[string]interface{})
						if inc, ok := stepsRaw["included"].([]interface{}); ok {
							for _, item := range inc {
								d := item.(map[string]interface{})
								stepsIncluded[d["id"].(string)] = d
							}
						}
						// Walk relationship links and flatten step attributes
						if stepData, ok := stepsRaw["data"].(map[string]interface{}); ok {
							if stepRels, ok := stepData["relationships"].(map[string]interface{}); ok {
								if ref, ok := stepRels["field_process_steps"].(map[string]interface{}); ok && ref["data"] != nil {
									if links, ok := ref["data"].([]interface{}); ok {
										var steps []map[string]interface{}
										for _, l := range links {
											cid := l.(map[string]interface{})["id"].(string)
											if child, ok := stepsIncluded[cid]; ok {
												ca := child["attributes"].(map[string]interface{})
												steps = append(steps, map[string]interface{}{
													"field_headline":    ca["field_headline"],
													"field_description": ca["field_description"],
												})
											}
										}
										attrs["field_process_steps"] = steps
									}
								}
							}
						}
					}
					stepsResp.Body.Close()
				}
			}

			// FINAL CTA
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
	if !strings.HasPrefix(rawURL, "http") {
		finalURL = baseURL + rawURL
	} else {
		finalURL = rawURL
	}

	if strings.Contains(finalURL, ".r2.dev/") && !strings.Contains(finalURL, "/s3fs-public/") {
		finalURL = strings.Replace(finalURL, ".r2.dev/", ".r2.dev/s3fs-public/", 1)
	}
	return finalURL
}
