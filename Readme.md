[![Go Report Card](https://goreportcard.com/badge/github.com/qwerty-iot/mongoq)](https://goreportcard.com/report/github.com/qwerty-iot/mongoq)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# mongoq

This is a simple library for converting query strings to mongo queries using Go's AST parser.

## Installation

Checkout the repository and run:

```bash
go get github.com/qwerty-iot/mongoq
```

## Usage

```golang
import "github.com/qwerty-iot/mongoq"

query, _ := mongoq.ParseQuery("name == Andrew && age >= 5")

fmt.Println("%v\n", query)
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[MIT](https://choosealicense.com/licenses/mit/)
