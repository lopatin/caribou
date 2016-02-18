package caribou

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
