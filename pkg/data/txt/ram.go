package txt

import "pokered/pkg/store"

var RAM = map[string](func() string){
	"PLAYER": func() string { return store.Player.Name },
	"RIVAL":  func() string { return store.Rival.Name },
	"TMName": func() string { return store.TMName },
}
