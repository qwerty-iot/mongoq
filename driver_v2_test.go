package mongoq

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"testing"
)

// Assert wire types, not only Go equality: a v1 ObjectID or Regex hidden behind
// any would otherwise silently encode as the wrong BSON type with the v2 driver.
func TestDriverV2WireTypes(t *testing.T) {
	for _, tc := range []struct {
		query, field string
		kind         bson.Type
	}{
		{`_id == "507f1f77bcf86cd799439011"`, "_id", bson.TypeObjectID},
		{`name == contains(sensor)`, "name", bson.TypeRegex},
		{`ts == date("2026-09-14T12:00:00Z")`, "ts", bson.TypeDateTime},
		{`name == ("sensor" | "tracker")`, "name", bson.TypeEmbeddedDocument},
	} {
		t.Run(tc.query, func(t *testing.T) {
			query, err := ParseQuery(tc.query)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := bson.Marshal(query)
			if err != nil {
				t.Fatal(err)
			}
			if got := bson.Raw(raw).Lookup(tc.field).Type; got != tc.kind {
				t.Fatalf("got %v, want %v", got, tc.kind)
			}
			if tc.field == "_id" {
				if bson.Raw(raw).Lookup("_id").ObjectID().Hex() != "507f1f77bcf86cd799439011" {
					t.Fatal("ObjectID changed")
				}
			}
		})
	}
}
