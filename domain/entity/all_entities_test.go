package entity

import "testing"

func TestAllTablesName(t *testing.T) {
	type tableNamer interface {
		TableName() string
	}
	allTables := AllTables()

	for _, table := range allTables {
		strct, ok := table.(tableNamer)
		if !ok {
			t.Error("error while getting tableNamer")
		}
		name := strct.TableName()
		if name == "" {
			t.Errorf("TableName for %T should not be empty: " + name)
		}
	}
}
