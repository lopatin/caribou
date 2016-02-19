package caribou

import (
	"reflect"
	"errors"
	riak "github.com/basho/riak-go-client"
)

// LoadRiakModel reads a riak.FetchMapResponse into the given model.
func LoadRiakModel(r interface{}, m Model) error {
	switch r.(type) {
	default:
		return errors.New("Invalid response type")
	case *riak.FetchMapResponse:
		// Convert the riak.Map CRDT data structure into a Go map.
		gomap, err := RiakMapToMap(*r.(*riak.FetchMapResponse).Map)
		if err != nil {
			return err
		}

		// Load the Go map into the model.
		err = LoadMapIntoModel(gomap, m)
		if err != nil {
			return err
		}

		// Save the context.
		m.SetContext(string(r.(*riak.FetchMapResponse).Context))
		return nil
	case *riak.UpdateMapResponse:
		// Convert the riak.Map CRDT data structure into a Go map.
		gomap, err := RiakMapToMap(*r.(*riak.UpdateMapResponse).Map)
		if err != nil {
			return err
		}

		// Load the Go map into the model.
		err = LoadMapIntoModel(gomap, m)
		if err != nil {
			return err
		}

		// Save the context.
		m.SetContext(string(r.(*riak.UpdateMapResponse).Context))
		return nil
	}
}

// BuildMapOperation builds a Riak CRDT MapOperation that is required to convert the model's
// snapshot to it's current state.
func BuildMapOperation(m Model) (*riak.MapOperation, error) {
	var op riak.MapOperation
	from := m.GetSnapshot()
	to := ToMap(m, true)
	err := fillMapOp(from, to, &op)
	return &op, err
}

// Recursive helper function for BuildMapOperation.
func fillMapOp(from map[string]interface{}, to map[string]interface{},
op *riak.MapOperation) error {

	// Remove fields that exist in `from` but not in `to`.
	for k, v := range from {
		if to[k] == nil || reflect.ValueOf(v).Kind() != reflect.ValueOf(to[k]).Kind() {
			switch reflect.ValueOf(v).Kind() {
			case reflect.Bool:
				op.RemoveFlag(k)
			case reflect.Map:
				op.RemoveMap(k)
			case reflect.Slice, reflect.Array:
				op.RemoveSet(k)
			case reflect.String, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32,
				reflect.Float64:
				op.RemoveRegister(k)
			}
		}
	}

	// Set fields.
	for k, v := range to {
		switch reflect.ValueOf(v).Kind() {
		default:
			return errors.New("Unrecognized field type")
		case reflect.Map:
			// Maps are handled recursively.
			if reflect.ValueOf(from[k]).Kind() == reflect.Map {
				fillMapOp(from[k].(map[string]interface{}), v.(map[string]interface{}), op.Map(k))
			} else {
				fillMapOp(map[string]interface{}{}, v.(map[string]interface{}), op.Map(k))
			}
		case reflect.Slice, reflect.Array:
			// If both the new and previous values are arrays then diff the arrays as sets and
			// register the necessary AddToSet and RemoveFromSets operations.
			fromArr := []string{}
			if reflect.ValueOf(from[k]).Kind() == reflect.Slice ||
			reflect.ValueOf(from[k]).Kind() == reflect.Array {
				fromArr = from[k].([]string)
			}

			// Build maps of item => true for both arrays. Only supports string lists.
			// TODO(lopatin): Detect if this is a byte array, and return an error because it is
			// not supported. Or you know ... add support for byte arrays.
			mv := make(map[string]bool)
			mf := make(map[string]bool)
			for _, s := range v.([]string) {
				if reflect.ValueOf(s).Kind() == reflect.String {
					mv[s] = true
				}
			}
			for _, s := range fromArr {
				if reflect.ValueOf(s).Kind() == reflect.String {
					mf[s] = true
				}
			}

			// Remove items from set if they don't exist in the target.
			for key := range mf {
				if !mv[key] {
					op.RemoveFromSet(k, []byte(key))
				}
			}

			// Add items that aren't in the set yet.
			for key := range mv {
				if !mf[key] {
					op.AddToSet(k, []byte(key))
				}
			}
		case reflect.Bool:
			op.SetFlag(k, v.(bool))
		case reflect.String, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32,
			reflect.Float64:
			reg, err := encodeRegister(v)
			if err != nil {
				return err
			}
			op.SetRegister(k, []byte(reg))
		}
	}

	return nil
}

// RiakMapToMap converts a riak.Map CRDT struct to a plain Go map. Nested maps are handled
// recursively. Counters are ignored because our data model doesn't support them. This function
// also tries to decode registers into numbers and only treats them as strings if decoding fails.
func RiakMapToMap(rm riak.Map) (map[string]interface{}, error) {
	// Create a new empty map.
	gm := make(map[string]interface{})

	// Copy over the sets. Convert v from [][]byte to []string.
	for k, v := range rm.Sets {
		vs := []string{}
		for _, item := range v {
			vs = append(vs, string(item))
		}
		gm[k] = vs
	}

	// Copy over the flags.
	for k, v := range rm.Flags {
		gm[k] = v
	}

	// Copy over the registers.
	for k, v := range rm.Registers {
		register, err := decodeRegister(string(v))
		if err != nil {
			return gm, err
		}
		gm[k] = register
	}

	// Recursively copy over the maps.
	for k, v := range rm.Maps {
		m, err := RiakMapToMap(*v)
		if err != nil {
			return gm, err
		}
		gm[k] = m
	}

	return gm, nil
}

// StoreModelInRiak saves the model in Riak using CRDT map operations.
func StoreModelInRiak(model Model, bucketName, key string, rs *RiakService) error {
	// Build the update map CRDT operation.
	op, err := BuildMapOperation(model)
	if err != nil {
		return err
	}

	// Build the update command.
	builder := riak.NewUpdateMapCommandBuilder().
	WithBucket(bucketName).
	WithBucketType(BucketTypeMaps).
	WithKey(key).
	WithReturnBody(true).
	WithMapOperation(op)

	// Attach context
	ctx := model.GetContext()
	if len(ctx) > 0 {
		builder.WithContext([]byte(ctx))
	}

	updateMapCmd, err := builder.Build()
	if err != nil {
		return err
	}

	// Run the command.
	err = rs.Exec(func(client *riak.Client) error {
		return client.Execute(updateMapCmd)
	})
	if err != nil {
		return err
	}

	// Load the response into the model.
	cmd := updateMapCmd.(*riak.UpdateMapCommand)
	return LoadRiakModel(cmd.Response, model)
}

// FindRiakModelByKey finds the Riak map with the given key and loads it into the specified model.
func FindRiakModelByKey(model Model, bucketName, key string, rs *RiakService) (bool, error) {
	// Create the command that will fetch the user map from Riak.
	cmd, err := riak.NewFetchMapCommandBuilder().
	WithBucket(bucketName).
	WithBucketType(BucketTypeMaps).
	WithKey(key).
	Build()
	if err != nil {
		return false, err
	}

	// Run the command
	err = rs.Exec(func(client *riak.Client) error {
		return client.Execute(cmd)
	})
	if err != nil {
		return false, err
	}

	// Check if not found
	fetchMapCmd := cmd.(*riak.FetchMapCommand)
	if fetchMapCmd.Response.IsNotFound || fetchMapCmd.Response.Map == nil {
		return false, nil
	}

	// Load the map
	return true, LoadRiakModel(fetchMapCmd.Response, model)
}

type Contexter interface {
	GetContext() string
	SetContext(string)
}
