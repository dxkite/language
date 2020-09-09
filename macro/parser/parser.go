package parser

import (
	"dxkite.cn/language/macro/ast"
	"dxkite.cn/language/macro/scanner"
	"dxkite.cn/language/macro/token"
	"errors"
	"fmt"
)

type parser struct {
	scanner *scanner.Scanner  // 扫描器
	errors  scanner.ErrorList // 错误列表
	// 下一个Token
	pos  token.Pos      // Token位置
	tok  token.Token    // Token
	lit  string         // 内容
	root *ast.BlockStmt // 根节点
}

func (p *parser) init(src []byte) {
	p.scanner = &scanner.Scanner{}
	p.scanner.Init(src)
	p.next()
	p.root = &ast.BlockStmt{}
}

func (p *parser) parse() {
	for p.tok != token.EOF {
		var node ast.Stmt
		if p.tok == token.MACRO {
			node = p.parseMacro()
		} else {
			node = p.parseTextLine()
		}
		if node != nil {
			p.root.Add(node)
		}
	}
}

func (p *parser) parseTextLine() ast.Stmt {
	node := ast.MacroLitArray{}
	for TokenNotIn(p.tok, token.EOF, token.MACRO) {
		// 解析
		if TokenIn(p.tok, token.IDENT, token.FLOAT, token.INT, token.SHARP) {
			node.AddLit(p.parseMacroLiter())
		} else {
			node.AddLit(p.parseTextOrLit())
		}
	}
	return nilIfEmpty(node)
}

// 解析宏
func (p *parser) parseMacro() (node ast.Stmt) {
	from := p.pos
	p.next()
	p.skipWhitespace()
	switch p.tok {
	case token.INCLUDE:
		node = p.parseInclude(from)
	case token.ERROR:
		node = p.parseError(from)
	case token.LINE, token.INT:
		node = p.parseLine(from)
	case token.DEFINE:
		node = p.parseDefine(from)
	case token.UNDEF:
		node = p.parseUnDefine(from)
	case token.IF, token.IFNODEF, token.IFDEF:
		node = p.parseIf()
	default:
		node = p.parseNop(from)
	}
	return
}

// #include <path>
// #include "path"
func (p *parser) parseInclude(from token.Pos) ast.Stmt {
	p.next()
	p.skipWhitespace()
	pos, tok, lit := p.next()
	var path string
	typ := ast.IncludeOuter
	if tok == token.STRING {
		path = lit
	} else if tok == token.LSS {
		path = lit
		typ = ast.IncludeInner
		for {
			if tok == token.GTR || tok == token.LF || tok == token.EOF {
				if tok == token.LF {
					p.next()
				}
				break
			}
			pos, tok, lit = p.next()
			path += lit
		}
	} else {
		p.errorf(pos, "unexpected token %v", tok)
	}
	p.scanToMacroEnd(true)
	return &ast.IncludeStmt{
		From: from,
		To:   pos,
		Path: path,
		Type: typ,
	}
}

// #error message
func (p *parser) parseError(from token.Pos) ast.Stmt {
	node := &ast.ErrorStmt{
		Offset: from,
		Msg:    p.lit,
	}
	to := p.scanToMacroEnd(false)
	node.Msg = clearBackslash(p.scanner.Lit(from, to))
	return node
}

// #line number
// #line number "file"
// # number "file"
// # number
func (p *parser) parseLine(from token.Pos) ast.Stmt {
	node := &ast.LineStmt{From: from}
	if p.tok == token.INT {
		node.Line = p.lit
		node.To = p.pos + token.Pos(len(p.lit))
		p.next()
	} else {
		p.next()
		p.skipWhitespace()
		node.To, _, node.Line = p.expected(token.INT)
		node.To = node.To + token.Pos(len(node.Line))
	}
	p.skipWhitespace()
	if p.tok == token.STRING {
		node.Path = p.lit
		node.To = p.pos + token.Pos(len(p.lit))
		p.next()
	}
	p.scanToMacroEnd(true)
	return node
}

