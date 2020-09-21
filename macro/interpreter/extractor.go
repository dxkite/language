package interpreter

import (
	"dxkite.cn/language/macro/ast"
	"dxkite.cn/language/macro/parser"
	"dxkite.cn/language/macro/token"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// 宏展开函数
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

// 空展开栈
func (e *ExtractEnv) EmptyStack() bool {
	return len(e.Stack) == 0
}

// 入栈
func (e *ExtractEnv) Push(name string) {
	e.Stack = append(e.Stack, name)
}

// 出栈
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

// 获取值
func (e *ExtractEnv) GetValue(name string) (v ast.MacroLiter, exist bool) {
	if e.Val == nil {
		return nil, false
	}
	v, exist = e.Val[name]
	return
}

// 宏变量：直接展开
// 已定义函数：展开形参（形参有#或##不进行宏参数的展开）=> 参数去除空白 => 展开当前宏；
// 未定义函数：作为宏展开函数名称 => 展开函数参数列表；
// 函数自调用：作为未定义函数展开；
type MacroExtractor struct {
	it *Interpreter
}

// 创建宏展开对象
func NewExtractor(it *Interpreter) *MacroExtractor {
	return &MacroExtractor{it: it}
}

// 展开宏字面量
func (e *MacroExtractor) Extract(v ast.MacroLiter, env *ExtractEnv) string {
	switch vv := v.(type) {
	case *ast.LitExpr:
		return vv.Value
	case *ast.Text:
		if parser.TokenNotIn(vv.Kind, token.BACKSLASH_NEWLINE, token.BLOCK_COMMENT) {
			return vv.Text
		}
	case *ast.MacroCallExpr:
		return e.ExtractFunc(vv, env)
	case *ast.Ident:
		return e.ExtractIdent(vv, env)
	case *ast.MacroLitArray:
		t := ""
		for _, v := range *vv {
			t += e.Extract(v, env)
		}
		return t
	case *ast.ParenExpr:
		return "(" + e.Extract(vv.X, env) + ")"
	case *ast.UnaryExpr:
		switch vv.Op {
		case token.DEFINED:
			return e.definedStr(vv)
		case token.SHARP:
			if v, ok := e.envParam(vv.X, env, "'#' is not followed by a macro parameter"); ok {
				return strconv.QuoteToGraphic(v)
			}
		}
	case *ast.BinaryExpr:
		return e.concatOp(vv, env)
	default:
		e.it.errorf(v.Pos(), "unknown expr %s", reflect.TypeOf(v))
	}
	return ""
}

// 将字面量转换成字符串
func (e *MacroExtractor) String(v ast.MacroLiter) string {
	switch vv := v.(type) {
	case *ast.LitExpr:
		return vv.Value
	case *ast.Text:
		if parser.TokenNotIn(vv.Kind, token.BACKSLASH_NEWLINE, token.BLOCK_COMMENT) {
			return vv.Text
		}
	case *ast.MacroCallExpr:
		return e.FuncRaw(vv)
	case *ast.Ident:
		return vv.Name
	case *ast.MacroLitArray:
		t := ""
		for _, v := range *vv {
			t += e.String(v)
		}
		return t
	case *ast.ParenExpr:
		return "(" + e.String(vv.X) + ")"
	default:
		e.it.errorf(v.Pos(), "unknown expr %s", reflect.TypeOf(v))
	}
	return ""
}

func (e *MacroExtractor) envParam(v ast.MacroLiter, env *ExtractEnv, msg string) (string, bool) {
	if x, ok := v.(*ast.Ident); ok {
		if vv, ok := env.Val[x.Name]; ok {
			return strings.TrimSpace(e.String(vv)), true
		}
	}
	e.it.errorf(v.Pos(), msg)
	return "", false
}

// 展开宏标识符
// 递归展开宏则返回字符串
// 不存在的宏展开则返回字符串
func (e *MacroExtractor) ExtractIdent(v *ast.Ident, env *ExtractEnv) string {
	// 已经展开过
	if env.InStack(v.Name) {
		return v.Name
	}
	defer env.Pop()
	if env.EmptyStack() {
		env.Push(v.Name)
		return e.IdentStr(v, NewEnv(v.Pos(), v.Name, env.Val))
	}
	env.Push(v.Name)
	// 如果是环境中的参数参数
	if v, ok := env.GetValue(v.Name); ok {
		return e.Extract(v, env)
	}
	return e.IdentStr(v, env)
}

// 展开函数
// 已定义函数：展开形参（形参有#或##不进行宏参数的展开）=> 参数去除空白 => 展开当前宏；
// 未定义函数：作为宏展开函数名称 => 展开函数参数列表；
// 函数自调用：作为未定义函数展开；
func (e *MacroExtractor) ExtractFunc(v *ast.MacroCallExpr, env *ExtractEnv) string {
	if f, ok := e.it.GetFunc(v.Name.Name); ok && !env.InStack(v.Name.Name) {
		// 已定义函数：展开形参（形参有#或##不进行宏参数的展开）=> 参数去除空白 => 展开当前宏；
		defer env.Pop()
		// 从全局调用的创建新的调用环境
		if env.EmptyStack() {
			env.Push(v.Name.Name)
			return e.Func(v, f, NewEnv(v.Pos(), v.Name.Name, env.Val))
		}
		env.Push(v.Name.Name)
		return e.Func(v, f, env)
	} else {
		// 函数自调用：作为未定义函数展开；
		// 未定义函数：作为宏展开函数名称 => 展开函数参数列表；
		return e.FuncRawStr(v, env)
	}
}

// 展开宏函数
func (e *MacroExtractor) Func(expr *ast.MacroCallExpr, vv *MacroFuncValue, env *ExtractEnv) string {
	if vv.IsEmptyBody() {
		return ""
	}
	if params, err := buildMacroParameter(expr, vv, env); err == nil {
		return e.Extract(vv.stmt.Body, NewExtractEnv(env.Pos(expr.Pos()), env.Stack, params))
	} else {
		e.it.error(expr.Pos(), err.Error())
	}
	return ""
}

// 构建函数参数
func buildMacroParameter(expr *ast.MacroCallExpr, def *MacroFuncValue, env *ExtractEnv) (map[string]ast.MacroLiter, error) {
	il := def.stmt.IdentList
	params := map[string]ast.MacroLiter{}
	if expr.ParamList == nil {
		return params, nil
	}
	pl := *expr.ParamList
	r := len(il)
	p := len(pl)
	if r != p {
		return nil, errors.New(fmt.Sprintf("%s expected %d params got %d", expr.Name.Name, r, p))
	}
	for i := range def.stmt.IdentList {
		params[il[i].Name] = trimSpace(bindParamIdent(pl[i], env))
	}
	return params, nil
}

// 去除两边空白
func trimSpace(v ast.MacroLiter) ast.MacroLiter {
	switch vv := v.(type) {
	case *ast.MacroLitArray:
		s, e, l := 0, 0, len(*vv)
		for i, item := range *vv {
			if v, ok := item.(*ast.Text); ok && v.IsEmpty() {
				continue
			} else {
				s = i
				break
			}
		}
		for i := l - 1; i >= 0; i-- {
			if v, ok := (*vv)[i].(*ast.Text); ok && v.IsEmpty() {
				continue
			} else {
				e = i
				break
			}
		}
		*vv = (*vv)[s : e+1]
	}
	return v
}

// 绑定当前环境参数到调用参数
func bindParamIdent(v ast.MacroLiter, env *ExtractEnv) ast.MacroLiter {
	switch vv := v.(type) {
	case *ast.Ident:
		if env.HasValue(vv.Name) {
			return env.Val[vv.Name]
		}
	case *ast.MacroLitArray:
		for i, v := range *vv {
			(*vv)[i] = bindParamIdent(v, env)
		}
	}
	return v
}

// 将函数调转换成原始字符串，只展开参数，不展开函数
func (e *MacroExtractor) FuncRawStr(expr *ast.MacroCallExpr, env *ExtractEnv) string {
	s := e.IdentStr(expr.Name, env)
	s += strings.Repeat(" ", int(expr.Lparen-expr.Name.End()))
	params := make([]string, 0)
	if expr.ParamList != nil {
		for _, item := range *expr.ParamList {
			params = append(params, e.Extract(item, env))
		}
	}
	s += "("
	s += strings.Join(params, ",")
	s += ")"
	return s
}

// 不展开函数，不展开参数，只作为字符串（用于#表达式）
func (e *MacroExtractor) FuncRaw(expr *ast.MacroCallExpr) string {
	s := expr.Name.Name
	s += strings.Repeat(" ", int(expr.Lparen-expr.Name.End()))
	params := make([]string, 0)
	if expr.ParamList != nil {
		for _, item := range *expr.ParamList {
			params = append(params, e.String(item))
		}
	}
	s += "("
	s += strings.Join(params, ",")
	s += ")"
	return s
}

// 不展开 defined (表达式处理用)
func (e *MacroExtractor) definedStr(v *ast.UnaryExpr) string {
	return "defined " + e.parseDefinedValue(v.X)
}

// 处理 defined 子句
func (e *MacroExtractor) parseDefinedValue(v ast.MacroLiter) string {
	if id, ok := v.(*ast.Ident); ok {
		return id.Name
	} else if p, ok := v.(*ast.ParenExpr); ok {
		return "(" + e.parseDefinedValue(p.X) + ")"
	} else {
		e.it.errorf(v.Pos(), "unexpected expr %s in defined", reflect.TypeOf(v))
		return ""
	}
}

// 展开宏标识符为字符串
func (e *MacroExtractor) IdentStr(id *ast.Ident, env *ExtractEnv) string {
	if v, ok := e.Ident(id, env); ok {
		return v
	}
	if id.Name == "__LINE__" {
		return strconv.Itoa(e.it.Position(env.Pos(id.Pos())).Line)
	}
	return id.Name
}

// 展开定义的标识符
// str 展开后的字符串 exist 是否在全局宏中存在
func (e *MacroExtractor) Ident(id *ast.Ident, env *ExtractEnv) (str string, exist bool) {
	if v, ok := e.it.GetValue(id.Name); ok {
		exist = true
		if v.IsEmptyBody() {
			str = ""
			return
		}
		switch vv := v.(type) {
		// 如果是宏定义函数，则返回名字
		case *MacroFuncValue:
			str = id.Name
		case *MacroLitValue:
			str = e.Extract(vv.stmt.Body, env)
		case MacroString:
			str = string(vv)
		}
		return
	}
	return "", false
}

// 参数连接
func (e *MacroExtractor) concatOp(v *ast.BinaryExpr, env *ExtractEnv) string {
	x, xok := e.getBinaryParam(v.X, env)
	y, yok := e.getBinaryParam(v.Y, env)
	if xok && yok {
		return x + y
	}
	s := x + "##" + y
	stmt, errs := parser.ParseBodyLiter([]byte(s), v.Pos())
	if len(errs) > 0 {
		e.it.errorf(v.Pos(), errs.Error())
		return e.Extract(v.X, env) + e.Extract(v.Y, env)
	}
	t := ""
	for _, v := range *stmt {
		if vv, ok := v.(*ast.BinaryExpr); ok {
			t += e.String(vv.X)
			t += e.String(vv.Y)
		} else {
			t += e.Extract(v, env)
		}
	}
	return t
}

// 获取 ## 参数
func (e *MacroExtractor) getBinaryParam(v ast.MacroLiter, env *ExtractEnv) (string, bool) {
	if vv, ok := v.(*ast.Ident); ok {
		if x, ok := env.Val[vv.Name]; ok {
			return e.String(x), isOkBinParam(x)
		}
	}
	return e.String(v), isOkBinParam(v)
}

func isOkBinParam(v ast.MacroLiter) bool {
	switch v.(type) {
	case *ast.Ident, *ast.LitExpr:
		return true
	}
	return false
}
