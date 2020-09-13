package interpreter

import (
	"bytes"
	"dxkite.cn/language/macro/ast"
	"dxkite.cn/language/macro/parser"
	"dxkite.cn/language/macro/token"
	"fmt"
	"strconv"
	"strings"
)

// 解释器
type interpreter struct {
	// 已经定义的宏
	Val map[string]string
	// 已经定义的函数
	Func map[string]*ast.FuncDefineStmt
	// 位置信息
	pos token.FilePos
	// 运行后的源码
	src *bytes.Buffer
}

// 执行AST
func (it *interpreter) Eval(node ast.Node, name string, pos token.FilePos) []byte {
	it.Val = map[string]string{
		"__FILE__": strconv.QuoteToGraphic(name),
	}
	it.Func = map[string]*ast.FuncDefineStmt{}
	it.src = &bytes.Buffer{}
	it.pos = pos
	it.evalStmt(node)
	return it.src.Bytes()
}

// 执行宏语句
func (it *interpreter) evalStmt(node ast.Node) {
	switch n := node.(type) {
	case *ast.BlockStmt:
		for _, sub := range *n {
			it.evalStmt(sub)
		}
	case *ast.MacroLitArray:
		it.src.WriteString(it.extractLitArray(n, it.Val))
	case *ast.Ident:
		it.src.WriteString(it.evalIdent(n))
	case *ast.ValDefineStmt:
		it.evalDefineVal(n)
	case *ast.FuncDefineStmt:
		it.evalDefineFunc(n)
	}
}

// 解析表达式
// 返回表达式最后的值(数值/标识符/字符串)
// 解析宏/宏函数 得到展开的宏
// 把展开的宏作为表达式解析并返回表达式的值
func (it interpreter) evalExpr(expr interface{}) interface{} {
	switch xx := expr.(type) {
	case uint8, int32, float64, *ast.Ident:
		return xx
	case ast.Expr:
		switch xx := xx.(type) {
		case *ast.ParenExpr:
			return it.evalExpr(xx.X)
		case *ast.UnaryExpr:
			return it.evalUnaryExpr(xx)
		case *ast.BinaryExpr:
			return it.evalBinaryExpr(xx)
		case *ast.MacroCallExpr:
			return it.evalCall(xx)
		}
		it.errorf(xx.Pos(), "unexpected token")
	case string:
		exp, errs := parser.ParseExpr([]byte(xx))
		if len(errs) > 0 {
			it.errorf(0, "error expr %s", xx)
		}
		return it.evalExpr(exp)
	}
	return nil
}

// 定义值
func (it *interpreter) evalDefineVal(stmt *ast.ValDefineStmt) {
	n := stmt.Name.Name
	if _, ok := it.Val[n]; ok || it.isInnerDefine(n) {
		it.errorf(stmt.Pos(), "warning: %s redefined", n)
	}
	it.Val[n] = it.extractLitArray(stmt.Body, it.Val)
	f := it.pos.CreatePosition(stmt.From).Line
	t := it.pos.CreatePosition(stmt.To).Line
	it.src.WriteString(strings.Repeat("\n", t-f+1))
}

// 定义函数
func (it *interpreter) evalDefineFunc(stmt *ast.FuncDefineStmt) {
	n := stmt.Name.Name
	if _, ok := it.Val[n]; ok || it.isInnerDefine(n) {
		it.errorf(stmt.Pos(), "warning: %s redefined", n)
	}
	it.Func[n] = stmt
	f := it.pos.CreatePosition(stmt.From).Line
	t := it.pos.CreatePosition(stmt.To).Line
	it.src.WriteString(strings.Repeat("\n", t-f+1))
}

// 内置定义
func (it interpreter) isInnerDefine(name string) bool {
	ar := []string{"__LINE__", "__FILE__", "__FUNCTION__"}
	for _, n := range ar {
		if n == name {
			return true
		}
	}
	return false
}

// 展开列表
func (it interpreter) parseLitArray(array *ast.MacroLitArray, extra map[string]string) string {
	t := ""
	for _, v := range *array {
		t += it.parseLitItem(v, extra)
	}
	return t
}

