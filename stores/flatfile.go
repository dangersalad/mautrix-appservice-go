// Package stores provides store implementations
package stores

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"maunium.net/go/mautrix"
	appservice "maunium.net/go/mautrix-appservice"
)

type flatfileStore struct {
	*appservice.BasicStateStore
	filename string
	file     *os.File
}

// NewFileStore returns a StateStore that just saves/loads the data from a flat file
func NewFileStore(filename string) (appservice.StateStore, error) {
	store := &flatfileStore{
		BasicStateStore: &appservice.BasicStateStore{
			Registrations:    make(map[string]bool),
			Members:          make(map[string]map[string]mautrix.Member),
			PowerLevels:      make(map[string]*mautrix.PowerLevels),
			TypingStateStore: appservice.NewTypingStateStore(),
		},
		filename: filename,
	}

	if err := store.load(); err != nil {
		return nil, fmt.Errorf("initializing from file %w", err)
	}

	return store, nil
}

func (store *flatfileStore) save() error {
	f, err := os.Create(store.filename)
	if err != nil {
		return fmt.Errorf("opening state file %s for writing: %w", store.filename, err)
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(store.BasicStateStore); err != nil {
		return fmt.Errorf("writing state to file %s: %w", store.filename, err)
	}
	return nil
}

func (store *flatfileStore) load() error {
	_, err := os.Stat(store.filename)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("getting file info for %s: %w", store.filename, err)
		}
		// if the file does not exist, that's fine, just return
		return nil
	}

	f, err := os.Open(store.filename)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("opening file %s for reading: %w", store.filename, err)
	}

	dec := json.NewDecoder(f)
	if err := dec.Decode(store.BasicStateStore); err != nil {
		return fmt.Errorf("parsing existing state file %s: %w", store.filename, err)
	}

	return nil

}

func (store *flatfileStore) IsRegistered(userID string) bool {
	return store.BasicStateStore.IsRegistered(userID)
}
func (store *flatfileStore) MarkRegistered(userID string) {
	store.BasicStateStore.MarkRegistered(userID)
	if err := store.save(); err != nil {
		panic(err)
	}
}

func (store *flatfileStore) IsTyping(roomID, userID string) bool {
	return store.BasicStateStore.IsTyping(roomID, userID)
}
func (store *flatfileStore) SetTyping(roomID, userID string, timeout int64) {
	store.BasicStateStore.SetTyping(roomID, userID, timeout)
	if err := store.save(); err != nil {
		panic(err)
	}
}

func (store *flatfileStore) IsInRoom(roomID, userID string) bool {
	return store.BasicStateStore.IsInRoom(roomID, userID)
}
func (store *flatfileStore) IsInvited(roomID, userID string) bool {
	return store.BasicStateStore.IsInvited(roomID, userID)
}
func (store *flatfileStore) IsMembership(roomID, userID string, allowedMemberships ...mautrix.Membership) bool {
	return store.BasicStateStore.IsMembership(roomID, userID, allowedMemberships...)
}
func (store *flatfileStore) GetMember(roomID, userID string) mautrix.Member {
	return store.BasicStateStore.GetMember(roomID, userID)
}
func (store *flatfileStore) TryGetMember(roomID, userID string) (mautrix.Member, bool) {
	return store.BasicStateStore.TryGetMember(roomID, userID)
}
func (store *flatfileStore) SetMembership(roomID, userID string, membership mautrix.Membership) {
	store.BasicStateStore.SetMembership(roomID, userID, membership)
	if err := store.save(); err != nil {
		panic(err)
	}
}
func (store *flatfileStore) SetMember(roomID, userID string, member mautrix.Member) {
	store.BasicStateStore.SetMember(roomID, userID, member)
	if err := store.save(); err != nil {
		panic(err)
	}
}

func (store *flatfileStore) SetPowerLevels(roomID string, levels *mautrix.PowerLevels) {
	store.BasicStateStore.SetPowerLevels(roomID, levels)
	if err := store.save(); err != nil {
		panic(err)
	}
}
func (store *flatfileStore) GetPowerLevels(roomID string) *mautrix.PowerLevels {
	return store.BasicStateStore.GetPowerLevels(roomID)
}
func (store *flatfileStore) GetPowerLevel(roomID, userID string) int {
	return store.BasicStateStore.GetPowerLevel(roomID, userID)

}
func (store *flatfileStore) GetPowerLevelRequirement(roomID string, eventType mautrix.EventType) int {
	return store.BasicStateStore.GetPowerLevelRequirement(roomID, eventType)

}
func (store *flatfileStore) HasPowerLevel(roomID, userID string, eventType mautrix.EventType) bool {
	return store.BasicStateStore.HasPowerLevel(roomID, userID, eventType)
}
