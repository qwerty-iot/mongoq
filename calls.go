package mongoq

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/bson"
)

func convertCallArgsToStringArray(name string, args []ast.Expr, expected int) ([]string, error) {
	var arr []string
	for _, arg := range args {
		switch targ := arg.(type) {
		case *ast.BasicLit:
			arr = append(arr, strings.Trim(targ.Value, `"`))
		case *ast.Ident:
			arr = append(arr, targ.Name)
		case *ast.SelectorExpr:
			// Handle selector expressions (e.g. "foo.bar")
			if id, ok := targ.X.(*ast.Ident); ok {
				arr = append(arr, id.Name+"."+targ.Sel.Name)
			}
		default:
			return nil, fmt.Errorf("%s() unsupported argument type: %v", name, arg)
		}
	}
	if len(arr) < expected {
		return nil, fmt.Errorf("%s() expected %d arguments, got %d", name, expected, len(arr))
	}
	return arr, nil
}

func callSearch(e *ast.CallExpr, parentOp *token.Token) (any, error) {
	args, err := convertCallArgsToStringArray("search", e.Args, -1)
	if err != nil {
		return nil, err
	}
	return bson.M{"$text": bson.M{"$search": strings.Join(args, " ")}}, nil
}

func callExists(e *ast.CallExpr, parentOp *token.Token) (any, error) {
	args, err := convertCallArgsToStringArray("exists", e.Args, 1)
	if err != nil {
		return nil, err
	}
	return bson.M{args[0]: bson.M{"$exists": true}}, nil
}

func callNotExists(e *ast.CallExpr, parentOp *token.Token) (any, error) {
	args, err := convertCallArgsToStringArray("nexists", e.Args, 1)
	if err != nil {
		return nil, err
	}
	return bson.M{args[0]: bson.M{"$exists": false}}, nil
}

func callContains(e *ast.CallExpr, parentOp *token.Token) (any, error) {
	args, err := convertCallArgsToStringArray("like", e.Args, 1)
	if err != nil {
		return nil, err
	}
	return primitive.Regex{Pattern: ".*" + args[0] + ".*", Options: "i"}, nil
}

func callRegex(e *ast.CallExpr, parentOp *token.Token) (any, error) {
	args, err := convertCallArgsToStringArray("regex", e.Args, 1)
	if err != nil {
		return nil, err
	}
	pattern := strings.Replace(args[0], "\\\\", "\\", -1)
	return primitive.Regex{Pattern: pattern, Options: "i"}, nil
}

func callDateRelative(e *ast.CallExpr, parentOp *token.Token) (any, error) {
	args, err := convertCallArgsToStringArray("dateRelative", e.Args, 1)
	if err != nil {
		return nil, err
	}
	dur, err := time.ParseDuration(args[0])
	if err != nil {
		return nil, err
	}
	ts := time.Now().UTC().Add(dur)
	return ts, nil
}

func callDate(e *ast.CallExpr, parentOp *token.Token) (any, error) {
	args, err := convertCallArgsToStringArray("date", e.Args, 1)
	if err != nil {
		return nil, err
	}
	if len(args) == 1 {
		ts, err := time.Parse(time.RFC3339, args[0])
		if err != nil {
			return nil, err
		}
		return ts, nil
	} else {
		ts, err := time.Parse(args[1], args[0])
		if err != nil {
			return nil, err
		}
		return ts, nil
	}
}