// 解析参数调用
func (it interpreter) parseLitItem(v ast.Expr, params map[string]string) string {
	switch vv := v.(type) {
	case *ast.Text:
		if parser.TokenNotIn(vv.Kind, token.BACKSLASH_NEWLINE, token.BLOCK_COMMENT) {
			return vv.Text
		}
	case *ast.MacroCallExpr:
		return it.parseCallToString(vv, params)
	case *ast.Ident:
		return vv.Name
	case *ast.MacroLitArray:
		return it.parseLitArray(vv, params)
	default:
		it.errorf(v.Pos(), "unknown expr %v", v)
	}
	return ""
}

// 展开列表
func (it interpreter) extractLitArray(array *ast.MacroLitArray, extra map[string]string) string {
	t := ""
	for _, v := range *array {
		t += it.extractLitItem(v, extra)
	}
	return t
}

// 展开参数调用
func (it interpreter) extractLitItem(v ast.Expr, params map[string]string) string {
	switch vv := v.(type) {
	case *ast.Text:
		if parser.TokenNotIn(vv.Kind, token.BACKSLASH_NEWLINE, token.BLOCK_COMMENT) {
			return vv.Text
		}
	case *ast.MacroCallExpr:
		return it.evalCallInner(vv, params)
	case *ast.Ident:
		if params != nil {
			if v, ok := params[vv.Name]; ok {
				return v
			}
		}
		return it.evalIdent(vv)
	case *ast.MacroLitArray:
		return it.extractLitArray(vv, params)
	default:
		it.errorf(v.Pos(), "unknown expr %v", v)
	}
	return ""
}

// 执行函数
func (it interpreter) evalCall(expr *ast.MacroCallExpr) string {
	return it.evalCallInner(expr, nil)
}

// 执行函数表达式
func (it interpreter) evalCallInner(expr *ast.MacroCallExpr, extra map[string]string) string {
	if fun, ok := it.Func[expr.Name.Name]; ok {
		r := len(fun.IdentList)
		p := len(*expr.ParamList)
		if r != p {
			it.errorf(expr.Pos(), "expected params %d got %d", r, p)
			goto exit
		}
		params := map[string]ast.MacroLiter{}
		for i := range fun.IdentList {
			params[fun.IdentList[i].Name] = (*expr.ParamList)[i]
		}
		return it.parseFuncBody(fun.Body, params, extra)
	}
exit:
	return it.parseCallToString(expr, extra)
}

// 解析列表
func (it interpreter) parseFuncBody(array *ast.MacroLitArray, params map[string]ast.MacroLiter, extra map[string]string) string {
	t := ""
	for _, v := range *array {
		t += it.parseFuncBodyItem(v, params, extra)
	}
	return t
}

// 生成调用函数体
func (it interpreter) parseFuncBodyItem(v ast.Expr, params map[string]ast.MacroLiter, extra map[string]string) string {
	switch vv := v.(type) {
	case *ast.Text:
		if parser.TokenNotIn(vv.Kind, token.BACKSLASH_NEWLINE, token.BLOCK_COMMENT) {
			return vv.Text
		}
	case *ast.MacroCallExpr:
		return it.evalCallInner(buildCallExpr(vv, it.extractCallParamList(params, extra)), nil)
	case *ast.Ident:
		if v, ok := params[vv.Name]; ok {
			return it.extractCallParamItem(v, extra)
		}
		return it.evalIdent(vv)
	case *ast.UnaryExpr:
		if v, ok := it.parseMacroParam(vv.X, params, "'#' is not followed by a macro parameter"); ok {
			return strconv.QuoteToGraphic(v)
		}
	case *ast.BinaryExpr:
		x, _ := it.parseMacroParam(vv.X, params, "'##' x must be a macro parameter")
		y, _ := it.parseMacroParam(vv.Y, params, "'##' y must be a macro parameter")
		return x + y
	}
	return ""
}

// 展开参数宏
func (it interpreter) extractMacroParam(v ast.Expr, params map[string]ast.MacroLiter, extra map[string]string, msg string) (string, bool) {
	if x, ok := v.(*ast.Ident); ok {
		if v, ok := params[x.Name]; ok {
			return strings.TrimSpace(it.extractLitItem(v, extra)), true
		}
	}
	it.errorf(v.Pos(), msg)
	return "", false
}

