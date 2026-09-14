package mongoq

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/stretchr/testify/suite"
)

func TestReportSuite(t *testing.T) {
	suite.Run(t, new(ReportSuite))
}

type ReportSuite struct {
	suite.Suite
}

func (s *ReportSuite) SetupSuite() {

}

type queryVector struct {
	n string
	e string
	x string
	r bson.M
}

func (s *ReportSuite) testVectors(vectors []queryVector) {
	for _, vector := range vectors {
		rslt, err := ParseQuery(vector.e)
		if vector.x != "" {
			if err != nil {
				s.Equal(vector.x, err.Error())
			} else {
				s.Equal(vector.x, nil)
			}
			s.Nil(vector.r)
		} else {
			s.NoError(err)
			s.Equal(vector.r, rslt)
		}
	}
}

func (s *ReportSuite) TestOne() {

	vectors := []queryVector{
		{n: "in-bad-inner", e: "name == (\"Alice\" | \"Bob\" & \"Charlie\")", r: nil, x: "unsupported use of: &"}}
	s.testVectors(vectors)
}

func (s *ReportSuite) TestGoodQueries() {

	vectors := []queryVector{
		{n: "parens", e: "person.age >= 18 && (person.name == \"Alice\" || name == \"Bob\")", r: bson.M{"$or": []any{bson.M{"person.name": "Alice"}, bson.M{"name": "Bob"}}, "person.age": bson.M{"$gte": int64(18)}}},
		{n: "multi-and", e: "age>10 && height<5 && width==4", r: bson.M{"age": bson.M{"$gt": int64(10)}, "height": bson.M{"$lt": int64(5)}, "width": int64(4)}},
		{n: "int-float-bool", e: "age>10 && height<5.1 && dead==true", r: bson.M{"age": bson.M{"$gt": int64(10)}, "height": bson.M{"$lt": 5.1}, "dead": true}},
		{n: "bool-string", e: "dead==true && alive==\"false\"", r: bson.M{"alive": "false", "dead": true}},
		{n: "multi-or", e: "age>10 || height<5 || width==4", r: bson.M{"$or": []any{bson.M{"age": bson.M{"$gt": int64(10)}}, bson.M{"height": bson.M{"$lt": int64(5)}}, bson.M{"width": int64(4)}}}},
		{n: "and", e: "name != \"Bob\" && age > 18", r: bson.M{"age": bson.M{"$gt": int64(18)}, "name": bson.M{"$ne": "Bob"}}},
		{n: "exists-and", e: "name && age > 10", r: bson.M{"age": bson.M{"$gt": int64(10)}, "name": bson.M{"$exists": true}}},
		{n: "exists", e: "name", r: bson.M{"name": bson.M{"$exists": true}}},
		{n: "exists(name)", e: "exists(name)", r: bson.M{"name": bson.M{"$exists": true}}},
		{n: "exists(name.foo)", e: "exists(name.foo)", r: bson.M{"name.foo": bson.M{"$exists": true}}},
		{n: "complex", e: "age > 10 && exists(name.foo)", r: bson.M{"age": bson.M{"$gt": int64(10)}, "name.foo": bson.M{"$exists": true}}},
		{n: "exists-sub", e: "person.age", r: bson.M{"person.age": bson.M{"$exists": true}}},
		{n: "not-exists", e: "!name", r: bson.M{"name": bson.M{"$exists": false}}},
		{n: "noquotes1", e: "name == Alice", r: bson.M{"name": "Alice"}},
		{e: "age > 10 && (name || !desc)", r: bson.M{"$or": []any{bson.M{"name": bson.M{"$exists": true}}, bson.M{"desc": bson.M{"$exists": false}}}, "age": bson.M{"$gt": int64(10)}}},
		{e: "age > 10 && age < 20", r: bson.M{"$and": []any{bson.M{"age": bson.M{"$gt": int64(10)}}, bson.M{"age": bson.M{"$lt": int64(20)}}}}},
		{e: "_id == \"5fc4722ae367f19055977d1f\"", r: bson.M{"_id": bson.ObjectID{0x5f, 0xc4, 0x72, 0x2a, 0xe3, 0x67, 0xf1, 0x90, 0x55, 0x97, 0x7d, 0x1f}}},
		{n: "type", e: "\"type\" == \"Alice\"", r: bson.M{"type": "Alice"}},
		{n: "double-nested", e: "level1.level2.level3 == \"Alice\"", r: bson.M{"level1.level2.level3": "Alice"}},
		{n: "int-float-bool", e: "age>10 && height<5.1 && dead==true", r: bson.M{"age": bson.M{"$gt": int64(10)}, "height": bson.M{"$lt": 5.1}, "dead": true}},
		{n: "one-val", e: "id==(\"64d7b3661b467d611d5f1401\")", r: bson.M{"id": bson.ObjectID{0x64, 0xd7, 0xb3, 0x66, 0x1b, 0x46, 0x7d, 0x61, 0x1d, 0x5f, 0x14, 0x01}}},
		{n: "date-rfc3339", e: "ts==date(\"2020-12-01T00:00:00Z\")", r: bson.M{"ts": time.Date(2020, 12, 1, 0, 0, 0, 0, time.UTC)}},
		{n: "date-custom", e: "ts==date(\"20201201\",\"20060102\")", r: bson.M{"ts": time.Date(2020, 12, 1, 0, 0, 0, 0, time.UTC)}},
		{n: "fts", e: "\"data.accelerometer_3313.0.x_value_5702\">5", r: bson.M{"data.accelerometer_3313.0.x_value_5702": bson.M{"$gt": int64(5)}}},
	}
	s.testVectors(vectors)
}

