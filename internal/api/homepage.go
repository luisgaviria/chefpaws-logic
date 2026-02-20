package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/luisgaviria/chefpaws-logic/internal/models"
)

// FetchHomepageData grabs the landing page and flattens all nested paragraphs
func FetchHomepageData(baseURL string) (models.HomepageResponse, error) {
  // 1. Updated URL to include field_media for the Hero
  url := baseURL + "/jsonapi/node/landing_page?include=field_sections,field_sections.field_feature_items,field_sections.field_media"
  resp, err := http.Get(url)
  if err != nil {
    return models.HomepageResponse{}, err
  }
  defer resp.Body.Close()

  var raw map[string]interface{}
  if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
    return models.HomepageResponse{}, err
  }

  // 2. Map 'included' data for quick lookup
  includedMap := make(map[string]map[string]interface{})
  if included, ok := raw["included"].([]interface{}); ok {
    for _, item := range included {
      data := item.(map[string]interface{})
      includedMap[data["id"].(string)] = data
    }
  }

  // 3. Process the main node
  nodes, ok := raw["data"].([]interface{})
  if !ok || len(nodes) == 0 {
    return models.HomepageResponse{}, fmt.Errorf("no landing page found")
  }

  firstNode := nodes[0].(map[string]interface{})
  nodeAttrs := firstNode["attributes"].(map[string]interface{})
  nodeRels := firstNode["relationships"].(map[string]interface{})
  sectionLinks := nodeRels["field_sections"].(map[string]interface{})["data"].([]interface{})

  var finalSections []models.Section

  for _, link := range sectionLinks {
    id := link.(map[string]interface{})["id"].(string)

    if details, found := includedMap[id]; found {
      attrs := details["attributes"].(map[string]interface{})
      sectionType := details["type"].(string)

      // --- HERO IMAGE RESOLUTION ---
      if sectionType == "paragraph--hero" {
        if rels, ok := details["relationships"].(map[string]interface{}); ok {
          if mediaRel, ok := rels["field_media"].(map[string]interface{}); ok {
            if mediaData, ok := mediaRel["data"].(map[string]interface{}); ok {
              mediaID := mediaData["id"].(string)
              // Cross-reference with the file entity in the included map
              if fileEntity, found := includedMap[mediaID]; found {
                fileAttrs := fileEntity["attributes"].(map[string]interface{})
                if uri, ok := fileAttrs["uri"].(map[string]interface{}); ok {
                  attrs["field_media"] = baseURL + uri["url"].(string)
                }
              }
            }
          }
        }
      }

      // --- NESTED FEATURE ITEMS ---
      if sectionType == "paragraph--features_grid" {
        itemsRel := details["relationships"].(map[string]interface{})["field_feature_items"].(map[string]interface{})
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