package models

// Section defines a single paragraph block from Drupal
type Section struct {
  Type    string                 `json:"type"`
  // Using interface{} allows Go to hold strings (Hero) or arrays (Features)
  Content map[string]interface{} `json:"content"` 
}

// HomepageResponse defines the flattened structure for Astro
type HomepageResponse struct {
  Title    string    `json:"title"`
  Sections []Section `json:"sections"`
}