// # any LF
func (p *parser) parseNop(from token.Pos) ast.Stmt {
	node := &ast.NopStmt{Offset: p.pos}
	p.next()
	to := p.scanToMacroEnd(false)
	node.Text = p.scanner.Lit(from, to)
	return node
}

func (p *parser) parseUnDefine(from token.Pos) ast.Stmt {
	p.next()
	p.skipWhitespace()
	pos, _, name := p.expected(token.IDENT)
	id := &ast.Ident{
		Offset: pos,
		Name:   name,
	}
	end := p.pos
	p.scanToMacroEnd(true)
	return &ast.ValUnDefineStmt{
		From: from,
		To:   end,
		Name: id,
	}
}

// define_statement    =
//    ( macro_prefix "#define" define_literal code_line ) .
func (p *parser) parseDefine(from token.Pos) ast.Stmt {
	p.next()
	p.skipWhitespace()
	pos, _, name := p.expected(token.IDENT)
	id := &ast.Ident{
		Offset: pos,
		Name:   name,
	}
	if p.tok == token.LPAREN {
		node := &ast.FuncDefineStmt{From: from, Name: id}
		lp, _, _ := p.expected(token.LPAREN)
		list, err := p.parseIdentList()
		if err != nil {
			return &ast.BadExpr{
				Offset: from,
				Token:  token.DEFINE,
				Lit:    p.scanner.Lit(from, p.pos),
			}
		}
		rp, _, _ := p.expected(token.RPAREN)
		node.IdentList = list
		node.LParam = lp
		node.RParam = rp
		p.skipWhitespace()
		node.Body = p.parseMacroBody()
		node.To = p.pos
		p.scanToMacroEnd(true)
		return node
	} else {
		node := &ast.ValDefineStmt{From: from, Name: id}
		p.skipWhitespace()
		node.Body = p.parseMacroBody()
		node.To = p.pos
		p.scanToMacroEnd(true)
		return node
	}
}

func (p *parser) parseIdentList() (params []*ast.Ident, err error) {
	params = []*ast.Ident{}
	p.skipWhitespace()
	for TokenNotIn(p.tok, token.RPAREN, token.LF, token.EOF) {
		if p.tok != token.IDENT {
			goto errExit
		} else {
			params = append(params, &ast.Ident{
				Offset: p.pos,
				Name:   p.lit,
			})
			p.next() // ident
			p.skipWhitespace()
			if TokenIn(p.tok, token.RPAREN) {
				return
			}
			if p.tok != token.COMMA {
				goto errExit
			}
			p.next() // ,
			p.skipWhitespace()
		}
	}
	return
errExit:
	msg := fmt.Sprintf("%v may not appear in macro parameter list", p.tok)
	p.error(p.pos, msg)
	err = errors.New(msg)
	p.scanToMacroEnd(false)
	return
}

// 解析宏定义体
// macro_body =
//    < text_line / macro_literal > .
func (p *parser) parseMacroBody() (node ast.MacroLitArray) {
	node = ast.MacroLitArray{}
	for !isMacroEnd(p.tok) {
		// 解析
		if TokenIn(p.tok, token.IDENT, token.FLOAT, token.INT, token.SHARP) {
			node.AddLit(p.parseMacroLiter())
		} else {
			node.AddLit(p.parseTextOrLit())
		}
	}
	return nilIfEmpty(node)
}

// 解析
// macro_literal =
//    ( macro_expr ["##" macro_expr ] ) .
func (p *parser) parseMacroLiter() (node ast.MacroLiter) {
	x := p.parseMacroExpr()
	if _, ok := p.tryNextNotEmpty(token.DOUBLE_SHARP); ok {
		p.skipWhitespace()
		offs := p.pos
		p.next()
		p.skipWhitespace()
		y := p.parseMacroExpr()
		return &ast.BinaryExpr{
			X:      x,
			Offset: offs,
			Op:     token.DOUBLE_SHARP,
			Y:      y,
		}
	}
	return x
}

