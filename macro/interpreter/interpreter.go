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
	// ast.ValDefineStmt
	// ast.FuncDefineStmt
	// string
	Val map[string]interface{}
	// 位置信息
	pos token.FilePos
	// 运行后的源码
	src *bytes.Buffer
}

// 执行ast
func (it *interpreter) Eval(node ast.Node, name string, pos token.FilePos) []byte {
	it.Val = map[string]interface{}{}
	it.Val["__FILE__"] = strconv.QuoteToGraphic(name)
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
		it.src.WriteString(it.extractMacroLine(n, -1))
	case *ast.Ident:
		it.src.WriteString(it.extractIdent(n, -1))
	case *ast.ValDefineStmt:
		it.evalDefineVal(n)
	case *ast.FuncDefineStmt:
		it.evalDefineFunc(n)
	case *ast.IfStmt:
		it.evalIf(n)
	case *ast.ElseIfStmt:
		it.evalElseIf(n)
	case *ast.UnDefineStmt:
		it.evalUnDefineStmt(n)
	}
}

// 定义一个宏
func (it *interpreter) evalDefineVal(stmt *ast.ValDefineStmt) {
	n := stmt.Name.Name
	if _, ok := it.Val[n]; ok || it.isInnerDefine(n) {
		it.errorf(stmt.Pos(), "warning: %s redefined", n)
	}
	it.Val[n] = stmt
	it.writePlaceholder(stmt)
}

// 定义一个宏函数
func (it *interpreter) evalDefineFunc(stmt *ast.FuncDefineStmt) {
	n := stmt.Name.Name
	if _, ok := it.Val[n]; ok || it.isInnerDefine(n) {
		it.errorf(stmt.Pos(), "warning: %s redefined", n)
	}
	it.Val[n] = stmt
	it.writePlaceholder(stmt)
}

// 取消定义
func (it *interpreter) evalUnDefineStmt(stmt *ast.UnDefineStmt) {
	n := stmt.Name.Name
	if _, ok := it.Val[n]; ok {
		delete(it.Val, n)
	}
	it.writePlaceholder(stmt)
}

// 检测内置预定义变量
func (it interpreter) isInnerDefine(name string) bool {
	ar := []string{"__LINE__", "__FUNCTION__"}
	for _, n := range ar {
		if n == name {
			return true
		}
	}
	return false
}

// 展开宏列表
func (it interpreter) extractLitArray(array *ast.MacroLitArray, extra map[string]string) string {
	t := ""
	for _, v := range *array {
		t += it.extractLitItem(v, extra)
	}
	return t
}

// 展开宏
func (it interpreter) extractLitItem(v ast.Expr, params map[string]string) string {
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
		return it.extractLitArray(vv, params)
	case *ast.LitExpr:
		return vv.Value
	default:
		it.errorf(v.Pos(), "unknown expr %v", v)
	}
	return ""
}

// 展开宏定义行
// extra 额外的参数
func (it interpreter) extractMacroLine(stmt interface{}, pos token.Pos) string {
	t := ""
	switch v := stmt.(type) {
	case string:
		return v
	case *ast.MacroLitArray:
		for _, v := range *v {
			if pos >= 0 {
				t += it.extractMacroItem(v, pos)
			} else {
				t += it.extractMacroItem(v, v.Pos())
			}
		}
	case *ast.BlockStmt:
		for _, v := range *v {
			t += it.extractMacroLine(v, pos)
		}
	}
	return t
}

// 展开参数调用
// params 调用时的参数
func (it interpreter) extractMacroItem(v interface{}, pos token.Pos) string {
	switch vv := v.(type) {
	case *ast.Text:
		if parser.TokenNotIn(vv.Kind, token.BACKSLASH_NEWLINE, token.BLOCK_COMMENT) {
			return vv.Text
		}
	case *ast.MacroCallExpr:
		return it.evalCallInner(vv, nil)
	case *ast.Ident:
		if pos >= 0 {
			fmt.Println("extract ident", vv.Name, it.pos.CreatePosition(pos))
		} else {
			fmt.Println("extract ident", vv.Name, it.pos.CreatePosition(vv.Pos()))
		}
		return it.extractIdent(vv, pos)
	case *ast.MacroLitArray:
		return it.extractMacroLine(vv, pos)
	default:
		it.errorf(pos, "unknown expr %v", v)
	}
	return ""
}

