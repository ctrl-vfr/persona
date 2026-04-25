// Package watcher exposes filesystem watchers and a multi-instance
// coordination layer (.instances.json) for the persona TUI. It uses
// fsnotify to react to local edits and to changes performed by other
// running instances.
package watcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ctrl-vfr/persona/internal/persona"
	"github.com/ctrl-vfr/persona/internal/storage"

	"github.com/fsnotify/fsnotify"
)

// maxInstancesFileBytes caps how much we read from .instances.json to
// guard against a corrupted or maliciously inflated lock file.
const maxInstancesFileBytes = 1 << 20 // 1 MiB

type PersonaWatcher struct {
	watcher         *fsnotify.Watcher
	manager         *storage.Manager
	personaName     string
	onUpdate        func(*persona.Persona)
	onHistoryUpdate func([]persona.Message)
	cancel          context.CancelFunc
}

type InstanceManager struct {
	lockFilePath string
	instanceID   string
	manager      *storage.Manager
}

// NewPersonaWatcher creates a new file watcher for a persona
func NewPersonaWatcher(manager *storage.Manager, personaName string) (*PersonaWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	pw := &PersonaWatcher{
		watcher:     watcher,
		manager:     manager,
		personaName: personaName,
	}

	// Watch the persona's directory
	personaPath, historyPath := manager.GetPersonaPath(personaName)
	personaDir := filepath.Dir(personaPath)

	if err := watcher.Add(personaDir); err != nil {
		err := watcher.Close()
		if err != nil {
			log.Printf("Warning: failed to close watcher: %v", err)
		}
		return nil, fmt.Errorf("failed to watch persona directory: %w", err)
	}

	// Also watch the history file specifically
	if err := watcher.Add(historyPath); err != nil {
		// History file might not exist yet, that's okay
		log.Printf("Warning: could not watch history file %s: %v", historyPath, err)
	}

	return pw, nil
}

// SetOnUpdate sets the callback for persona updates
func (pw *PersonaWatcher) SetOnUpdate(callback func(*persona.Persona)) {
	pw.onUpdate = callback
}

// SetOnHistoryUpdate sets the callback for history updates
func (pw *PersonaWatcher) SetOnHistoryUpdate(callback func([]persona.Message)) {
	pw.onHistoryUpdate = callback
}

// Start begins watching for file changes. The goroutine exits when
// ctx is cancelled or when Stop is called.
func (pw *PersonaWatcher) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	pw.cancel = cancel
	go func() {
		for {
			select {
			case event, ok := <-pw.watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					pw.handleFileChange(event.Name)
				}
			case err, ok := <-pw.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Watcher error: %v", err)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop stops the file watcher. Safe to call multiple times.
func (pw *PersonaWatcher) Stop() {
	if pw.cancel != nil {
		pw.cancel()
	}
	if err := pw.watcher.Close(); err != nil {
		log.Printf("Warning: failed to close watcher: %v", err)
	}
}

// handleFileChange processes file change events
func (pw *PersonaWatcher) handleFileChange(filename string) {
	// Add a small delay to avoid multiple rapid events
	time.Sleep(100 * time.Millisecond)

	basename := filepath.Base(filename)

	switch basename {
	case "persona.json":
		if pw.onUpdate != nil {
			if p, err := pw.manager.GetPersona(pw.personaName); err == nil {
				pw.onUpdate(p)
			}
		}

	case "history.json":
		if pw.onHistoryUpdate != nil {
			if p, err := pw.manager.GetPersona(pw.personaName); err == nil {
				pw.onHistoryUpdate(p.History)
			}
		}
	}
}

// NewInstanceManager creates a new instance manager
func NewInstanceManager(manager *storage.Manager) *InstanceManager {
	instanceID := fmt.Sprintf("persona_%d_%d", os.Getpid(), time.Now().Unix())
	lockFilePath := filepath.Join(manager.BasePath, ".instances.json")

	return &InstanceManager{
		lockFilePath: lockFilePath,
		instanceID:   instanceID,
		manager:      manager,
	}
}

// RegisterInstance registers this instance
func (im *InstanceManager) RegisterInstance() error {
	instances, err := im.loadInstances()
	if err != nil {
		instances = make(map[string]InstanceInfo)
	}

	instances[im.instanceID] = InstanceInfo{
		PID:       os.Getpid(),
		StartTime: time.Now(),
		LastSeen:  time.Now(),
	}

	return im.saveInstances(instances)
}

// UpdateLastSeen updates the last seen time for this instance
func (im *InstanceManager) UpdateLastSeen() error {
	instances, err := im.loadInstances()
	if err != nil {
		return err
	}

	if info, exists := instances[im.instanceID]; exists {
		info.LastSeen = time.Now()
		instances[im.instanceID] = info
		return im.saveInstances(instances)
	}

	return fmt.Errorf("instance not found")
}

// UnregisterInstance removes this instance
func (im *InstanceManager) UnregisterInstance() error {
	instances, err := im.loadInstances()
	if err != nil {
		return err
	}

	delete(instances, im.instanceID)
	return im.saveInstances(instances)
}

// GetActiveInstances returns all active instances
func (im *InstanceManager) GetActiveInstances() (map[string]InstanceInfo, error) {
	instances, err := im.loadInstances()
	if err != nil {
		return nil, err
	}

	// Clean up stale instances (not seen for more than 5 minutes)
	now := time.Now()
	for id, info := range instances {
		if now.Sub(info.LastSeen) > 5*time.Minute {
			delete(instances, id)
		}
	}

	// Save cleaned instances
	if err := im.saveInstances(instances); err != nil {
		log.Printf("Warning: failed to save cleaned instances: %v", err)
	}

	return instances, nil
}

// InstanceInfo holds information about a running instance
type InstanceInfo struct {
	PID       int       `json:"pid"`
	StartTime time.Time `json:"start_time"`
	LastSeen  time.Time `json:"last_seen"`
}

// loadInstances loads instance information from file. The read is
// bounded to avoid OOM if the lock file is corrupted/oversized.
func (im *InstanceManager) loadInstances() (map[string]InstanceInfo, error) {
	f, err := os.Open(im.lockFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]InstanceInfo), nil
		}
		return nil, fmt.Errorf("open instances file: %w", err)
	}
	defer func() { _ = f.Close() }()

	data, err := io.ReadAll(io.LimitReader(f, maxInstancesFileBytes))
	if err != nil {
		return nil, fmt.Errorf("read instances file: %w", err)
	}

	var instances map[string]InstanceInfo
	if err := json.Unmarshal(data, &instances); err != nil {
		return nil, fmt.Errorf("parse instances file: %w", err)
	}
	return instances, nil
}

// saveInstances saves instance information to file atomically.
func (im *InstanceManager) saveInstances(instances map[string]InstanceInfo) error {
	data, err := json.MarshalIndent(instances, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal instances: %w", err)
	}
	return storage.WriteFileAtomic(im.lockFilePath, data, storage.UserFileMode)
}

// StartHeartbeat starts a heartbeat goroutine to keep this instance
// marked as active. It exits when ctx is cancelled.
func (im *InstanceManager) StartHeartbeat(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := im.UpdateLastSeen(); err != nil {
					log.Printf("Warning: failed to update last seen: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