// 解析
// macro_expr =
//    identifier
//    / integer_literal
//    / float_literal
//    / ( "#" identifier )
//    / macro_call_expr .
func (p *parser) parseMacroExpr() (node ast.MacroLiter) {
	pos, tok, name := p.pos, p.tok, p.lit
	if TokenIn(p.tok, token.FLOAT, token.INT) {
		lit := &ast.LitExpr{
			Offset: pos,
			Kind:   tok,
			Value:  name,
		}
		p.next()
		return lit
	} else if p.tok == token.SHARP {
		p.next()
		offs, typ, lit := p.pos, p.tok, p.lit
		if typ == token.IDENT {
			node = &ast.UnaryExpr{
				Offset: pos,
				Op:     token.SHARP,
				X: &ast.Ident{
					Offset: offs,
					Name:   lit,
				},
			}
			p.next()
		} else {
			node = &ast.Text{
				Offset: pos,
				Kind:   tok,
				Text:   name,
			}
		}
		return node
	} else if p.tok == token.IDENT || p.tok.IsKeyword() {
		return p.parseMacroCallExprOrIdent()
	}
	return &ast.BadExpr{
		Offset: p.pos,
		Token:  p.tok,
		Lit:    p.lit,
	}
}

// macro_call_expr =
//    identifier ( "("  [ macro_argument ]  ")"  ) .
func (p *parser) parseMacroCallExprOrIdent() (node ast.MacroLiter) {
	id := &ast.Ident{
		Offset: p.pos,
		Name:   p.lit,
	}
	p.next()
	if _, ok := p.tryNextNotEmpty(token.LPAREN); ok {
		p.skipWhitespace()
		lp, _, _ := p.expected(token.LPAREN)
		list := p.parseMacroArgument()
		rp, _, _ := p.expected(token.RPAREN)
		return &ast.MacroCallExpr{
			From:      id.Pos(),
			To:        p.pos,
			LParam:    lp,
			Name:      id,
			RParam:    rp,
			ParamList: list,
		}
	}
	return id
}

// 尝试找到 token
func (p *parser) tryNextNotEmpty(tok ...token.Token) (token.Token, bool) {
	pp := p.clone()
	defer p.reset(pp)
	p.skipWhitespace()
	if !isMacroEnd(p.tok) {
		if TokenIn(p.tok, tok...) {
			return p.tok, true
		}
	}
	return 0, false
}

// 检查下一个 token
func (p *parser) tryNext(tok ...token.Token) (token.Token, bool) {
	pp := p.clone()
	defer p.reset(pp)
	if !isMacroEnd(p.tok) {
		if TokenIn(p.tok, tok...) {
			return p.tok, true
		}
	}
	return 0, false
}

// 检查下一个 token
func (p *parser) tryNextIs(tok token.Token) bool {
	_, ok := p.tryNext(token.LF)
	return ok
}

// 找到调用参数
// macro_argument =
//    macro_param_lit  <  "," macro_param_lit  >  .
func (p *parser) parseMacroArgument() (node ast.MacroLitArray) {
	node = ast.MacroLitArray{}
	for TokenNotIn(p.tok, token.RPAREN) && !isMacroEnd(p.tok) {
		lit := p.parseMacroArgumentLit()
		node.AddLit(lit)
		if p.tok == token.COMMA {
			p.next() // ,
		}
	}
	return nilIfEmpty(node)
}

// 参数字面量
// macro_param_lit =
//    < macro_item > .
func (p *parser) parseMacroArgumentLit() (node ast.MacroLiter) {
	list := ast.MacroLitArray{}
	for TokenNotIn(p.tok, token.COMMA, token.RPAREN) && !isMacroEnd(p.tok) {
		lit := p.parseMacroArgumentLitItem()
		list.AddLit(lit)
	}
	if len(list) == 1 {
		return list[0]
	}
	return nilIfEmpty(list)
}