// 解析参数宏
func (it interpreter) parseMacroParam(v ast.Expr, params map[string]ast.MacroLiter, msg string) (string, bool) {
	if x, ok := v.(*ast.Ident); ok {
		if vv, ok := params[x.Name]; ok {
			return strings.TrimSpace(it.parseLitItem(vv, nil)), true
		}
	}
	it.errorf(v.Pos(), msg)
	return "", false
}

// 未定义函数
func (it interpreter) parseCallToString(expr *ast.MacroCallExpr, extra map[string]string) string {
	s := expr.Name.Name
	s += strings.Repeat(" ", int(expr.LParen-expr.Name.End()))
	s += "("
	s += strings.Join(it.parseCallParamList(*expr.ParamList, extra), ",")
	s += ")"
	return s
}

// 应用函数参数
func (it interpreter) parseCallParamList(list []ast.MacroLiter, extra map[string]string) []string {
	params := []string{}
	for _, item := range list {
		params = append(params, it.extractCallParamItem(item, extra))
	}
	return params
}

// 展开函数参数
func (it interpreter) extractCallParamList(list map[string]ast.MacroLiter, extra map[string]string) map[string]string {
	params := map[string]string{}
	for name, item := range list {
		params[name] = strings.TrimSpace(it.extractCallParamItem(item, extra))
	}
	return params
}

// 创建新调用
func buildCallExpr(c *ast.MacroCallExpr, params map[string]string) *ast.MacroCallExpr {
	cc := &ast.MacroCallExpr{
		From: c.From,
		To:   c.To,
		Name: &ast.Ident{
			Offset: c.Name.Offset,
			Name:   c.Name.Name,
		},
		LParen:    c.LParen,
		ParamList: &ast.MacroLitArray{},
		RParen:    c.RParen,
	}
	*cc.ParamList = append(*cc.ParamList, *c.ParamList...)
	cc.ParamList = bindParamList(cc.ParamList, c.ParamList, params)
	return cc
}

// 绑定当前环境参数到调用参数
func bindParamList(list, from *ast.MacroLitArray, params map[string]string) *ast.MacroLitArray {
	for i, v := range *from {
		if vv, ok := v.(*ast.Ident); ok {
			if x, ok := params[vv.Name]; ok {
				(*list)[i] = &ast.Text{
					Offset: vv.Offset,
					Kind:   token.TEXT,
					Text:   x,
				}
			}
		} else if vv, ok := v.(*ast.MacroLitArray); ok {
			(*list)[i] = bindParamList(vv, from, params)
		}
	}
	return list
}

// 解开函数参数
func (it interpreter) extractCallParamItem(item ast.MacroLiter, extra map[string]string) string {
	switch t := item.(type) {
	case *ast.MacroLitArray:
		return it.extractLitArray(t, extra)
	default:
		return it.extractLitItem(t, extra)
	}
}

// 二元运算
func (it interpreter) evalBinaryExpr(expr *ast.BinaryExpr) interface{} {
	return uint8(0)
}

// 表达式到整数
func (it interpreter) intExpr(value interface{}) int32 {
	switch ex := value.(type) {
	case *ast.LitExpr:
		if ex.Kind == token.INT {
			return it.evalInt(ex)
		}
		if ex.Kind == token.CHAR {
			return int32(charValue(ex.Value))
		}
		if ex.Kind == token.FLOAT {
			return int32(it.evalFloat(ex))
		}
	case float64:
		return int32(ex)
	case uint8:
		return int32(ex)
	case int32:
		return ex
	}
	return 0
}

// 表达式到浮点数
func (it interpreter) floatExpr(value interface{}) float64 {
	switch ex := value.(type) {
	case *ast.LitExpr:
		if ex.Kind == token.INT {
			return it.evalFloat(ex)
		}
		if ex.Kind == token.INT {
			return float64(it.evalInt(ex))
		}
		if ex.Kind == token.CHAR {
			return float64(charValue(ex.Value))
		}
	case float64:
		return ex
	case uint8:
		return float64(ex)
	case int32:
		return float64(ex)
	}
	return 0
}

// 特殊转义
var chMap = map[uint8]uint8{
	'a':  7,
	'b':  8,
	'f':  12,
	'n':  10,
	'r':  13,
	't':  9,
	'v':  11,
	'\\': 92,
	'\'': 39,
	'"':  34,
}

