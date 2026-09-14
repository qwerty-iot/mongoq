[![Go Report Card](https://goreportcard.com/badge/github.com/qwerty-iot/mongoq)](https://goreportcard.com/report/github.com/qwerty-iot/mongoq)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# mongoq

This is a simple library for converting query strings to mongo queries using Go's AST parser.

## Installation

Checkout the repository and run:

```bash
go get github.com/qwerty-iot/mongoq/v2
```

## Usage

```golang
import "github.com/qwerty-iot/mongoq/v2"

query, _ := mongoq.ParseQuery("name == Andrew && age >= 5")

fmt.Println("%v\n", query)
```

## Driver v2 migration

Version 2 requires Go 1.25 or newer and uses MongoDB Go Driver v2.9.1.
`ParseQuery` returns `go.mongodb.org/mongo-driver/v2/bson.M`, with native v2
ObjectID, Regex, and other BSON values. Query syntax is unchanged. Update the
module import to `/v2` and consume the result with driver v2; do not mix it with
driver v1 BSON values. This is a breaking Go API change, not a database migration.

The v2.0.0 reference used by Tartabit's local replacements is a planned release
version; the local migration does not publish a tag.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[MIT](https://choosealicense.com/licenses/mit/)
