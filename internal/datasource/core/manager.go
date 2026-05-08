package core

import (
	"context"
	"fmt"
	"sync"

	ds "github.com/muyue/comic-harmony-backend/internal/datasource/model"
)

type DataSourceManager struct {
	mu       sync.RWMutex
	sources  map[int64]IDataSource
	registry map[ds.DataSourceType]SourceFactory
}

type SourceFactory func(config ds.DataSourceConfig) (IDataSource, error)

func NewDataSourceManager() *DataSourceManager {
	return &DataSourceManager{
		sources:  make(map[int64]IDataSource),
		registry: make(map[ds.DataSourceType]SourceFactory),
	}
}

func (m *DataSourceManager) Register(sourceType ds.DataSourceType, factory SourceFactory) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.registry[sourceType] = factory
}

func (m *DataSourceManager) Add(config ds.DataSourceConfig) (IDataSource, error) {
	m.mu.RLock()
	factory, ok := m.registry[config.Type]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unsupported source type: %s", config.Type)
	}
	source, err := factory(config)
	if err != nil {
		return nil, fmt.Errorf("create source: %w", err)
	}
	m.mu.Lock()
	m.sources[config.ID] = source
	m.mu.Unlock()
	return source, nil
}

func (m *DataSourceManager) Remove(id int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sources, id)
}

func (m *DataSourceManager) Get(id int64) (IDataSource, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sources[id]
	return s, ok
}

func (m *DataSourceManager) List() []ds.DataSourceConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var configs []ds.DataSourceConfig
	for _, s := range m.sources {
		configs = append(configs, s.Config())
	}
	return configs
}

func (m *DataSourceManager) HealthCheckAll(ctx context.Context) map[int64]error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	results := make(map[int64]error)
	for id, source := range m.sources {
		if err := source.HealthCheck(ctx); err != nil {
			results[id] = err
		}
	}
	return results
}

// --- Global singleton ---

var globalManager = NewDataSourceManager()

func Global() *DataSourceManager {
	return globalManager
}

func RegisterFactory(sourceType ds.DataSourceType, factory SourceFactory) {
	globalManager.Register(sourceType, factory)
}
