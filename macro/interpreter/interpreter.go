package interpreter

import (
	"bytes"
	"dxkite.cn/language/macro/ast"
	"dxkite.cn/language/macro/token"
	"fmt"
	"strconv"
)

// 解释器
type interpreter struct {
	// 已经定义的宏
	Val map[string]string
	// 已经定义的函数
	Func map[string]ast.FuncDefineStmt
	// 运行后的源码
	src bytes.Buffer
}

// 执行AST
func (it interpreter) Eval(node ast.Node) []byte {
	return nil
}

// 解析表达式
func (it interpreter) evalExpr(expr ast.Expr) interface{} {
	return false
}

// 二元运算
func (it interpreter) evalBinaryExpr(expr *ast.BinaryExpr) interface{} {
	x := it.evalExpr(expr.X)
	y := it.evalExpr(expr.Y)
	return evalBinary(x, y, expr.Op)
}

// 一元运算
func (it interpreter) evalUnaryExpr(expr *ast.UnaryExpr) interface{} {
	switch expr.Op {
	case token.LNOT: // !number
		v := it.evalExpr(expr.X)
		switch n := v.(type) {
		case int64:
			return !(n > 0)
		case float64:
			return !(n > 0)
		default:
			it.errorf(expr.X.Pos(), "unexpected token %v in LNOT expr", expr.X)
		}
	case token.NOT: // ~number
		v := it.evalExpr(expr.X)
		switch n := v.(type) {
		case int64:
			return ^n
		case float64:
			return ^int64(n)
		default:
			it.errorf(expr.X.Pos(), "unexpected token %v in LNOT expr", expr.X)
		}
	case token.SUB: // - number
		v := it.evalExpr(expr.X)
		switch n := v.(type) {
		case int64:
			return -n
		case float64:
			return -n
		default:
			it.errorf(expr.X.Pos(), "unexpected token %v in LNOT expr", expr.X)
		}
	case token.DEFINED: // defined ident
		id, _ := expr.X.(*ast.Ident)
		if _, ok := it.Val[id.Name]; ok {
			return int64(1)
		}
		if _, ok := it.Func[id.Name]; ok {
			return int64(1)
		}
		return int64(0)
	case token.MACRO: // #ident
		id, _ := expr.X.(*ast.Ident)
		return strconv.QuoteToGraphic(it.evalIdent(id))
	}
	return ""
}

// 解析ID标识符
func (it interpreter) evalIdent(*ast.Ident) string {
	return ""
}

// 解析字面量表达式去运算
func (it interpreter) evalLitExpr(expr *ast.LitExpr) interface{} {
	if expr.Kind == token.INT {
		return it.evalInt(expr)
	}
	if expr.Kind == token.FLOAT {
		return it.evalFloat(expr)
	}
	it.errorf(expr.Offset, " token %s is not valid in preprocessor expressions", strconv.Quote(expr.Value))
	return 0
}

// 解析数字
func (it interpreter) evalInt(expr *ast.LitExpr) int64 {
	v, err := strconv.ParseInt(expr.Value, 0, 64)
	if err != nil {
		it.errorf(expr.Offset, "error parse int %s", err.Error())
	}
	return v
}

// 解析数字（浮点数）
func (it interpreter) evalFloat(expr *ast.LitExpr) float64 {
	v, err := strconv.ParseFloat(expr.Value, 64)
	if err != nil {
		it.errorf(expr.Offset, "error parse float %s", err.Error())
	}
	return v
}

func (it interpreter) error(pos token.Pos, msg string) {
	fmt.Println("error", pos, msg)
}

func (it interpreter) errorf(pos token.Pos, format string, args ...interface{}) {
	it.error(pos, fmt.Sprintf(format, args...))
}
