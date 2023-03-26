package mongoq

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/qwerty-iot/tox"
)

var OnErrorCallback func(originalExpression string, err error)

func onError(originalExpression string, err error) {
	if OnErrorCallback != nil {
		OnErrorCallback(originalExpression, err)
	}
}

func ParseQuery(expr string) (bson.M, error) {
	// Parse the expression and generate an AST
	expr = strings.TrimSpace(expr)
	fset := token.NewFileSet()
	exprAst, err := parser.ParseExprFrom(fset, "", expr, 0)
	if err != nil {
		onError(expr, err)
		return nil, err
	}

	// Convert the AST to a MongoDB query
	query, err := convertExprToMongoQuery(exprAst, nil)
	if err != nil {
		onError(expr, err)
		return nil, err
	}

	m, ok := query.(bson.M)
	if !ok {
		return nil, fmt.Errorf("failed to convert to bson.M")
	}

	return bson.M(m), nil
}

func mergeArrays(leftQuery any, rightQuery any) []any {
	la, lok := leftQuery.([]any)
	ra, rok := rightQuery.([]any)
	var rslt []any
	if lok {
		rslt = la
	} else {
		rslt = []any{leftQuery}
	}
	if rok {
		rslt = append(rslt, ra...)
	} else {
		rslt = append(rslt, rightQuery)
	}
	return rslt
}

func isRegex(value any) (string, bool) {
	literal, ok := value.(string)
	if !ok {
		return "", false
	}
	if len(literal) < 2 {
		return "", false // if string length is less than 2, there's no first and last character
	}
	firstChar := literal[0]             // get the first character of the string
	lastChar := literal[len(literal)-1] // get the last character of the string
	if firstChar == '/' && lastChar == '/' {
		return literal[1 : len(literal)-1], true // remove the first and last characters and return the string
	} else {
		return "", false
	}
}

