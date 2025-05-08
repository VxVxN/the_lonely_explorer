package eventmanager

import (
	_map "github.com/VxVxN/the_lonely_explorer/internal/map"
	"github.com/VxVxN/the_lonely_explorer/pkg/player"
)

type EventManager struct {
	player  *player.Player
	gameMap *_map.Map

	events []Event
}

func NewEventManager(player *player.Player, gameMap *_map.Map) *EventManager {
	return &EventManager{
		player:  player,
		gameMap: gameMap,
	}
}

func (em *EventManager) SetEvents(events []Event) {
	em.events = events
}

func (em *EventManager) Update() {
	for _, event := range em.events {
		if event.Done() {
			continue
		}
		if event.Check(em.player, em.gameMap) {
			event.Action()
		}
	}
}

type Event interface {
	Check(player *player.Player, gameMap *_map.Map) bool
	Action()
	Done() bool
}

type MeetEvent struct {
	whom []int
	baseEvent
}

func NewMeetEvent(whom []int, action func()) *MeetEvent {
	return &MeetEvent{
		whom: whom,
		baseEvent: baseEvent{
			action: action,
		},
	}
}

func (e *MeetEvent) Check(player *player.Player, gameMap *_map.Map) bool {
	tileSize := gameMap.Data.TileWidth
	for _, whom := range e.whom {
		if gameMap.Layers[1][int(player.X)/tileSize][int(player.Y)/tileSize] == whom ||
			gameMap.Layers[1][int(player.X+1)/tileSize][int(player.Y)/tileSize] == whom ||
			gameMap.Layers[1][int(player.X)/tileSize][int(player.Y+1)/tileSize] == whom ||
			gameMap.Layers[1][int(player.X-1)/tileSize][int(player.Y)/tileSize] == whom ||
			gameMap.Layers[1][int(player.X)/tileSize][int(player.Y-1)/tileSize] == whom {
			return true
		}
	}
	return false
}
