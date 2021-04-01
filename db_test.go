package main

import (
	"fmt"
	"os"
	"testing"
)

func getDBManager(t *testing.T, dataFiles ...string) *dBManager {
	t.Helper()
	manager, err := NewDBManager(":memory:")
	if err != nil {
		t.Fatalf("failed to create db instance: %v", err)
	}

	// read schema file and execute it always
	// this also checks that schema.sql is in good shape, since during
	// the build it is only embedded
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		t.Fatalf("failed to read sql schema file")
		return nil
	}
	if err := manager.rawExec(string(schema)); err != nil {
		t.Fatalf("failed to crate db schema: %v", err)
		return nil
	}

	// load all provided data files
	for _, dataFile := range dataFiles {
		data, err := os.ReadFile(dataFile)
		if err != nil {
			t.Fatalf("unable to read datafile %q: %v", dataFile, err)
		}
		if err := manager.rawExec(string(data)); err != nil {
			t.Fatalf("failed to execute data file: %v", err)
		}
	}
	return manager
}

func TestDBManager_UserByID(t *testing.T) {
	manager := getDBManager(t, "testdata/test.sql")

	data := []struct {
		id           int
		expectToFind bool
	}{
		{1, true},
		{99, false},
	}

	for _, d := range data {
		d := d
		t.Run(fmt.Sprintf("%d - %v", d.id, d.expectToFind), func(t *testing.T) {
			u, err := manager.UserByID(d.id)
			if d.expectToFind {
				if err != nil {
					t.Fatalf("expected to find user, but did not get one")
					return
				}
				// sanity check that we got the same user ID we asked for
				if u.ID != d.id {
					t.Fatalf("unexpected user ID, got: %d, expected: %d", u.ID, d.id)
				}
			} else {
				if err == nil {
					t.Fatalf("did not expect to find user, but got: %v", u)
				}
			}

		})
	}
}