// 展开ID标识符
// id 展开的标识符（含位置）
// pos > 0 则表示在标识符内部展开
func (it interpreter) extractIdent(id *ast.Ident, pos token.Pos) string {
	if v, ok := it.Val[id.Name]; ok {
		// 如果是宏定义函数，则返回名字
		if _, ok := v.(*ast.FuncDefineStmt); ok {
			return id.Name
		}
		if vv, ok := v.(*ast.ValDefineStmt); ok {
			if pos >= 0 {
				return it.extractMacroLine(vv.Body, pos)
			}
			return it.extractMacroLine(vv.Body, id.Pos())
		}
		if vv, ok := v.(string); ok {
			return vv
		}
	}
	if id.Name == "__LINE__" {
		if pos >= 0 {
			return strconv.Itoa(it.pos.CreatePosition(pos).Line)
		}
		return strconv.Itoa(it.pos.CreatePosition(id.Pos()).Line)
	}
	return id.Name
}

// 执行函数
func (it interpreter) evalCall(expr *ast.MacroCallExpr) string {
	return it.evalCallInner(expr, nil)
}

// 执行函数表达式
func (it interpreter) evalCallInner(expr *ast.MacroCallExpr, extra map[string]string) string {
	if expr.Name.Name == "defined" {
		goto exit
	}
	if v, ok := it.Val[expr.Name.Name]; ok {
		fun, ok := v.(*ast.FuncDefineStmt)
		if !ok {
			goto exit
		}
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
		return it.extractFuncBody(expr, fun.Body, params, extra)
	}
exit:
	return it.parseCallToString(expr, extra)
}

// 展开宏定义函数
func (it interpreter) extractFuncBody(expr *ast.MacroCallExpr, array *ast.MacroLitArray, params map[string]ast.MacroLiter, extra map[string]string) string {
	t := ""
	for _, v := range *array {
		t += it.parseFuncBodyItem(expr, v, params, extra)
	}
	return t
}

// 生成调用函数体
func (it interpreter) parseFuncBodyItem(expr *ast.MacroCallExpr, v ast.Expr, params map[string]ast.MacroLiter, env map[string]string) string {
	switch vv := v.(type) {
	case *ast.Text:
		if parser.TokenNotIn(vv.Kind, token.BACKSLASH_NEWLINE, token.BLOCK_COMMENT) {
			return vv.Text
		}
	case *ast.MacroCallExpr:
		// 不允许递归调用
		if vv.Name.Name == expr.Name.Name {
			return it.parseCallToString(vv, it.extractCallParamList(expr, params, env))
		}
		return it.evalCallInner(buildCallExpr(vv, it.extractCallParamList(expr, params, env)), nil)
	case *ast.Ident:
		// 展开宏
		if v, ok := params[vv.Name]; ok {
			return it.extractCallParamItem(expr, v, env)
		}
		return it.extractIdentPos(expr.Pos(), vv)
	case *ast.UnaryExpr:
		if v, ok := it.parseMacroParam(vv.X, params, "'#' is not followed by a macro parameter"); ok {
			return strconv.QuoteToGraphic(v)
		}
	case *ast.BinaryExpr:
		x, _ := it.parseMacroParam(vv.X, params, "'##' x must be a macro parameter")
		y, _ := it.parseMacroParam(vv.Y, params, "'##' y must be a macro parameter")
		return x + y
	case *ast.LitExpr:
		return vv.Value
	default:
		it.errorf(v.Pos(), "unknown item %v", v)
	}
	return ""
}

// 解析参数宏
func (it interpreter) parseMacroParam(v ast.Expr, params map[string]ast.MacroLiter, msg string) (string, bool) {
	if x, ok := v.(*ast.Ident); ok {
		if vv, ok := params[x.Name]; ok {
			return strings.TrimSpace(it.extractLitItem(vv, nil)), true
		}
	}
	it.errorf(v.Pos(), msg)
	return "", false
}

