package config

import (
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type ConfigWatcher struct {
	watcher     *fsnotify.Watcher
	config     *MultiProviderConfig
	onChange   []func(*MultiProviderConfig)
	mu         sync.RWMutex
	stopChan   chan struct{}
}

var globalWatcher *ConfigWatcher
var watcherOnce sync.Once

func GetWatcher() *ConfigWatcher {
	watcherOnce.Do(func() {
		globalWatcher = &ConfigWatcher{
			onChange: make([]func(*MultiProviderConfig), 0),
			stopChan: make(chan struct{}),
		}
	})
	return globalWatcher
}

func (cw *ConfigWatcher) Start(cfg *MultiProviderConfig) error {
	cw.mu.Lock()
	cw.config = cfg
	cw.mu.Unlock()

	if !cfg.HotReload.Enabled {
		return nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	cw.watcher = watcher

	configPath := GetConfigPath()
	if err := watcher.Add(configPath); err != nil {
		return err
	}

	go cw.watchLoop()
	log.Printf("[D] Config watcher started, watching: %s", configPath)
	return nil
}

func (cw *ConfigWatcher) Stop() error {
	if cw.watcher != nil {
		close(cw.stopChan)
		return cw.watcher.Close()
	}
	return nil
}

func (cw *ConfigWatcher) OnChange(callback func(*MultiProviderConfig)) {
	cw.onChange = append(cw.onChange, callback)
}

func (cw *ConfigWatcher) watchLoop() {
	debounce := time.NewTicker(time.Duration(cw.config.HotReload.DebounceMs) * time.Millisecond)
	defer debounce.Stop()

	for {
		select {
		case event, ok := <-cw.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				select {
				case <-debounce.C:
					cw.reload()
				default:
				}
			}
		case err, ok := <-cw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("[!] Config watcher error: %v", err)
		case <-cw.stopChan:
			return
		}
	}
}

func (cw *ConfigWatcher) reload() {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	log.Printf("[D] Reloading config from disk...")

	newCfg, err := LoadMultiProvider()
	if err != nil {
		log.Printf("[!] Failed to reload config: %v", err)
		return
	}

	cw.config = newCfg

	for _, callback := range cw.onChange {
		go callback(newCfg)
	}

	log.Printf("[D] Config reloaded successfully, active provider: %s", newCfg.Active)
}

func (cw *ConfigWatcher) GetConfig() *MultiProviderConfig {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.config
}

func (cw *ConfigWatcher) UpdateConfig(cfg *MultiProviderConfig) {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	if err := SaveMultiProvider(cfg); err != nil {
		log.Printf("[!] Failed to save config: %v", err)
		return
	}

	cw.config = cfg
}

func GetConfigDirectory() string {
	configPath := GetConfigPath()
	return filepath.Dir(configPath)
}
