package caribou

type Snapshotter interface {
	GetSnapshot() map[string]interface{}
	SetSnapshot(map[string]interface{})
}