func (s *ReportSuite) TestNestedQueries() {

	vectors := []queryVector{
		{n: "single", e: "level1 == \"Alice\"", r: bson.M{"level1": "Alice"}},
		{n: "single-nested", e: "level1.level2 == \"Alice\"", r: bson.M{"level1.level2": "Alice"}},
		{n: "double-nested", e: "level1.level2.level3 == \"Alice\"", r: bson.M{"level1.level2.level3": "Alice"}},
		{n: "triple-nested", e: "level1.level2.level3.level4 == \"Alice\"", r: bson.M{"level1.level2.level3.level4": "Alice"}},
	}
	s.testVectors(vectors)
}

func (s *ReportSuite) TestInNin() {

	vectors := []queryVector{
		{n: "in2", e: "name == (\"Alice\"| \"Bob\")", r: bson.M{"name": bson.M{"$in": []interface{}{"Alice", "Bob"}}}},
		{n: "in3", e: "name == (\"Alice\" | \"Bob\" | \"Charlie\")", r: bson.M{"name": bson.M{"$in": []any{"Alice", "Bob", "Charlie"}}}},
		{n: "in4", e: "name == (\"Alice\" | \"Bob\" | \"Charlie\" | \"Maya\")", r: bson.M{"name": bson.M{"$in": []any{"Alice", "Bob", "Charlie", "Maya"}}}},
		{n: "nin2", e: "name != (\"Alice\" | \"Bob\")", r: bson.M{"name": bson.M{"$nin": []any{"Alice", "Bob"}}}},
		{n: "nin3", e: "name != (\"Alice\" | \"Bob\" | \"Charlie\")", r: bson.M{"name": bson.M{"$nin": []any{"Alice", "Bob", "Charlie"}}}},
		{n: "nin4", e: "name != (\"Alice\" | \"Bob\" | \"Charlie\" | \"Maya\")", r: bson.M{"name": bson.M{"$nin": []any{"Alice", "Bob", "Charlie", "Maya"}}}},
		{n: "in-bad-op", e: "name > (\"Alice\" | \"Bob\" | \"Charlie\")", r: nil, x: "invalid right operand for operator '>'"},
		{n: "in-bad-inner", e: "name == (\"Alice\" | \"Bob\" & \"Charlie\")", r: nil, x: "unsupported use of: &"},
	}
	s.testVectors(vectors)
}

func (s *ReportSuite) TestAll() {

	vectors := []queryVector{
		{n: "in2", e: "name == (\"Alice\"| \"Bob\")", r: bson.M{"name": bson.M{"$in": []interface{}{"Alice", "Bob"}}}},
		{n: "all1", e: "name == (\"Alice\" & \"Bob\" & \"Charlie\")", r: bson.M{"name": bson.M{"$all": []any{"Alice", "Bob", "Charlie"}}}},
	}
	s.testVectors(vectors)
}

func (s *ReportSuite) TestBadQueries() {

	vectors := []queryVector{
		{n: "err-gr-string", e: "person.age >= \"test\"", r: nil, x: "invalid right operand for operator '>='"},
		{e: "_id == 5fc4722ae367f19055977d1f", x: "1:9: expected 'EOF', found fc4722ae367f19055977d1f"},
	}
	s.testVectors(vectors)
}

func (s *ReportSuite) TestRegexQueries() {

	vectors := []queryVector{
		{n: "regex1", e: "name == regex(\".*Alice.*\")", r: bson.M{"name": bson.Regex{Pattern: ".*Alice.*", Options: "i"}}},
		{n: "regex2", e: "name ==/.*Alice.*/", x: "1:8: expected operand, found '/'"},
		{n: "regex3", e: "name == contains(Alice)", r: bson.M{"name": bson.Regex{Pattern: ".*Alice.*", Options: "i"}}},
		{n: "ncontains", e: "name == ncontains(Alice)", r: bson.M{"name": bson.M{"$not": bson.M{"$regex": ".*Alice.*", "$options": "i"}}}},
		{n: "regex3", e: "name == \"Alice*\"", r: bson.M{"name": bson.Regex{Pattern: "Alice.*", Options: "i"}}},
	}
	s.testVectors(vectors)
}

func (s *ReportSuite) TestFullTextSearchQueries() {

	vectors := []queryVector{
		{n: "fts", e: "search(bob, willy, joe)", r: bson.M{"$text": bson.M{"$search": "bob willy joe"}}},
		{n: "fts-neg", e: "search(bob, willy, \"-joe\")", r: bson.M{"$text": bson.M{"$search": "bob willy -joe"}}},
	}
	s.testVectors(vectors)
}

func (s *ReportSuite) TestUserQueries() {

	vectors := []queryVector{
		{n: "nested", e: "tagArray==(\"customer:ARAMARK\" & \"_manufacturer:Chevrolet\") || tagArray==(\"customer:ARAMARK\" & \"_manufacturer:GMC\") || tagArray==(\"customer:ARAMARK\" & \"_manufacturer:Buick\") || tagArray==(\"customer:ARAMARK\" & \"_manufacturer:Cadillac\")",
			r: bson.M{"$or": []interface{}{
				bson.M{"tagArray": bson.M{"$all": []interface{}{"customer:ARAMARK", "_manufacturer:Chevrolet"}}},
				bson.M{"tagArray": bson.M{"$all": []interface{}{"customer:ARAMARK", "_manufacturer:GMC"}}},
				bson.M{"tagArray": bson.M{"$all": []interface{}{"customer:ARAMARK", "_manufacturer:Buick"}}},
				bson.M{"tagArray": bson.M{"$all": []interface{}{"customer:ARAMARK", "_manufacturer:Cadillac"}}},
			}}},
	}
	s.testVectors(vectors)
}