// macro_item =
//    identifier
//    / macro_call_expr
//    / "text not , " .
func (p *parser) parseMacroArgumentLitItem() (node ast.MacroLiter) {
	if _, ok := p.tryNextNotEmpty(token.IDENT); ok {
		p.skipWhitespace()
		return p.parseMacroCallExprOrIdent()
	}
	return p.parseTextOrLit()
}

func (p *parser) parseTextOrLit() (node ast.MacroLiter) {
	if TokenIn(p.tok, token.INT, token.STRING, token.CHAR, token.FLOAT) {
		node = &ast.LitExpr{
			Offset: p.pos,
			Kind:   p.tok,
			Value:  p.lit,
		}
	} else {
		node = &ast.Text{
			Offset: p.pos,
			Kind:   p.tok,
			Text:   p.lit,
		}
	}
	p.next()
	return
}

func (p *parser) parseComment() ast.Stmt {
	return &ast.Comment{
		Offset: p.pos,
		Kind:   p.tok,
		Text:   p.lit,
	}
}

func (p *parser) parseIf() ast.Stmt {
	node := &ast.IfStmt{From: p.pos}
	return node
}

// 解析表达式
// expr =
// 	(numeric__expr)
func (p *parser) parseExpr() (expr ast.Expr) {
	return p.parseExprPrecedence(token.LowestPrec)
}

func (p *parser) parseExprPrecedence(prec int) (expr ast.Expr) {
	p.skipWhitespace()
	expr = p.parseOpExpr(prec + 1)
	p.skipWhitespace()
	for p.tok.Precedence() > prec {
		op := p.tok
		offs := p.pos
		p.next()
		p.skipWhitespace()
		y := p.parseOpExpr(prec + 1)
		p.skipWhitespace()
		expr = &ast.BinaryExpr{
			X:      expr,
			Offset: offs,
			Op:     op,
			Y:      y,
		}
	}
	return expr
}

// 	( ("-" / "~" / "defined" / "!" ) parseTermExpr )
func (p *parser) parseOpExpr(prec int) (expr ast.Expr) {
	if prec >= token.UnaryPrec {
		return p.parseUnaryExpr()
	} else {
		return p.parseExprPrecedence(prec + 1)
	}
}

// 	( ("-" / "~" / "defined" / "!" ) parseTermExpr )
func (p *parser) parseUnaryExpr() ast.Expr {
	p.skipWhitespace()
	var expr *ast.UnaryExpr
	var last *ast.UnaryExpr
	var tok token.Token
	for TokenIn(p.tok, token.LNOT, token.DEFINED, token.NOT, token.SUB) {
		offs := p.pos
		op := p.tok
		p.next()
		p.skipWhitespace()
		last = &ast.UnaryExpr{
			Offset: offs,
			Op:     op,
			X:      nil,
		}
		if expr == nil {
			expr = last
		} else {
			expr.X = last
		}
	}
	if last != nil {
		last.X = p.parseUnaryX(tok)
		return expr
	}
	return p.parseTermExpr()
}

func (p *parser) parseUnaryX(op token.Token) (expr ast.Expr) {
	if op == token.DEFINED {
		off, _, name := p.expected(token.IDENT)
		return &ast.Ident{
			Offset: off,
			Name:   name,
		}
	} else {
		return p.parseExpr()
	}
}

// 	"(" expr ")"
//  / numeric_expression
//  / identifier
//  / macro_call_expr.
func (p *parser) parseTermExpr() (expr ast.Expr) {
	p.skipWhitespace()
	switch p.tok {
	case token.LPAREN:
		p.expected(token.LPAREN)
		expr = p.parseExpr()
		p.expected(token.LPAREN)
		return expr
	case token.INT, token.FLOAT, token.CHAR, token.STRING:
		return p.parseLiteralExpr()
	case token.IDENT:
		return p.parseMacroCallExprOrIdent()
	}
	if p.tok.IsKeyword() {
		return &ast.Ident{
			Offset: p.pos,
			Name:   p.lit,
		}
	}
	return &ast.BadExpr{
		Offset: p.pos,
		Token:  p.tok,
		Lit:    p.lit,
	}
}

