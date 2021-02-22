package testutils

import (
	"container/list"
	"fmt"
	"go/ast"
	"go/parser"
	"reflect"
	"strconv"
	"testing"
)

type evalExprOp interface {
	Do(from reflect.Value) reflect.Value
}

type evalSelectOp struct {
	FieldName string
}

func (op *evalSelectOp) Do(from reflect.Value) reflect.Value {
	if from.Kind() == reflect.Ptr && from.IsNil() {
		return from.FieldByName(op.FieldName)
	}
	return reflect.Indirect(from).FieldByName(op.FieldName)
}

type evalIndexOp struct {
	IndexValue int
}

func (op *evalIndexOp) Do(from reflect.Value) reflect.Value {
	return from.Index(op.IndexValue)
}

func evalExpr(exprStr string, obj interface{}) (interface{}, error) {
	var expr ast.Node
	expr, err := parser.ParseExpr(exprStr)
	if err != nil {
		return nil, err
	}
	ops := []evalExprOp{}
	next := list.New()
	next.PushBack(expr)
	for {
		n := next.Back()
		if n == nil {
			break
		}
		next.Remove(n)
		switch v := n.Value.(type) {
		case *ast.SelectorExpr:
			ops = append(ops, &evalSelectOp{v.Sel.Name})
			next.PushBack(v.X)
		case *ast.IndexExpr:
			next.PushBack(v.X)
			switch idx := v.Index.(type) {
			case *ast.BasicLit:
				idxVal, err := strconv.Atoi(idx.Value)
				if err != nil {
					return nil, fmt.Errorf("could not parse numerical index: %v", err)
				}
				ops = append(ops, &evalIndexOp{idxVal})
			default:
				return nil, fmt.Errorf("unexpected index value %#v", idx)
			}
		case *ast.Ident:
			// This should only be the case for the base expression (i.e. `base.`)
		default:
			return nil, fmt.Errorf("unhandled AST type %T", v)
		}
	}

	// reverse ops
	for i, j := 0, len(ops)-1; i < j; i, j = i+1, j-1 {
		ops[i], ops[j] = ops[j], ops[i]
	}

	value := reflect.ValueOf(obj)
	for _, op := range ops {
		value = op.Do(value)
	}

	return value.Interface(), nil
}

func AssertEqNest(t *testing.T, base interface{}, expr string, expected interface{}) {
	v, err := evalExpr(expr, base)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(v, expected) {
		t.Helper()
		t.Errorf("%s: %[2]T(%[2]v) != %[3]T(%[3]v)", expr, v, expected)
	}
}

func AssertEq(t *testing.T, value, expected interface{}) {
	if value != expected {
		t.Helper()
		t.Errorf("%[1]T(%[1]v) != %[2]T(%[2]v)", value, expected)
	}
}
