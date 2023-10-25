package room

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sync"
)

type Manager struct {
	Logger *zap.Logger
	Mu     sync.Mutex
	Rooms  map[string]*Room
}

func NewRoomManager(
	logger *zap.Logger,
) (*Manager, error) {
	return &Manager{
		Logger: logger,
		Mu:     sync.Mutex{},
		Rooms:  make(map[string]*Room),
	}, nil
}

// IsRoomExists checks if a room exists.
func (h *Manager) IsRoomExists(roomName string) bool {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	_, ok := h.Rooms[roomName]
	return ok
}

// AddNewRoom adds a new room to the handler.
func (h *Manager) AddNewRoom(roomName string) (*Room, error) {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	h.Rooms[roomName] = NewRoom(roomName)

	return h.Rooms[roomName], nil
}

var ErrRoomNotFound = errors.New("room not found")

// GetRoom returns a room by its name.
func (h *Manager) GetRoom(roomName string) (*Room, error) {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	if foundRoom, found := h.Rooms[roomName]; found {
		return foundRoom, nil
	}

	return nil, ErrRoomNotFound
}

// DeleteRoom deletes a room by its name.
func (h *Manager) DeleteRoom(roomName string) error {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	delete(h.Rooms, roomName)
	return nil
}
