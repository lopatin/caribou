package caribou

import (
	"testing"
	"fmt"
)

//type Account struct {
//	ModelMetadata
//	Country string
//}

//func (m *Account) Migrations() ([]*Migration) {
//	return []*Migration{
//		&Migration{"state_to_country", func(m map[string]interface{}) map[string]interface{} {
//			if m["State"] != nil {
//				m["Country"] = "US of A"
//			}
//			delete(m, "State")
//			return m
//		}},
//	};
//}

func TestMigration(t *testing.T) {
	var a Account
	err := LoadJSONModel([]byte(`{"ModelMetadata": {"Version": ""}, "State": "Texas"}`), &a)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(a.Country)
}
