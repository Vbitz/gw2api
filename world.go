package main

// World represents a Guild Wars 2 world/server
type World struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Population string `json:"population"`
}

// Map represents a Guild Wars 2 map
type Map struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	MinLevel      int     `json:"min_level"`
	MaxLevel      int     `json:"max_level"`
	DefaultFloor  int     `json:"default_floor"`
	Type          string  `json:"type"`
	Floors        []int   `json:"floors"`
	RegionID      int     `json:"region_id"`
	RegionName    string  `json:"region_name"`
	ContinentID   int     `json:"continent_id"`
	ContinentName string  `json:"continent_name"`
	MapRect       [][]int `json:"map_rect"`
	ContinentRect [][]int `json:"continent_rect"`
}

// Continent represents a Guild Wars 2 continent
type Continent struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	ContinentDims []int  `json:"continent_dims"`
	MinZoom       int    `json:"min_zoom"`
	MaxZoom       int    `json:"max_zoom"`
	Floors        []int  `json:"floors"`
}
