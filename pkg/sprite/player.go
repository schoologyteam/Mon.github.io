package sprite

import (
	"fmt"
	"image/png"
	"pokered/pkg/audio"
	"pokered/pkg/data/tilecoll"
	"pokered/pkg/data/tileset"
	"pokered/pkg/joypad"
	"pokered/pkg/store"
	"pokered/pkg/util"
	"pokered/pkg/world"

	ebiten "github.com/hajimehoshi/ebiten/v2"
)

var ledgeJumpCounter byte

// InitPlayer initialize player sprite
func InitPlayer(state uint, x, y int) {
	imgs := make([]*ebiten.Image, 10)
	for i := 0; i < 10; i++ {
		name := "red"
		switch state {
		case Cycling:
			name = "red_cycling"
		case Seel:
			name = "seel"
		}

		path := fmt.Sprintf("/%s_%d.png", name, i)
		f, err := store.FS.Open(path)
		if err != nil {
			util.NotFoundFileError(path)
		}
		defer f.Close()

		img, _ := png.Decode(f)
		imgs[i] = ebiten.NewImageFromImage(img)
	}

	s := &store.Sprite{
		ScreenXPixel: 16 * 4,
		ScreenYPixel: 16*4 - 4,
		MapXCoord:    x,
		MapYCoord:    y,
		VRAM: store.SpriteImage{
			Index:  1,
			Images: imgs,
		},
	}
	store.SpriteData[0] = s
}

// ChangePlayerSprite change player sprite image
func ChangePlayerSprite(state uint) {
	p := store.SpriteData[0]
	if store.IsInvalidSprite(0) {
		return
	}

	imgs := make([]*ebiten.Image, 10)
	for i := 0; i < 10; i++ {
		name := "red"
		switch state {
		case Cycling:
			name = "red_cycling"
		case Seel:
			name = "seel"
		}

		path := fmt.Sprintf("/%s_%d.png", name, i)
		f, err := store.FS.Open(path)
		if err != nil {
			util.NotFoundFileError(path)
		}
		defer f.Close()

		img, _ := png.Decode(f)
		imgs[i] = ebiten.NewImageFromImage(img)
	}
	p.VRAM.Images = imgs
}

// UpdatePlayerSprite update sprite direction and anim counter
// if in moving, increment anim counter
// if player is starting moving, change direction and increment anim counter
func UpdatePlayerSprite() {
	p := store.SpriteData[0]
	if store.IsInvalidSprite(0) {
		return
	}

	if p.WalkCounter > 0 {
		p.AnimationFrame++
		if p.AnimationCounter() == 4 {
			p.AnimationFrame = 0
		}
	}
	p.VRAM.Index = int(p.Direction + p.AnimationCounter())
}

// AdvancePlayerSprite advance player's walk by a frame
func AdvancePlayerSprite() {
	p := store.SpriteData[0]
	if store.IsInvalidSprite(0) {
		return
	}
	p.WalkCounter--
	if p.WalkCounter < 0 {
		p.WalkCounter = 0
	}
	if p.WalkCounter == 0 {
		p.RightHand = !p.RightHand
		p.MapXCoord += p.DeltaX
		p.MapYCoord += p.DeltaY
	}

	for i, s := range store.SpriteData {
		if i == 0 {
			continue
		}
		if store.IsInvalidSprite(uint(i)) {
			return
		}
		s.ScreenXPixel -= p.DeltaX
		s.ScreenYPixel -= p.DeltaY
	}
}

// CollisionCheckForPlayer check if collision occurs in player moving ahead
func CollisionCheckForPlayer() bool {
	collision := false
	defer func() {
		if collision {
			audio.PlaySound(audio.SFX_COLLISION)
		}
	}()

	p := store.SpriteData[0]
	if store.IsInvalidSprite(0) {
		return false
	}

	if store.Flag.D736.InLedgeOrFishingAnim {
		return false
	}

	for offset, s := range store.SpriteData {
		if offset == 0 {
			continue
		}
		if store.IsInvalidSprite(uint(offset)) {
			break
		}
		if s.Hidden {
			continue
		}

		switch p.Direction {
		case util.Up:
			if p.MapXCoord == s.MapXCoord && p.MapYCoord-1 == s.MapYCoord {
				collision = true
			}
		case util.Down:
			if p.MapXCoord == s.MapXCoord && p.MapYCoord+1 == s.MapYCoord {
				collision = true
			}
		case util.Left:
			if p.MapXCoord-1 == s.MapXCoord && p.MapYCoord == s.MapYCoord {
				collision = true
			}
		case util.Right:
			if p.MapXCoord+1 == s.MapXCoord && p.MapYCoord == s.MapYCoord {
				collision = true
			}
		}

		if collision {
			return collision
		}
	}

	if HandleLedges() {
		return false
	}

	tilesetID, frontTileID := world.FrontTileID(p.MapXCoord, p.MapYCoord, p.Direction)
	if frontTileID == world.NoPassable || !util.Contains(tilecoll.Get(tilesetID), byte(frontTileID)) {
		if frontTileID != world.Passable {
			collision = true
			return collision
		}
	}

	_, curTileID := world.CurTileID(p.MapXCoord, p.MapYCoord)
	if tilecoll.IsCollisionPair(tilesetID, byte(curTileID), byte(frontTileID), false) {
		collision = true
		return collision
	}
	return collision
}

// HandleLedges 段差飛び降りできるかチェックして、飛び降りれるなら飛び降りる処理(キー入力の強制や飛び降りモーションアニメの再生)を行う
func HandleLedges() bool {
	if store.Flag.D736.InLedgeOrFishingAnim {
		return false
	}

	if !world.IsCurTileset(tileset.Overworld) {
		return false
	}

	result := isJumpingLedge(0)
	if result {
		store.Flag.D736.InLedgeOrFishingAnim = true
		audio.PlaySound(audio.SFX_LEDGE)
	}
	return result
}

func HandleMidJump() {
	var jumpingYScreenPixel = [...]int{
		0x38, 0x36, 0x34, 0x32, 0x31, 0x30, 0x30, 0x30, 0x31, 0x32, 0x33, 0x34, 0x36, 0x38, 0x3C, 0x3C,
	}

	p := store.SpriteData[0]
	if p == nil {
		return
	}

	if ledgeJumpCounter < 32 {
		p.ScreenYPixel = jumpingYScreenPixel[ledgeJumpCounter/2]
		ledgeJumpCounter++
		return
	}

	// finished jump
	UpdateSprites()
	store.DelayFrames = 3

	joypad.JoyHeld, joypad.JoyPressed, joypad.JoyReleased, joypad.JoyIgnore = joypad.Input{}, joypad.Input{}, joypad.Input{}, joypad.Input{}
	ledgeJumpCounter = 0
	p.AnimationFrame = 0
	store.Flag.D736.InLedgeOrFishingAnim = false
}

// IsPlayerFacingEdgeOfMap check player faces edge of the current map
func IsPlayerFacingEdgeOfMap() bool {
	p := store.SpriteData[0]
	if store.IsInvalidSprite(0) {
		return false
	}
	return world.FaceEdgeOfMap(p.MapXCoord, p.MapYCoord, p.Direction)
}

func IsWarpTileInFrontOfPlayer() bool {
	return false
}

func PlayerCurTileID() (uint, int) {
	p := store.SpriteData[0]
	return world.CurTileID(p.MapXCoord, p.MapYCoord)
}