func convertBinaryOp(e *ast.BinaryExpr, parentOp *token.Token) (any, error) {
	operator := binaryOpToMongoOperator(e.Op)

	leftQuery, err := convertExprToMongoQuery(e.X, &e.Op)
	if err != nil {
		return nil, err
	}
	rightQuery, err := convertExprToMongoQuery(e.Y, &e.Op)
	if err != nil {
		return nil, err
	}

	switch operator {
	case "$eq":
		if rv, rok := isRegex(rightQuery); rok {
			return bson.M{
				tox.ToString(leftQuery): primitive.Regex{Pattern: rv, Options: "i"},
			}, nil
		}
		return bson.M{
			tox.ToString(leftQuery): rightQuery,
		}, nil
	case "$ne":
		if rm, rok := rightQuery.(bson.M); rok {
			if rin, rinf := rm["$in"]; rinf {
				return bson.M{
					tox.ToString(leftQuery): bson.M{"$nin": rin},
				}, nil
			}
		}
		return bson.M{
			tox.ToString(leftQuery): bson.M{operator: rightQuery},
		}, nil
	case "$gt", "$gte", "$lt", "$lte":
		switch rightQuery.(type) {
		case int64, float64:
		// noop
		default:
			return nil, fmt.Errorf("invalid right operand for operator '%s'", e.Op.String())
		}
		return bson.M{
			tox.ToString(leftQuery): bson.M{operator: rightQuery},
		}, nil
	case "$and":
		lm, lok := leftQuery.(bson.M)
		rm, rok := rightQuery.(bson.M)
		if lok && rok {
			useAnd := false
			for rk, _ := range rm {
				if _, found := lm[rk]; found {
					// revert to $and
					useAnd = true
					break
				}
			}
			if useAnd {
				return bson.M{
					operator: []any{leftQuery, rightQuery},
				}, nil
			} else {
				for rk, rv := range rm {
					if _, found := lm[rk]; found {
						// revert to $and
						useAnd = true
						break
					}
					lm[rk] = rv
				}
				return lm, nil
			}
		} else {
			return nil, fmt.Errorf("unsupported use of: '&&'")
		}
	case "$or":
		if parentOp != nil && *parentOp == token.LOR {
			// nested or
			return []any{leftQuery, rightQuery}, nil
		} else {
			return bson.M{
				operator: mergeArrays(leftQuery, rightQuery),
			}, nil
		}
	case "$in":
		rslt := mergeArrays(leftQuery, rightQuery)
		if parentOp != nil && *parentOp == token.OR {
			// nested or
			return rslt, nil
		} else if parentOp != nil && *parentOp == token.EQL {
			operator = "$in"
		} else if parentOp != nil && *parentOp == token.NEQ {
			operator = "$nin"
		}
		return bson.M{
			operator: rslt,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported operator: '%s'", e.Op.String())
	}
}

func convertExprToMongoQuery(expr ast.Expr, parentOp *token.Token) (any, error) {
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		// Handle binary expressions (e.g. "foo == bar")
		return convertBinaryOp(e, parentOp)
	case *ast.UnaryExpr:
		// Handle unary expressions (e.g. "!foo")
		if e.Op == token.NOT {
			query, err := convertExprToMongoQuery(e.X, &e.Op)
			if err != nil {
				return nil, err
			}
			if qs, ok := query.(string); ok {
				return bson.M{
					qs: bson.M{"$exists": false},
				}, nil
			}
			return bson.M{
				"$not": query,
			}, nil
		}
	case *ast.BasicLit:
		// Handle literal expressions (e.g. "true", "123")
		switch e.Kind {
		case token.INT:
			return tox.ToInt64(e.Value), nil
		case token.FLOAT:
			return tox.ToFloat64(e.Value), nil
		case token.STRING:
			lcv := strings.ToLower(e.Value)
			if lcv == "true" {
				return true, nil
			} else if lcv == "false" {
				return false, nil
			}
			strValue := strings.Trim(e.Value, `"`)
			if parentOp == nil || *parentOp == token.LAND {
				return bson.M{strValue: bson.M{"$exists": true}}, nil
			} else if oid, oidErr := primitive.ObjectIDFromHex(strValue); oidErr == nil {
				return oid, nil
			} else {
				return strValue, nil
			}
		default:
			return nil, fmt.Errorf("unsupported literal: %v %v", e.Kind, e)
			/*case token.:
				return true, nil
			case token.FALSE:
				return false, nil*/
		}
	case *ast.Ident:
		// Handle identifier expressions (e.g. "foo"), ie strings without quotes
		lcv := strings.ToLower(e.Name)
		if lcv == "true" {
			return true, nil
		} else if lcv == "false" {
			return false, nil
		}
		if parentOp == nil || binarOpIsLogical(*parentOp) {
			return bson.M{e.Name: bson.M{"$exists": true}}, nil
		} else if oid, oidErr := primitive.ObjectIDFromHex(e.Name); oidErr == nil {
			return oid, nil
		} else {
			return e.Name, nil
		}
	case *ast.ParenExpr:
		// Handle parenthesized expressions (e.g. "(foo == bar)")
		return convertExprToMongoQuery(e.X, nil)
	case *ast.SelectorExpr:
		// Handle selector expressions (e.g. "foo.bar")
		if id, ok := e.X.(*ast.Ident); ok {
			if parentOp == nil || binarOpIsLogical(*parentOp) {
				return bson.M{id.Name + "." + e.Sel.Name: bson.M{"$exists": true}}, nil
			} else {
				return id.Name + "." + e.Sel.Name, nil
			}
		}
	default:
		return nil, fmt.Errorf("unsupported ast: %V (%T)", e, e)
	}
	return nil, fmt.Errorf("unreachable")
}

func binarOpIsLogical(op token.Token) bool {
	switch op {
	case token.LAND, token.LOR:
		return true
	}
	return false
}

func binaryOpToMongoOperator(op token.Token) string {
	switch op {
	case token.EQL:
		return "$eq"
	case token.NEQ:
		return "$ne"
	case token.LSS:
		return "$lt"
	case token.GTR:
		return "$gt"
	case token.LEQ:
		return "$lte"
	case token.GEQ:
		return "$gte"
	case token.LAND:
		return "$and"
	case token.LOR:
		return "$or"
	case token.OR:
		return "$in"
	}
	return ""
}
