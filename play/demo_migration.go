package main

import (
	"fmt"
	"mapstructure"
	"encoding/json"
)

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


// A Caribou is a data container that represents data at a certain version. Versions are strings
// and every migration contains the destination version in addition to a function that performs
// the data migration to the next version.
type Caribou interface {
	// Migrations returns a chronologically ordered list of migrations. All Migration names must
	// be unique within the scope of this Caribou.
	Migrations() []*Migration

	// GetVersion returns the current version of this Caribou.
	GetVersion() string

	// SetVersion sets a new version for the Caribou. Note that this is a simple setter and
	// doesn't actually migrate anything.
	SetVersion(string)
}

// A Migration is a named map -> map function.
type Migration struct {
	// Name is the title or id of this migration that should briefly describe the transformation
	// happening in Migrate. This will be the version name of any Caribou that this Migration is
	// performed on until the next migration happens.
	Name    string

	// Migrate is the transformation function that turns from its previous version to the version
	// represented by this Migration.
	Migrate func(map[string]interface{}) map[string]interface{}
}

// FastForwardMap migrates the given map from current version of the Caribou to the latest
// version.
func FastForwardMap(c Caribou, mp map[string]interface{}) (map[string]interface{}, error) {
	var fastFwd func(map[string]interface{}, string) (map[string]interface{}, error)

	fastFwd = func(mp map[string]interface{}, v string) (map[string]interface{}, error) {
		// Return if there are no migrations or we are at the latest one.
		if len(c.Migrations()) == 0 ||
		c.Migrations()[len(c.Migrations())-1].Name == v {

			return mp, nil
		}

		// Otherwise migrate up and recurse on FastForward.
		migrationIndex := -1
		for i, v := range c.Migrations() {
			if v.Name == c.GetVersion() {
				migrationIndex = i
				break
			}
		}
		c.Migrations()[migrationIndex+1].Migrate(mp)
		return fastFwd(mp, c.Migrations()[migrationIndex+1].Name)
	}

	return fastFwd(mp, c.GetVersion())
}

type Snapshotter interface {
	GetSnapshot() map[string]interface{}
	SetSnapshot(map[string]interface{})
}


type Model interface {
	Caribou
	Snapshotter
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

	m, err = FastForwardMap(model, m)
	if err != nil {
		return err
	}

	// Load fast forwarded map now.
	err = mapstructure.Decode(m, model)
	if err != nil {
		return err
	}

	// Set the snapshot to the map that we are loading from.
	model.SetSnapshot(m)

	// Make sure that we are at the latest version of the model.
	//	err = FastForward(model, m)
	//	if err != nil {
	//		return err
	//	}

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

// START OMIT
type Account struct {
	ModelMetadata
	Country string
}

func (m *Account) Migrations() ([]*Migration) {
	return []*Migration{
		&Migration{"state_to_country", func(m map[string]interface{}) map[string]interface{} {
			if m["State"] != nil {
				m["Country"] = "US of A"
			}
			delete(m, "State")
			return m
		}},
	};
}

func main() {
	var a Account
	LoadJSONModel([]byte(`{"ModelMetadata": {"Version": ""}, "State": "Texas"}`), &a)
	fmt.Println(a.Country)
}
// END OMIT
