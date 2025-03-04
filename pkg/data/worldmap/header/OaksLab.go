package header

import (
	"pokered/pkg/data/tileset"
	"pokered/pkg/data/txt"
	"pokered/pkg/data/worldmap/blk"
)

// OaksLab_h:
// 	db DOJO ; tileset
// 	db OAKS_LAB_HEIGHT, OAKS_LAB_WIDTH ; dimensions (y, x)
// 	dw OaksLab_Blocks ; blocks
// 	dw OaksLab_TextPointers ; texts
// 	dw OaksLab_Script ; scripts
// 	db 0 ; connections
// 	dw OaksLab_Object ; objects

var OaksLab = &Header{
	Tileset: tileset.Gym,
	Height:  6,
	Width:   5,
	blk:     blk.OaksLab[:],
	Text: []string{
		txt.Route1Text1,
		txt.Route1Text2,
		txt.Route1Text3,
	},
}