// 未定义函数
func (it interpreter) parseCallToString(expr *ast.MacroCallExpr, extra map[string]string) string {
	s := it.extractIdentPos(expr.Pos(), expr.Name)
	s += strings.Repeat(" ", int(expr.LParen-expr.Name.End()))
	s += "("
	if expr.Name.Name != "defined" {
		s += strings.Join(it.parseCallParamList(expr, *expr.ParamList, extra), ",")
	} else {
		s += it.extractLitArray(expr.ParamList, nil)
	}
	s += ")"
	return s
}

// 应用函数参数
func (it interpreter) parseCallParamList(expr *ast.MacroCallExpr, list []ast.MacroLiter, extra map[string]string) []string {
	params := []string{}
	for _, item := range list {
		params = append(params, it.extractCallParamItem(expr, item, extra))
	}
	return params
}

// 展开函数参数
func (it interpreter) extractCallParamList(expr *ast.MacroCallExpr, list map[string]ast.MacroLiter, extra map[string]string) map[string]string {
	params := map[string]string{}
	for name, item := range list {
		params[name] = strings.TrimSpace(it.extractCallParamItem(expr, item, extra))
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
	*cc.ParamList = make([]ast.MacroLiter, len(*c.ParamList))
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
func (it interpreter) extractCallParamItem(expr *ast.MacroCallExpr, item ast.MacroLiter, env map[string]string) string {
	switch t := item.(type) {
	case *ast.MacroLitArray:
		return it.extractMacroLine(t, t.Pos())
	default:
		return it.extractMacroItem(t, t.Pos())
	}
}

// #if
func (it *interpreter) evalIf(stmt *ast.IfStmt) {
	v := it.evalIfBoolExpr(stmt.X)
	// #if
	it.writePlaceholder(stmt.X)
	it.evalCondition(v, stmt.Then, stmt.Else)
	it.src.WriteString("\n") // #endif
}

// #elif
func (it *interpreter) evalElseIf(stmt *ast.ElseIfStmt) {
	v := it.evalIfBoolExpr(stmt.X)
	it.evalCondition(v, stmt.Then, stmt.Else)
}

func (it *interpreter) evalIfBoolExpr(expr ast.Expr) bool {
	t, _ := expr.(*ast.Text)
	e, err := parser.ParseSubTextStmt([]byte(t.Text), t.Pos())
	if len(err) > 0 {
		it.errorf(expr.Pos(), "unexpected if expr %v", t)
	}
	ee := it.extractMacroLine(e, e.Pos())
	return it.evalExpr(ee, e.Pos())
}

func (it *interpreter) evalCondition(v bool, ts, fs ast.Stmt) {
	if v {
		it.evalStmt(ts)
		if fs != nil {
			it.writePlaceholder(fs)
		}
	} else if fs != nil {
		it.writePlaceholder(ts)
		it.evalStmt(fs)
	}
}

// 写入占位
func (it *interpreter) writePlaceholder(node ast.Node) {
	f := it.pos.CreatePosition(node.Pos()).Line
	t := it.pos.CreatePosition(node.End()).Line
	it.src.WriteString(strings.Repeat("\n", t-f+1))
}

// 二元运算
func (it interpreter) evalBinaryExpr(expr *ast.BinaryExpr) interface{} {
	x := it.evalValue(expr.X)
	y := it.evalValue(expr.Y)
	tx := typeOf(x)
	ty := typeOf(y)
	t := maxNumberType(tx, ty)
	switch expr.Op {
	case token.ADD, token.SUB, token.MUL, token.QUO,
		token.GEQ, token.GTR, token.LSS, token.LEQ, token.EQL, token.NEQ,
		token.LAND, token.LOR:
		return it.evalOpCast(x, y, t, expr.Op)
	case token.SHR, token.SHL, token.REM,
		token.AND, token.OR:
		return it.evalOpInt(x, y, expr.X, expr.Y, tx, ty, expr.Op)
	default:
		it.errorf(expr.Pos(), "unknown operator %s", expr.Op)
	}
	return uint8(0)
}

func maxNumberType(tx, ty token.Token) token.Token {
	if tx == token.FLOAT || ty == token.FLOAT {
		return token.FLOAT
	}
	if tx == token.CHAR || ty == token.CHAR {
		return token.CHAR
	}
	return token.INT
}

func typeOf(x interface{}) token.Token {
	_, xf := x.(float64)
	if xf {
		return token.FLOAT
	}
	_, ux := x.(uint8)
	if ux {
		return token.CHAR
	}
	return token.INT
}

// 运行表达式
func (it interpreter) evalExpr(expr string, pos token.Pos) bool {
	exp, errs := parser.ParseExpr([]byte(expr), pos)
	if len(errs) > 0 {
		it.errorf(pos, "error parse expr %s", expr)
	}
	switch v := it.evalValue(exp).(type) {
	case bool:
		return v
	case uint8:
		return v > 0
	case int32:
		return v > 0
	case float64:
		return v > 0
	default:
		return false
	}
}

// 解析表达式
// 返回表达式最后的值(数值/标识符/字符串)
// 解析宏/宏函数 得到展开的宏
// 把展开的宏作为表达式解析并返回表达式的值
func (it interpreter) evalValue(expr interface{}) interface{} {
	switch xx := expr.(type) {
	case uint8, int32, float64:
		return xx
	case *ast.Ident:
		return it.evalIdent(xx)
	case *ast.LitExpr:
		switch xx.Kind {
		case token.CHAR:
			return it.expectedChar(xx)
		case token.INT:
			return it.intValue(xx)
		case token.FLOAT:
			return it.floatValue(xx)
		}
	case ast.Expr:
		switch xx := xx.(type) {
		case *ast.ParenExpr:
			return it.evalValue(xx.X)
		case *ast.UnaryExpr:
			return it.evalUnaryExpr(xx)
		case *ast.BinaryExpr:
			return it.evalBinaryExpr(xx)
		case *ast.MacroCallExpr:
			return it.evalCall(xx)
		}
		it.errorf(xx.Pos(), "unexpected token %v", xx)
	}
	return nil
}

// 解析Ident
func (it interpreter) expectedIdent(expr ast.Expr) *ast.Ident {
	switch xx := expr.(type) {
	case *ast.Ident:
		return xx
	case *ast.ParenExpr:
		return it.expectedIdent(xx.X)
	}
	return nil
}

func (it interpreter) evalOpInt(x, y interface{}, ex, ey ast.Expr, tx, ty, op token.Token) interface{} {
	var yy int32
	if _, ok := y.(float64); ty == token.FLOAT && ok {
		yy = it.intValue(y)
		it.errorf(ey.Pos(), "float in y expression")
	} else {
		yy = it.intValue(y)
	}
	if tx == token.FLOAT {
		xx := it.intValue(x)
		it.errorf(ex.Pos(), "float in x expression")
		switch op {
		case token.SHL:
			return xx << yy
		case token.SHR:
			return xx >> yy
		case token.REM:
			return xx % yy
		case token.AND:
			return xx & yy
		case token.OR:
			return xx | yy
		}
	}
	if tx == token.INT {
		xx := it.intValue(x)
		switch op {
		case token.SHL:
			return xx << yy
		case token.SHR:
			return xx >> yy
		case token.REM:
			return xx % yy
		case token.AND:
			return xx & yy
		case token.OR:
			return xx | yy
		}
	}
	if tx == token.CHAR {
		xx := x.(uint8)
		yy := uint8(yy)
		switch op {
		case token.SHL:
			return xx << yy
		case token.SHR:
			return xx >> yy
		case token.REM:
			return xx % yy
		case token.AND:
			return xx & yy
		case token.OR:
			return xx | yy
		}
	}
	return uint8(0)
}

// 自动转换成数据量高的类型进行运算
func (it interpreter) evalOpCast(x, y interface{}, t, op token.Token) interface{} {
	if t == token.FLOAT {
		xx, yy := it.floatValue(x), it.floatValue(y)
		switch op {
		case token.ADD:
			return xx + yy
		case token.SUB:
			return xx - yy
		case token.MUL:
			return xx * yy
		case token.QUO:
			return xx / yy
		case token.GTR:
			return xx > yy
		case token.GEQ:
			return xx >= yy
		case token.LSS:
			return xx < yy
		case token.LEQ:
			return xx <= yy
		case token.EQL:
			return xx == yy
		case token.NEQ:
			return xx != yy
		case token.LAND:
			return (xx > 0) && (yy > 0)
		case token.LOR:
			return (xx > 0) && (yy > 0)
		}
	}
	if t == token.INT {
		xx, yy := it.intValue(x), it.intValue(y)
		switch op {
		case token.ADD:
			return xx + yy
		case token.SUB:
			return xx - yy
		case token.MUL:
			return xx * yy
		case token.QUO:
			return xx / yy
		case token.GTR:
			return xx > yy
		case token.GEQ:
			return xx >= yy
		case token.LSS:
			return xx < yy
		case token.LEQ:
			return xx <= yy
		case token.EQL:
			return xx == yy
		case token.NEQ:
			return xx != yy
		case token.LAND:
			return (xx > 0) && (yy > 0)
		case token.LOR:
			return (xx > 0) && (yy > 0)
		}
	}
	if t == token.CHAR {
		xx, _ := x.(uint8)
		yy, _ := y.(uint8)
		switch op {
		case token.ADD:
			return xx + yy
		case token.SUB:
			return xx - yy
		case token.MUL:
			return xx * yy
		case token.QUO:
			return xx / yy
		case token.GTR:
			return xx > yy
		case token.GEQ:
			return xx >= yy
		case token.LSS:
			return xx < yy
		case token.LEQ:
			return xx <= yy
		case token.EQL:
			return xx == yy
		case token.NEQ:
			return xx != yy
		case token.LAND:
			return (xx > 0) && (yy > 0)
		case token.LOR:
			return (xx > 0) && (yy > 0)
		}
	}
	return uint8(0)
}

// 表达式到整数
func (it interpreter) intValue(value interface{}) int32 {
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

func (it interpreter) expectedChar(expr *ast.LitExpr) uint8 {
	if expr.Kind != token.CHAR {
		it.errorf(expr.Pos(), "unexpected token %v", expr)
		return 0
	}
	c, ok := tryCharValue(expr.Value)
	if ok == false {
		it.errorf(expr.Pos(), "error char expr %s", expr.Value)
	}
	return c
}

// 表达式到浮点数
func (it interpreter) floatValue(value interface{}) float64 {
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
	if len(ch) == 1 {
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
		v := it.expectedValue(it.evalValue(expr))
		if vv, ok := v.(uint8); ok {
			return ^vv
		}
		if vv, ok := v.(int32); ok {
			return ^vv
		}
	case token.LNOT: // !
		v := it.expectedValue(it.evalValue(expr))
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
		v := it.expectedValue(it.evalValue(expr))
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
		id := it.expectedIdent(expr.X)
		if id != nil {
			if _, ok := it.Val[id.Name]; ok {
				return uint8(1)
			}
			if _, ok := it.Val[id.Name]; ok {
				return uint8(1)
			}
		} else {
			it.errorf(expr.X.Pos(), "'defined' is not followed by a ident %v", expr.X)
			return uint8(0)
		}
	}
	it.errorf(expr.X.Pos(), "unexpected value %v in unary expr", expr)
	return uint8(0)
}

// 展开ID标识符
// pos 展开标识符的位置
// id 展开的标识符
func (it interpreter) extractIdentPos(pos token.Pos, id *ast.Ident) string {
	if v, ok := it.Val[id.Name]; ok {
		// 如果是宏定义函数，则返回名字
		if _, ok := v.(*ast.FuncDefineStmt); ok {
			return id.Name
		}
		if vv, ok := v.(*ast.ValDefineStmt); ok {
			return it.extractMacroLine(vv.Body, vv.Pos())
		}
		if vv, ok := v.(string); ok {
			return vv
		}
	}
	if id.Name == "__LINE__" {
		return strconv.Itoa(it.pos.CreatePosition(pos).Line)
	}
	return id.Name
}

// 获取宏定义值
func (it interpreter) evalIdent(id *ast.Ident) interface{} {
	vv := ""
	if v, ok := it.Val[id.Name]; ok {
		vv = it.extractMacroLine(v, id.Pos())
	}
	if id.Name == "__LINE__" {
		vv = strconv.Itoa(it.pos.CreatePosition(id.Pos()).Line)
	}
	exp, errs := parser.ParseExpr([]byte(vv), id.Pos())
	if len(errs) > 0 {
		it.errorf(id.Pos(), "error extract ident expr %s", vv)
	}
	return it.evalValue(exp)
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
		vv := it.extractIdentPos(v.Pos(), v)
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
