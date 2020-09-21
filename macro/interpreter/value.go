package interpreter

import (
	"dxkite.cn/language/macro/ast"
)

type MacroValue interface {
	// 带参数展开
	macroValue()
	// 空定义体
	IsEmptyBody() bool
}

// 普通字符串
type MacroString string

func (m MacroString) macroValue() {}
func (m MacroString) IsEmptyBody() bool {
	return len(m) == 0
}

// 定义宏
type MacroLitValue struct {
	it   *Interpreter
	stmt *ast.ValDefineStmt
}

func (m *MacroLitValue) macroValue() {}
func (m MacroLitValue) IsEmptyBody() bool {
	return m.stmt.Body == nil
}

// 定义宏函数
type MacroFuncValue struct {
	it   *Interpreter
	stmt *ast.FuncDefineStmt
}

func (m *MacroFuncValue) macroValue() {}
func (m MacroFuncValue) IsEmptyBody() bool {
	return m.stmt.Body == nil
}