// 字符串值
func charValue(ch string) uint8 {
	v, _ := tryCharValue(ch)
	return v
}

func tryCharValue(ch string) (uint8, bool) {
	l := len(ch)
	if l <= 2 {
		return 0, false
	}
	if ch[0] != '\'' && ch[l-1] != '\'' {
		return 0, false
	}
	ch = ch[1 : l-1] // ''
	// x
	if l == 1 {
		return ch[0], true
	} else if ch[0] == '\\' {
		// \000
		// \xff
		i := 1
		base := 8
		switch ch[1] {
		case 'x':
			i = 2
			base = 16
		case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', '"', '\'':
			return chMap[ch[1]], true
		}
		x := 0
		for i < len(ch) {
			x = x*base + digitVal(rune(ch[i]))
			i++
		}
		return uint8(x), true
	}
	return 0, false
}

func lower(ch rune) rune { return ('a' - 'A') | ch }
func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= lower(ch) && lower(ch) <= 'f':
		return int(lower(ch) - 'a' + 10)
	}
	return 16
}

// 一元运算
func (it interpreter) evalUnaryExpr(expr *ast.UnaryExpr) interface{} {
	switch expr.Op {
	case token.NOT: // ~
		v := it.expectedValue(it.evalExpr(expr))
		if vv, ok := v.(uint8); ok {
			return ^vv
		}
		if vv, ok := v.(int32); ok {
			return ^vv
		}
	case token.LNOT: // !
		v := it.expectedValue(it.evalExpr(expr))
		if vv, ok := v.(uint8); ok {
			return !(vv > 0)
		}
		if vv, ok := v.(int32); ok {
			return !(vv > 0)
		}
		if vv, ok := v.(float64); ok {
			return !(vv > 0)
		}
	case token.SUB: // -
		v := it.expectedValue(it.evalExpr(expr))
		if vv, ok := v.(uint8); ok {
			return -vv
		}
		if vv, ok := v.(int32); ok {
			return -vv
		}
		if vv, ok := v.(float64); ok {
			return -vv
		}
	case token.DEFINED: // defined ident
		ex := it.evalExpr(expr.X)
		if id, ok := ex.(*ast.Ident); ok {
			if _, ok := it.Val[id.Name]; ok {
				return uint8(1)
			}
			if _, ok := it.Func[id.Name]; ok {
				return uint8(1)
			}
		} else {
			it.errorf(expr.X.Pos(), "'defined' is not followed by a ident")
			return uint8(0)
		}
	}
	it.errorf(expr.X.Pos(), "unexpected value %v in unary expr", expr)
	return uint8(0)
}

// 解析ID标识符
func (it interpreter) evalIdent(id *ast.Ident) string {
	if v, ok := it.Val[id.Name]; ok {
		return v
	}
	if id.Name == "__LINE__" {
		return strconv.Itoa(it.pos.CreatePosition(id.Pos()).Line)
	}
	return id.Name
}

// 转换表达式到值
func (it interpreter) expectedValue(value interface{}) interface{} {
	switch v := value.(type) {
	case *ast.LitExpr:
		if v.Kind == token.FLOAT {
			return it.evalFloat(v)
		}
		if v.Kind == token.INT {
			return it.evalInt(v)
		}
		if v.Kind == token.CHAR {
			return charValue(v.Value)
		}
	case float64, uint8, int32:
		return v
	case *ast.Ident:
		vv := it.evalIdent(v)
		if v, err := strconv.ParseInt(vv, 0, 32); err == nil {
			return v
		}
		if v, err := strconv.ParseFloat(vv, 64); err == nil {
			return v
		}
		if v, ok := tryCharValue(vv); ok {
			return v
		}
		it.errorf(v.Pos(), "unexpected ident value %v", v)
	case ast.Expr:
		it.errorf(v.Pos(), "unexpected value %v", v)
	}
	return uint8(0)
}

// 解析数字
func (it interpreter) evalInt(expr *ast.LitExpr) int32 {
	v, err := strconv.ParseInt(expr.Value, 0, 32)
	if err != nil {
		it.errorf(expr.Offset, "error parse int %s", err.Error())
	}
	return int32(v)
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
