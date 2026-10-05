package mongoq

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestParseQueryRejectsUnsupportedCallTargets(t *testing.T) {
	for _, tc := range []struct {
		name, query, targetType string
	}{
		{"call result", `exists(foo)(bar)`, "*ast.CallExpr"},
		{"missing conjunction", `exists(foo) (status == "online")`, "*ast.CallExpr"},
		{"call result in comparison", `name == contains("foo")("bar")`, "*ast.CallExpr"},
		{"selector", `foo.bar("x")`, "*ast.SelectorExpr"},
		{"parenthesized function", `(exists)(foo)`, "*ast.ParenExpr"},
		{"indexed function", `foo[0](bar)`, "*ast.IndexExpr"},
		{"literal", `"foo"(bar)`, "*ast.BasicLit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			query, err := ParseQuery(tc.query)
			require.ErrorContains(t, err, "unsupported function expression: "+tc.targetType)
			require.Nil(t, query)
		})
	}
}

func TestParseQuerySupportedCallTargets(t *testing.T) {
	for _, tc := range []struct {
		query string
		want  bson.M
	}{
		{`exists(foo) && (status == "online")`, bson.M{"foo": bson.M{"$exists": true}, "status": "online"}},
		{`name == contains("foo")`, bson.M{"name": bson.Regex{Pattern: ".*foo.*", Options: "i"}}},
	} {
		t.Run(tc.query, func(t *testing.T) {
			query, err := ParseQuery(tc.query)
			require.NoError(t, err)
			require.Equal(t, tc.want, query)
		})
	}
}
