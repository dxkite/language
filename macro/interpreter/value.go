package interpreter

import (
	"dxkite.cn/language/macro/ast"
	"dxkite.cn/language/macro/parser"
	"dxkite.cn/language/macro/token"
	"strconv"
	"strings"
)

type MacroValue interface {
	// 带参数展开
	Extract(pos token.Pos, params map[string]ast.MacroLiter) string
}

// 普通字符串
type MacroString string

func (m MacroString) Extract(token.Pos, map[string]ast.MacroLiter) string {
	return string(m)
}

// 定义宏
type MacroLitValue struct {
	it   *Interpreter
	stmt *ast.ValDefineStmt
}

func (m *MacroLitValue) Extract(pos token.Pos, outer map[string]ast.MacroLiter) string {
	if m.stmt.Body == nil {
		return m.stmt.Name.Name
	}
	return m.it.extractMacroLine(m.stmt.Body, pos, outer)
}

// 定义宏函数
type MacroFuncValue struct {
	it   *Interpreter
	stmt *ast.FuncDefineStmt
}

func (m *MacroFuncValue) Extract(pos token.Pos, outer map[string]ast.MacroLiter) string {
	return m.extractFuncBody(pos, outer)
}

// 展开宏定义函数
func (m *MacroFuncValue) extractFuncBody(pos token.Pos, outer map[string]ast.MacroLiter) string {
	t := ""
	for _, v := range *m.stmt.Body {
		t += m.parseFuncBodyItem(pos, v, outer)
	}
	return t
}

// 生成调用函数体
func (m *MacroFuncValue) parseFuncBodyItem(pos token.Pos, v ast.MacroLiter, outer map[string]ast.MacroLiter) string {
	switch vv := v.(type) {
	case *ast.Text:
		if parser.TokenNotIn(vv.Kind, token.BACKSLASH_NEWLINE, token.BLOCK_COMMENT) {
			return vv.Text
		}
	case *ast.MacroCallExpr:
		// 不允许递归调用
		// TODO 递归调用处理
		if vv.Name.Name == m.stmt.Name.Name {
			return m.it.macroFuncString(pos, vv, outer)
		}
		return m.it.extractMacroFuncWith(pos, vv, outer)
	case *ast.Ident:
		if outer != nil {
			if _, ok := outer[vv.Name]; ok {
				return m.it.extractMacroItem(outer[vv.Name], pos, outer)
			}
		}
		return m.it.extractMacroValue(pos, vv)
	case *ast.UnaryExpr:
		if v, ok := m.parseMacroParam(vv.X, pos, outer, "'#' is not followed by a macro parameter"); ok {
			return strconv.QuoteToGraphic(v)
		}
	case *ast.BinaryExpr:
		return m.extractBinary(vv, pos, outer)
	case *ast.LitExpr:
		return vv.Value
	case *ast.MacroLitArray:
		t := ""
		for _, v := range *m.stmt.Body {
			t += m.parseFuncBodyItem(pos, v, outer)
		}
		return t
	default:
		m.it.errorf(v.Pos(), "unknown item %v", v)
	}
	return ""
}

// 解析参数宏
func (m *MacroFuncValue) parseMacroParam(v ast.Expr, pos token.Pos, params map[string]ast.MacroLiter, msg string) (string, bool) {
	if x, ok := v.(*ast.Ident); ok {
		if vv, ok := params[x.Name]; ok {
			return strings.TrimSpace(m.it.macroValueItemString(vv, pos, params)), true
		}
	}
	m.it.errorf(v.Pos(), msg)
	return "", false
}

// 解析参数宏
func (m *MacroFuncValue) extractBinary(v *ast.BinaryExpr, pos token.Pos, params map[string]ast.MacroLiter) string {
	// TODO extract ## operator parameter
	x, _ := m.parseMacroParam(v.X, pos, params, "'##' x must be a macro parameter")
	y, _ := m.parseMacroParam(v.Y, pos, params, "'##' y must be a macro parameter")
	return x + y
}
