package caribou

import (
	"github.com/mitchellh/mapstructure"
	"encoding/json"
)

type Model interface {
	Caribou
	Snapshotter
	Contexter
}


// LoadMapIntoModel fills in the model data with the contents of the map. The supplied map is
// stored as the model snapshot.
func LoadMapIntoModel(m map[string]interface{}, model Model) error {

	// Load map into struct. This sets the metadata, although the actual fields may be garbled
	// due to not being migrated yet.
	err := mapstructure.Decode(m, model)
	if err != nil {
		return err
	}

	// Return the fast forwarded version of m
	m, err = FastForwardMap(model, m)
	if err != nil {
		return err
	}

	// Load fast forwarded map into struct now.
	err = mapstructure.Decode(m, model)
	if err != nil {
		return err
	}

	// Set the snapshot to the map that we are loading from.
	model.SetSnapshot(m)

	return nil
}

// LoadJSONModel is the same as LoadMap except it acceps the input as a JSON byte array instead
// of a map.
func LoadJSONModel(data []byte, model Model) error {
	var m map[string]interface{}

	// Unmarshal JSON into a map.
	err := json.Unmarshal(data, &m)
	if err != nil {

		return err
	}

	// Load model from the map.
	err = LoadMapIntoModel(m, model)
	if err != nil {
		return err
	}

	return nil
}

