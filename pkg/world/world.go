package world

import (
	"pokered/pkg/audio"
	"pokered/pkg/data/tileset"
	"pokered/pkg/data/worldmap/header"
	"pokered/pkg/data/worldmap/object"
	"pokered/pkg/data/worldmap/song"
	"pokered/pkg/screen"
	"pokered/pkg/store"
	"pokered/pkg/util"

	ebiten "github.com/hajimehoshi/ebiten/v2"
)

// World data
type World struct {
	MapID  int
	Image  *ebiten.Image
	Header *header.Header
	Object *object.Object
}

// CurWorld current map data
var CurWorld *World

// LastWorld last map data
var LastWorld *World

// WarpTo used in loadMapData
var WarpTo [2]int = [2]int{-1, -1}

// map exterior range(block)
const exterior int = 3

// LoadWorldData load world data
func LoadWorldData(id int) {
	curAnimTiles = []animTile{}

	h, o := header.Get(id), object.Get(id)

	o.Initialized = false
	img := ebiten.NewImage(int(h.Width*32)+2*exterior*32, int(h.Height*32)+2*exterior*32)
	loadBlockset(h.Tileset)

	for y := 0; y < int(h.Height)+2*exterior; y++ {
		for x := 0; x < int(h.Width)+2*exterior; x++ {
			switch {
			case y < int(exterior):
				northCon := h.Connections.North
				if northCon.OK {
					northMapH, northMapO := header.Get(northCon.DestMapID), object.Get(northCon.DestMapID)
					if x < int(exterior) || x > int(h.Width)+exterior-1 {
						DrawImageBlock(img, northMapO.Border, x, y)
						continue
					}
					blockID := northMapH.Blk(int((northMapH.Height-uint(exterior-y))*northMapH.Width) + (x - exterior))
					DrawImageBlock(img, blockID, x, y)
				} else {
					DrawImageBlock(img, o.Border, x, y)
				}

			case y > int(h.Height)+exterior-1:
				southCon := h.Connections.South
				if southCon.OK {
					southMapH := header.Get(southCon.DestMapID)
					if x < int(exterior) || x > int(h.Width)+1 {
						DrawImageBlock(img, o.Border, x, y)
						continue
					}
					blockID := southMapH.Blk(int((uint(y)-h.Height-uint(exterior))*southMapH.Width) + (x - exterior))
					DrawImageBlock(img, blockID, x, y)
				} else {
					DrawImageBlock(img, o.Border, x, y)
				}

			case x < int(exterior) || x > int(h.Width)+2:
				DrawImageBlock(img, o.Border, x, y)

			default:
				blockID := h.Blk((y-exterior)*int(h.Width) + (x - exterior))
				DrawImageBlock(img, blockID, x, y)
			}
		}
	}

	CurWorld = &World{
		MapID:  id,
		Image:  img,
		Header: h,
		Object: o,
	}
	PlayDefaultMusicFadeOutCurrent(id)
}

// PlayDefaultMusicFadeOutCurrent Fade out the current music and then play the default music.
// ref: PlayDefaultMusicFadeOutCurrent
func PlayDefaultMusicFadeOutCurrent(mapID int) {
	musicID := song.MapMusics[mapID]
	if musicID == 0 {
		return
	}
	switch store.Player.State {
	case store.BikeState:
	case store.SurfState:
	}

	audio.NewMusicID = musicID
	if musicID == audio.LastMusicID {
		return
	}

	audio.StopMusic(10)
	audio.NewMusicID = musicID
}

// VBlank script executed in VBlank
func VBlank(XCoord, YCoord, deltaX, deltaY, walkCounter int, direction uint) {
	if CurWorld == nil {
		return
	}
	updateMovingTiles()
	x := -32*exterior - XCoord*16 + 64
	y := -32*exterior - YCoord*16 + 64
	if walkCounter > 0 {
		x -= deltaX * (16 - walkCounter)
		y -= deltaY * (16 - walkCounter)
	}
	screen.AddLayer("overworld", screen.World, CurWorld.Image, x, y)
}

func DrawImageBlock(target *ebiten.Image, blockID byte, x, y int) {
	checkAnimTiles(uint(blockID), x, y)
	block := CurBlockset.Data[blockID]
	util.DrawImageBlock(target, block, x, y)
}

// CheckIfInOutsideMap If the player is in an outside map (a town or route), set the z flag
func CheckIfInOutsideMap() bool {
	tilesetID := CurBlockset.TilesetID
	return tilesetID == tileset.Overworld || tilesetID == tileset.Plateau
}
