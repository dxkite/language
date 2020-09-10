package interpreter

import (
	"bytes"
	"dxkite.cn/language/macro/ast"
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
func (it interpreter) evalExpr(expr ast.Expr) bool {
	return false
}