// 字符类型
// literal_expression =
//    integer_literal token.INT
//    / float_literal token.FLOAT
//    / string token.STRING
//    / char  token.CHAR
func (p *parser) parseLiteralExpr() (expr ast.Expr) {
	expr = &ast.LitExpr{
		Offset: p.pos,
		Kind:   p.tok,
		Value:  p.lit,
	}
	p.next()
	return
}

// 获取下一个Token
func (p *parser) next() (pos token.Pos, tok token.Token, lit string) {
	pos, tok, lit = p.pos, p.tok, p.lit
	p.pos, p.tok, p.lit = p.scanner.Scan()
	return
}

func (p *parser) expected(typ ...token.Token) (pos token.Pos, tok token.Token, lit string) {
	pos, tok, lit = p.pos, p.tok, p.lit
	if !tokenIn(tok, typ) {
		p.errorf(pos, "expected %v token, got %v", typ, tok)
	}
	p.next()
	return
}

func TokenNotIn(tok token.Token, typ ...token.Token) bool {
	return !tokenIn(tok, typ)
}

func isMacroEnd(tok token.Token) bool {
	return TokenIn(tok, token.EOF, token.LF)
}

func TokenIn(tok token.Token, typ ...token.Token) bool {
	return tokenIn(tok, typ)
}

func tokenIn(tok token.Token, typ []token.Token) bool {
	for _, tt := range typ {
		if tok == tt {
			return true
		}
	}
	return false
}

func (p *parser) skipWhitespace() string {
	t := ""
	for p.tok == token.TEXT && IsEmptyText(p.lit) {
		t += p.lit
		p.next()
	}
	return t
}

// 清除空白
func clearBackslash(text string) string {
	t := ""
	l := len(text)
	for i := 0; i < l; i++ {
		b := text[i]
		if b == '\\' && i+1 < l && text[i+1] == '\n' {
			i++
		} else {
			t += string(b)
		}
	}
	return t
}

// 判断是否是空文本
func IsEmptyText(text string) bool {
	for _, b := range text {
		switch b {
		case ' ', '\t', '\r':
		default:
			return false
		}
	}
	return true
}

// 扫描到宏末尾
func (p *parser) scanToMacroEnd(needEmpty bool) (pos token.Pos) {
	for !isMacroEnd(p.tok) {
		cur := p.tok
		if needEmpty && !p.isMacroLineEmpty() && (cur == token.BACKSLASH && p.tryNextIs(token.LF)) {
			p.errorf(p.pos, "unexpected token %v", p.tok)
		}
		p.next() // current
		if cur == token.BACKSLASH && p.tok == token.LF {
			p.next() // \n
		}
	}
	if p.tok == token.LF {
		pos = p.pos
		p.next()
		return
	}
	return p.pos
}

// 当前为空
func (p *parser) isMacroLineEmpty() bool {
	if p.tok == token.COMMENT || p.tok == token.BLOCK_COMMENT {
		return true
	}
	if p.tok == token.TEXT && IsEmptyText(p.lit) {
		return true
	}
	return false
}

func (p *parser) error(pos token.Pos, msg string) {
	fmt.Println("error", pos, msg)
}

func (p *parser) errorf(pos token.Pos, format string, args ...interface{}) {
	p.error(pos, fmt.Sprintf(format, args...))
}

// 保存状态
func (p *parser) clone() *parser {
	return &parser{
		scanner: p.scanner.CloneWithoutSrc(),
		errors:  p.errors,
		pos:     p.pos,
		tok:     p.tok,
		lit:     p.lit,
	}
}

// 回复状态
func (p *parser) reset(state *parser) {
	root := p.root
	src := p.scanner.GetSrc()
	*p = *state
	p.root = root
	p.scanner.SetSrc(src)
}

func Parse(src []byte) ast.Node {
	p := &parser{}
	p.init(src)
	p.parse()
	return p.root
}

func nilIfEmpty(array ast.MacroLitArray) ast.MacroLitArray {
	if len(array) > 0 {
		return array
	}
	return nil
}
