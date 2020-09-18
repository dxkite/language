package interpreter

import (
	"dxkite.cn/language/macro/ast"
	"dxkite.cn/language/macro/parser"
	"dxkite.cn/language/macro/token"
	"strconv"
	"strings"
)

type ExtractEnv struct {
	pos   token.Pos                 // 外部位置
	Stack []string                  // 当前展开的宏
	Val   map[string]ast.MacroLiter // 外部参数
}

// 创建环境
func NewEnv(pos token.Pos, name string, outer map[string]ast.MacroLiter) *ExtractEnv {
	return NewExtractEnv(pos, []string{name}, outer)
}

// 创建全局环境
func NewGlobalEnv(pos token.Pos) *ExtractEnv {
	return NewExtractEnv(pos, []string{}, nil)
}

func NewExtractEnv(pos token.Pos, stack []string, outer map[string]ast.MacroLiter) *ExtractEnv {
	return &ExtractEnv{
		pos:   pos,
		Stack: stack,
		Val:   outer,
	}
}

func (e *ExtractEnv) Pos(pos token.Pos) token.Pos {
	if e.pos != token.NoPos {
		return e.pos
	}
	return pos
}

// 是否是当前展开的宏
func (e *ExtractEnv) InStack(name string) bool {
	for _, n := range e.Stack {
		if name == n {
			return true
		}
	}
	return false
}

func (e *ExtractEnv) EmptyStack() bool {
	return len(e.Stack) == 0
}

func (e *ExtractEnv) Push(name string) {
	e.Stack = append(e.Stack, name)
}

func (e *ExtractEnv) Pop() string {
	l := len(e.Stack) - 1
	s := e.Stack[l]
	e.Stack = e.Stack[:l]
	return s
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
	return len(e.Stack) == 0
}

// 展开函数
func (e *ExtractEnv) ExtractFunc(it *Interpreter, v *ast.MacroCallExpr) string {
	//fmt.Println("Extract fun", v.Name.Name, "from", e.Stack, "at", it.pos.CreatePosition(e.Pos(v.Pos())))
	//fmt.Println("call at", it.pos.CreatePosition(v.Pos()))
	if e.InStack(v.Name.Name) {
		return it.macroFuncString(v, e)
	}
	defer e.Pop()
	// 从全局调用的创建新的调用环境
	if e.EmptyStack() {
		e.Push(v.Name.Name)
		return it.extractMacroFuncWith(v, NewEnv(v.Pos(), v.Name.Name, e.Val))
	}
	e.Push(v.Name.Name)
	return it.extractMacroFuncWith(v, e)
}

// 展开宏值
func (e *ExtractEnv) ExtractVal(it *Interpreter, v *ast.Ident) string {
	if e.InStack(v.Name) {
		return v.Name
	}
	defer e.Pop()
	if e.EmptyStack() {
		e.Push(v.Name)
		return it.extractMacroValue(v, NewEnv(v.Pos(), v.Name, e.Val))
	}
	e.Push(v.Name)
	if e.HasValue(v.Name) {
		return it.extract(e.Val[v.Name], e)
	}
	return it.extractMacroValue(v, e)
}

type MacroValue interface {
	// 带参数展开
	Extract(env *ExtractEnv) string
	// 空定义体
	IsEmptyBody() bool
}

// 普通字符串
type MacroString string

func (m MacroString) Extract(*ExtractEnv) string {
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
	return m.it.extract(m.stmt.Body, env)
}

func (m *MacroLitValue) String() string {
	return m.it.litString(m.stmt.Body)
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
		for _, v := range *vv {
			t += m.parseFuncBodyItem(v, env)
		}
		return t
	default:
		m.it.errorf(v.Pos(), "unknown item %v", v)
	}
	return ""
}

// 解析参数宏
func (m *MacroFuncValue) parseMacroParam(v ast.MacroLiter, env *ExtractEnv, msg string) (string, bool) {
	if x, ok := v.(*ast.Ident); ok {
		if vv, ok := env.Val[x.Name]; ok {
			return strings.TrimSpace(m.it.macroValueItemString(vv, env)), true
		}
	}
	m.it.errorf(v.Pos(), msg)
	return "", false
}

func (m *MacroFuncValue) getBinaryParam(v ast.MacroLiter, env *ExtractEnv) (string, bool) {
	if vv, ok := v.(*ast.Ident); ok {
		if x, ok := env.Val[vv.Name]; ok {
			return m.it.litString(x), isOkBinParam(x)
		}
	}
	return m.it.litString(v), isOkBinParam(v)
}

func isOkBinParam(v ast.MacroLiter) bool {
	switch v.(type) {
	case *ast.Ident, *ast.LitExpr:
		return true
	}
	return false
}

// 解析参数宏
func (m *MacroFuncValue) extractBinary(v *ast.BinaryExpr, env *ExtractEnv) string {
	x, xok := m.getBinaryParam(v.X, env)
	y, yok := m.getBinaryParam(v.Y, env)
	if xok && yok {
		return x + y
	}
	s := x + "##" + y
	stmt, errs := parser.ParseBodyLiter([]byte(s), v.Pos())
	if len(errs) > 0 {
		m.it.errorf(v.Pos(), errs.Error())
		return m.it.extract(v.X, env) + m.it.extract(v.Y, env)
	}
	t := ""
	for _, v := range *stmt {
		if vv, ok := v.(*ast.BinaryExpr); ok {
			t += m.it.litString(vv.X)
			t += m.it.litString(vv.Y)
		} else {
			t += m.parseFuncBodyItem(v, env)
		}
	}
	return t
}
