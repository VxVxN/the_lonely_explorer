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
	whom   []int
	filter func(x, y int) bool
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

// NewMeetEventWhere works like NewMeetEvent, but only the tiles accepted by filter (tile coordinates) count.
func NewMeetEventWhere(whom []int, filter func(x, y int) bool, action func()) *MeetEvent {
	event := NewMeetEvent(whom, action)
	event.filter = filter
	return event
}

func (e *MeetEvent) Check(player *player.Player, gameMap *_map.Map) bool {
	tileSize := gameMap.Data.TileWidth
	probes := [][2]int{
		{int(player.X) / tileSize, int(player.Y) / tileSize},
		{int(player.X+1) / tileSize, int(player.Y) / tileSize},
		{int(player.X) / tileSize, int(player.Y+1) / tileSize},
		{int(player.X-1) / tileSize, int(player.Y) / tileSize},
		{int(player.X) / tileSize, int(player.Y-1) / tileSize},
	}
	for _, whom := range e.whom {
		for _, probe := range probes {
			if gameMap.Layers[1][probe[0]][probe[1]] != whom {
				continue
			}
			if e.filter == nil || e.filter(probe[0], probe[1]) {
				return true
			}
		}
	}
	return false
}
