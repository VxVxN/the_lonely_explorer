package _map

import (
	"encoding/json"
	"os"
)

// DataMap generated by NotTiled
type DataMap struct {
	Version      float64 `json:"version"`
	Type         string  `json:"type"`
	Infinite     bool    `json:"infinite"`
	Tiledversion string  `json:"tiledversion"`
	Orientation  string  `json:"orientation"`
	Renderorder  string  `json:"renderorder"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	TileWidth    int     `json:"tilewidth"`
	TileHeight   int     `json:"tileheight"`
	Nextlayerid  int     `json:"nextlayerid"`
	Nextobjectid int     `json:"nextobjectid"`
	Properties   []struct {
		Name  string `json:"name"`
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"properties"`
	Tilesets []struct {
		Name        string `json:"name"`
		Firstgid    int    `json:"firstgid"`
		Tilewidth   int    `json:"tilewidth"`
		Tileheight  int    `json:"tileheight"`
		Spacing     int    `json:"spacing"`
		Margin      int    `json:"margin"`
		Columns     int    `json:"columns"`
		Tilecount   int    `json:"tilecount"`
		Image       string `json:"image"`
		Imagewidth  int    `json:"imagewidth"`
		Imageheight int    `json:"imageheight"`
		Tiles       []struct {
			Id         int `json:"id"`
			Properties []struct {
				Name  string `json:"name"`
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"properties"`
		} `json:"tiles"`
		Properties []interface{} `json:"properties"`
	} `json:"tilesets"`
	Layers []struct {
		Type       string        `json:"type"`
		Id         int           `json:"id"`
		Name       string        `json:"name"`
		X          int           `json:"x"`
		Y          int           `json:"y"`
		Width      int           `json:"width"`
		Height     int           `json:"height"`
		Visible    bool          `json:"visible"`
		Opacity    int           `json:"opacity"`
		Offsetx    int           `json:"offsetx"`
		Offsety    int           `json:"offsety"`
		Data       []int         `json:"data"`
		Properties []interface{} `json:"properties"`
	} `json:"layers"`
}
type Map struct {
	Data   *DataMap
	Layers []Layer
}

type Layer [][]int

func NewMap(path string) (*Map, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data DataMap
	jsonParser := json.NewDecoder(file)
	if err = jsonParser.Decode(&data); err != nil {
		return nil, err
	}

	layers := make([]Layer, len(data.Layers))
	for i := range layers {
		layers[i] = make(Layer, data.Layers[i].Width)
		for j := range layers[i] {
			layers[i][j] = make([]int, data.Layers[i].Height)
		}
	}
	for i, layer := range data.Layers {
		for j, datum := range layer.Data {
			x := j % data.Width
			y := j / data.Height
			layers[i][x][y] = datum
		}
	}

	return &Map{Data: &data, Layers: layers}, nil
}
