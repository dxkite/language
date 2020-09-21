package interpreter

import (
	"bytes"
	"dxkite.cn/language/macro/ast"
	"dxkite.cn/language/macro/parser"
	"dxkite.cn/language/macro/token"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// 解释器
type Interpreter struct {
	// 已经定义的宏
	Val map[string]MacroValue
	// 位置信息
	pos token.FilePos
	// 运行后的源码
	src *bytes.Buffer
}

// 执行ast
func (it *Interpreter) Eval(node ast.Node, name string, pos token.FilePos) []byte {
	it.Val = map[string]MacroValue{}
	it.Val["__FILE__"] = MacroString(strconv.QuoteToGraphic(name))
	it.src = &bytes.Buffer{}
	it.pos = pos
	it.evalStmt(node)
	return it.src.Bytes()
}

// 设置宏参数
func (it *Interpreter) SetValue(name, value string) {
	it.Val[name] = MacroString(value)
}

// 获取宏参数
func (it *Interpreter) GetValue(name string) (v MacroValue, exist bool) {
	v, exist = it.Val[name]
	return
}

// 转换成位置
func (it *Interpreter) Position(pos token.Pos) token.Position {
	return it.pos.CreatePosition(pos)
}

// 执行宏语句
func (it *Interpreter) evalStmt(node ast.Node) {
	switch n := node.(type) {
	case *ast.BlockStmt:
		for _, sub := range *n {
			it.evalStmt(sub)
		}
	case *ast.MacroLitArray:
		it.src.WriteString(NewExtractor(it).Extract(n, NewGlobalEnv(token.NoPos)))
	case *ast.Ident:
		it.src.WriteString(NewExtractor(it).Extract(n, NewGlobalEnv(token.NoPos)))
	case *ast.ValDefineStmt:
		it.evalDefineVal(n)
	case *ast.UnDefineStmt:
		it.evalUnDefineStmt(n)
	case *ast.FuncDefineStmt:
		it.evalDefineFunc(n)
	case *ast.IfStmt:
		it.evalIf(n)
	case *ast.ElseIfStmt:
		it.evalElseIf(n)
	case *ast.IfDefStmt:
		it.evalIfDefined(n)
	case *ast.IfNoDefStmt:
		it.evalIfNoDefined(n)
	case *ast.IncludeStmt:
		fmt.Println("include", n.Path)
		it.writePlaceholder(n)
	case *ast.MacroCmdStmt:
		if n.Kind != token.ERROR {
			it.writePlaceholder(n)
		} else {
			it.error(n.Pos(), n.Cmd)
		}
	case *ast.LineStmt:
		fmt.Println("set line", n.Line)
	default:
		fmt.Println("unexpected cmd")
	}
}

// 定义一个宏
func (it *Interpreter) evalDefineVal(stmt *ast.ValDefineStmt) {
	n := stmt.Name.Name
	if _, ok := it.Val[n]; ok || isInnerDefine(n) {
		it.errorf(stmt.Pos(), "warning: %s redefined", n)
	}
	it.Val[n] = &MacroLitValue{it, stmt}
	it.writePlaceholder(stmt)
}

// 定义一个宏函数
func (it *Interpreter) evalDefineFunc(stmt *ast.FuncDefineStmt) {
	n := stmt.Name.Name
	if _, ok := it.Val[n]; ok || isInnerDefine(n) {
		it.errorf(stmt.Pos(), "warning: %s redefined", n)
	}
	it.Val[n] = &MacroFuncValue{it, stmt}
	it.writePlaceholder(stmt)
}

// 取消定义
func (it *Interpreter) evalUnDefineStmt(stmt *ast.UnDefineStmt) {
	n := stmt.Name.Name
	if _, ok := it.Val[n]; ok {
		delete(it.Val, n)
	}
	it.writePlaceholder(stmt)
}

// 检测内置预定义变量
// __LINE__ 展开为当前行
// __FUNCTION__ 不做处理
func isInnerDefine(name string) bool {
	ar := []string{"__LINE__", "__FUNCTION__"}
	for _, n := range ar {
		if n == name {
			return true
		}
	}
	return false
}

// 宏占位
func (it *Interpreter) writePlaceholder(node ast.Node) {
	f := it.pos.CreatePosition(node.Pos()).Line
	t := it.pos.CreatePosition(node.End()).Line
	it.src.WriteString(strings.Repeat("\n", t-f+1))
}

func (it Interpreter) parseDefined(v *ast.UnaryExpr) string {
	s := "defined"
	s += " " + it.parseDefinedValue(v.X)
	return s
}

func (it Interpreter) parseDefinedValue(v ast.MacroLiter) string {
	if id, ok := v.(*ast.Ident); ok {
		return id.Name
	} else if p, ok := v.(*ast.ParenExpr); ok {
		return "(" + it.parseDefinedValue(p.X) + ")"
	} else {
		it.errorf(v.Pos(), "unexpected expr %s in defined", reflect.TypeOf(v))
		return ""
	}
}

// #if
func (it *Interpreter) evalIf(stmt *ast.IfStmt) {
	v := it.evalIfBoolExpr(stmt.X)
	// #if
	it.writePlaceholder(stmt.X)
	it.evalCondition(v, stmt.Then, stmt.Else)
	it.src.WriteString("\n") // #endif
}

// #elif
func (it *Interpreter) evalElseIf(stmt *ast.ElseIfStmt) {
	v := it.evalIfBoolExpr(stmt.X)
	it.evalCondition(v, stmt.Then, stmt.Else)
}

// #ifdef
func (it *Interpreter) evalIfDefined(stmt *ast.IfDefStmt) {
	v := it.evalDefined(stmt.Name, "#ifdef")
	// #ifdef
	it.writePlaceholder(stmt.Name)
	it.evalCondition(v, stmt.Then, stmt.Else)
	it.src.WriteString("\n") // #endif
}

// #ifndef
func (it *Interpreter) evalIfNoDefined(stmt *ast.IfNoDefStmt) {
	v := it.evalDefined(stmt.Name, "#ifndef")
	// #ifdef
	it.writePlaceholder(stmt.Name)
	it.evalCondition(!v, stmt.Then, stmt.Else)
	it.src.WriteString("\n") // #endif
}

func (it *Interpreter) evalIfBoolExpr(expr ast.MacroLiter) bool {
	ee := NewExtractor(it).Extract(expr, NewGlobalEnv(expr.Pos()))
	//fmt.Println("expr", strconv.QuoteToGraphic(ee))
	return it.evalExpr(ee, expr.Pos())
}

func (it *Interpreter) evalCondition(v bool, ts, fs ast.Stmt) {
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

// 二元运算
func (it Interpreter) evalBinaryExpr(expr *ast.BinaryExpr) interface{} {
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
func (it Interpreter) evalExpr(expr string, pos token.Pos) bool {
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
func (it *Interpreter) evalValue(expr interface{}) interface{} {
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
	case *ast.ParenExpr:
		return it.evalValue(xx.X)
	case *ast.UnaryExpr:
		return it.evalUnaryExpr(xx)
	case *ast.BinaryExpr:
		return it.evalBinaryExpr(xx)
	case *ast.MacroCallExpr:
		return NewExtractor(it).Extract(xx, NewGlobalEnv(xx.Pos()))
	case ast.MacroLiter:
		it.errorf(xx.Pos(), "unexpected token %v", xx)
	}
	return nil
}

// 解析Ident
func (it Interpreter) expectedIdent(expr ast.MacroLiter) *ast.Ident {
	switch xx := expr.(type) {
	case *ast.Ident:
		return xx
	case *ast.ParenExpr:
		return it.expectedIdent(xx.X)
	}
	return nil
}

func (it Interpreter) evalOpInt(x, y interface{}, ex, ey ast.MacroLiter, tx, ty, op token.Token) interface{} {
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
func (it Interpreter) evalOpCast(x, y interface{}, t, op token.Token) interface{} {
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
func (it Interpreter) intValue(value interface{}) int32 {
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

func (it Interpreter) expectedChar(expr *ast.LitExpr) uint8 {
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
func (it Interpreter) floatValue(value interface{}) float64 {
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
func (it Interpreter) evalUnaryExpr(expr *ast.UnaryExpr) interface{} {
	switch expr.Op {
	case token.NOT: // ~
		v := it.expectedValue(it.evalValue(expr.X))
		if vv, ok := v.(uint8); ok {
			return ^vv
		}
		if vv, ok := v.(int32); ok {
			return ^vv
		}
	case token.LNOT: // !
		v := it.expectedValue(it.evalValue(expr.X))
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
		v := it.expectedValue(it.evalValue(expr.X))
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
		if it.evalDefined(expr.X, "defined") {
			return uint8(1)
		}
		return uint8(0)
	}
	it.errorf(expr.X.Pos(), "unexpected value %v in %s expr", expr.X, expr.Op)
	return uint8(0)
}

// 定义指令
func (it Interpreter) evalDefined(expr ast.MacroLiter, typ string) bool {
	id := it.expectedIdent(expr)
	if id != nil {
		if _, ok := it.Val[id.Name]; ok {
			return true
		}
		return false
	}
	it.errorf(expr.Pos(), "'%s' is not followed by a ident %v", typ, expr)
	return false
}

// 获取宏定义值
func (it *Interpreter) evalIdent(id *ast.Ident) interface{} {
	if v, ok := NewExtractor(it).Ident(id, NewGlobalEnv(id.Pos())); ok {
		exp, errs := parser.ParseExpr([]byte(v), id.Pos())
		if len(errs) > 0 {
			it.errorf(id.Pos(), "error Extract ident expr %s", v)
		}
		return it.evalValue(exp)
	}
	return uint8(0)
}

// 转换表达式到值
func (it *Interpreter) expectedValue(value interface{}) interface{} {
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
		vv := NewExtractor(it).Extract(v, NewGlobalEnv(v.Pos()))
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
	case ast.MacroLiter:
		it.errorf(v.Pos(), "unexpected value %v", v)
	}
	return uint8(0)
}

// 解析数字
func (it Interpreter) evalInt(expr *ast.LitExpr) int32 {
	v, err := strconv.ParseInt(expr.Value, 0, 32)
	if err != nil {
		it.errorf(expr.Offset, "error parse int %s", err.Error())
	}
	return int32(v)
}

// 解析数字（浮点数）
func (it Interpreter) evalFloat(expr *ast.LitExpr) float64 {
	v, err := strconv.ParseFloat(expr.Value, 64)
	if err != nil {
		it.errorf(expr.Offset, "error parse float %s", err.Error())
	}
	return v
}

func (it Interpreter) error(pos token.Pos, msg string) {
	fmt.Println("error", it.pos.CreatePosition(pos), msg)
}

func (it Interpreter) errorf(pos token.Pos, format string, args ...interface{}) {
	it.error(pos, fmt.Sprintf(format, args...))
}
