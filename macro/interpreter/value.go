package interpreter

import (
	"dxkite.cn/language/macro/ast"
	"dxkite.cn/language/macro/parser"
	"dxkite.cn/language/macro/token"
	"strconv"
	"strings"
)

type ExtractEnv struct {
	pos           token.Pos                 // 外部位置
	Name          string                    // 外部函数
	Val           map[string]ast.MacroLiter // 外部参数
	keepDefinedId bool                      // 保留定义的ID
}

// 创建环境
func NewEnv(pos token.Pos, name string, outer map[string]ast.MacroLiter) *ExtractEnv {
	return &ExtractEnv{
		pos:  pos,
		Name: name,
		Val:  outer,
	}
}

// 创建全局环境
func NewGlobalEnv(pos token.Pos) *ExtractEnv {
	return &ExtractEnv{
		pos:  pos,
		Name: "",
		Val:  nil,
	}
}

func (e *ExtractEnv) Pos(pos token.Pos) token.Pos {
	if e.pos != token.NoPos {
		return e.pos
	}
	return pos
}

// 是否是调用的函数
func (e *ExtractEnv) IsCaller(name string) bool {
	return name == e.Name
}

// 是否是调用的函数
func (e *ExtractEnv) HasValue(name string) bool {
	if e.Val == nil {
		return false
	}
	_, ok := e.Val[name]
	return ok
}

func (e *ExtractEnv) InMacro() bool {
	return len(e.Name) == 0
}

// 展开函数
func (e *ExtractEnv) ExtractFunc(it *Interpreter, v *ast.MacroCallExpr) string {
	//fmt.Println("extract fun", v.Name.Name, "from", e.Name, "at", it.pos.CreatePosition(e.Pos(v.Pos())))
	if e.IsCaller(v.Name.Name) {
		return it.macroFuncString(v, e)
	}
	// 从全局调用的创建新的调用环境
	if e.Name == "" {
		return it.extractMacroFuncWith(v, NewEnv(e.pos, v.Name.Name, e.Val))
	}
	return it.extractMacroFuncWith(v, e)
}

// 展开宏值
func (e *ExtractEnv) ExtractVal(it *Interpreter, v *ast.Ident) string {
	if e.HasValue(v.Name) {
		return it.extractItem(e.Val[v.Name], e)
	}
	return it.extractMacroValue(e.Pos(v.Pos()), v)
}

type MacroValue interface {
	// 带参数展开
	Extract(env *ExtractEnv) string
	// 空定义体
	IsEmptyBody() bool
}

// 普通字符串
type MacroString string

func (m MacroString) Extract(env *ExtractEnv) string {
	return string(m)
}

func (m MacroString) IsEmptyBody() bool {
	return false
}

// 定义宏
type MacroLitValue struct {
	it   *Interpreter
	stmt *ast.ValDefineStmt
}

func (m *MacroLitValue) Extract(env *ExtractEnv) string {
	if m.IsEmptyBody() {
		return ""
	}
	return m.it.extractLine(m.stmt.Body, env)
}

func (m MacroLitValue) IsEmptyBody() bool {
	return m.stmt.Body == nil
}

// 定义宏函数
type MacroFuncValue struct {
	it   *Interpreter
	stmt *ast.FuncDefineStmt
}

func (m *MacroFuncValue) Extract(env *ExtractEnv) string {
	if m.IsEmptyBody() {
		return ""
	}
	return m.extractFuncBody(env)
}

func (m MacroFuncValue) IsEmptyBody() bool {
	return m.stmt.Body == nil
}

// 展开宏定义函数
func (m *MacroFuncValue) extractFuncBody(env *ExtractEnv) string {
	t := ""
	for _, v := range *m.stmt.Body {
		t += m.parseFuncBodyItem(v, env)
	}
	return t
}

// 生成调用函数体
func (m *MacroFuncValue) parseFuncBodyItem(v ast.MacroLiter, env *ExtractEnv) string {
	switch vv := v.(type) {
	case *ast.Text:
		if parser.TokenNotIn(vv.Kind, token.BACKSLASH_NEWLINE, token.BLOCK_COMMENT) {
			return vv.Text
		}
	case *ast.MacroCallExpr:
		return env.ExtractFunc(m.it, vv)
	case *ast.Ident:
		return env.ExtractVal(m.it, vv)
	case *ast.UnaryExpr:
		if v, ok := m.parseMacroParam(vv.X, env, "'#' is not followed by a macro parameter"); ok {
			return strconv.QuoteToGraphic(v)
		}
	case *ast.BinaryExpr:
		return m.extractBinary(vv, env)
	case *ast.LitExpr:
		return vv.Value
	case *ast.MacroLitArray:
		t := ""
		for _, v := range *m.stmt.Body {
			t += m.parseFuncBodyItem(v, env)
		}
		return t
	default:
		m.it.errorf(v.Pos(), "unknown item %v", v)
	}
	return ""
}

// 解析参数宏
func (m *MacroFuncValue) parseMacroParam(v ast.Expr, env *ExtractEnv, msg string) (string, bool) {
	if x, ok := v.(*ast.Ident); ok {
		if vv, ok := env.Val[x.Name]; ok {
			return strings.TrimSpace(m.it.macroValueItemString(vv, env)), true
		}
	}
	m.it.errorf(v.Pos(), msg)
	return "", false
}

// 解析参数宏
func (m *MacroFuncValue) extractBinary(v *ast.BinaryExpr, env *ExtractEnv) string {
	// TODO extract ## operator parameter
	x, _ := m.parseMacroParam(v.X, env, "'##' x must be a macro parameter")
	y, _ := m.parseMacroParam(v.Y, env, "'##' y must be a macro parameter")
	return x + y
}
