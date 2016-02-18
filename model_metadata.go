package caribou

type ModelMetadata struct {
	Version string
	Snapshot map[string]interface{}
}

//
// Implement Caribou
//

func (m *ModelMetadata) Migrations() []*Migration {
	return []*Migration{};
}

func (m *ModelMetadata) GetVersion() string {
	return m.Version;
}

func (m *ModelMetadata) SetVersion(version string) {
	m.Version = version;
}

//
// Implement Snapshotter
//

// GetSnapshot is the Snapshot getter.
func (m *ModelMetadata) GetSnapshot() map[string]interface{} {
	return m.Snapshot
}

// SetSnapshot is the Snapshot setter.
func (m *ModelMetadata) SetSnapshot(v map[string]interface{}) {
	m.Snapshot = v
}
