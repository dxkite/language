package parser

import (
	"dxkite.cn/language/macro/ast"
	"dxkite.cn/language/macro/scanner"
	"dxkite.cn/language/macro/token"
	"errors"
	"fmt"
)

type Parser struct {
	scanner scanner.Scanner   // 扫描器
	errors  scanner.ErrorList // 错误列表
	// 下一个Token
	pos token.Pos   // Token位置
	tok token.Token // Token
	lit string      // 内容
}

func (p *Parser) Init(src []byte) {
	p.scanner = scanner.NewScanner(src)
	p.errors = scanner.ErrorList{}
	p.next()
}

func (p *Parser) InitOffset(src []byte, tok token.Pos) {
	p.scanner = scanner.NewOffsetScanner(src, tok)
	p.errors = scanner.ErrorList{}
	p.next()
}

// 解析宏语句
func (p *Parser) Parse() ast.Node {
	return p.ParseStmts()
}

// 解析语句
func (p *Parser) ParseStmts() ast.Stmt {
	block := &ast.BlockStmt{}
	for p.tok != token.EOF {
		var node ast.Stmt
		if p.tok == token.MACRO {
			from := p.pos
			p.next()
			p.skipWhitespace()
			if TokenIn(p.tok, token.IF, token.IFDEF, token.IFNDEF) {
				node = p.parseMacroLogicStmt(from)
			} else {
				node = p.parseMacroStmt(from)
			}
		} else {
			node = p.ParseTextStmt()
		}
		if node != nil {
			block.Add(node)
		}
	}
	return block
}

// 获取当前的宏名
func (p *Parser) curMacroIs(tok ...token.Token) bool {
	pp := p.clone()
	defer p.reset(pp)
	p.next() // #
	p.skipWhitespace()
	return TokenIn(p.tok, tok...)
}

// 解析体
func (p *Parser) parseBodyStmts() ast.Stmt {
	block := &ast.BlockStmt{}
	for p.tok != token.EOF {
		var node ast.Stmt
		if p.tok == token.MACRO {
			if p.curMacroIs(p.tok, token.EOF, token.ELSE, token.ELSEIF, token.ENDIF) {
				break
			}
			from := p.pos
			p.next()
			p.skipWhitespace()
			if TokenIn(p.tok, token.IF, token.IFDEF, token.IFNDEF) {
				node = p.parseMacroLogicStmt(from)
			} else {
				node = p.parseMacroStmt(from)
			}
		} else {
			node = p.ParseTextStmt()
		}
		if node != nil {
			block.Add(node)
		}
	}
	return block
}

// 解析宏语句
func (p *Parser) parseMacroStmt(from token.Pos) (node ast.Stmt) {
	switch p.tok {
	case token.INCLUDE:
		node = p.parseInclude(from)
	case token.ERROR, token.PRAGMA, token.WARNING:
		node = p.parseCmd(from)
	case token.LINE, token.INT:
		node = p.parseLine(from)
	case token.DEFINE:
		node = p.parseDefine(from)
	case token.UNDEF:
		node = p.parseUnDefine(from)
	default:
		node = p.parseInvalidStmt(from)
	}
	return
}

// 解析宏语句命令
func (p *Parser) parseMacroLogicStmt(from token.Pos) (node ast.CondStmt) {
	_, tok, _ := p.next()
	var cond ast.CondStmt
	if tok == token.IF {
		cond = &ast.IfStmt{
			X: p.parseIfExpr(),
		}
	} else {
		p.next()
		p.skipWhitespace()
		off, _, name := p.expectedIdent()
		id := &ast.Ident{
			Offset: off,
			Name:   name,
		}
		if tok == token.IFDEF {
			cond = &ast.IfDefStmt{
				Name: id,
			}
		} else {
			cond = &ast.IfNoDefStmt{
				Name: id,
			}
		}
		p.scanToMacroEnd(true)
	}

	node = cond
	cond.SetTrueStmt(p.parseBodyStmts())

	for p.tok == token.MACRO && p.curMacroIs(token.ELSEIF) {
		off := p.pos
		p.next() // #
		p.skipWhitespace()
		p.next() // elseif
		eif := &ast.ElseIfStmt{
			X: p.parseIfExpr(),
		}
		tt := p.parseBodyStmts()
		eif.SetTrueStmt(tt)
		eif.SetFromTO(off, p.pos)
		cond.SetFalseStmt(eif)
		cond = eif
	}

	if p.tok == token.MACRO && p.curMacroIs(token.ELSE) {
		p.next() // #
		p.skipWhitespace()
		p.next() // else
		p.scanToMacroEnd(true)
		cond.SetFalseStmt(p.parseBodyStmts())
	}

	if p.tok == token.MACRO && p.curMacroIs(token.ENDIF) {
		p.next() // #
		p.skipWhitespace()
		p.next() // endif
		p.scanToMacroEnd(true)
	} else {
		p.errorf(p.pos, "expected #endif")
	}
	node.SetFromTO(from, p.pos)
	return
}

// 提取表达式，不解析
func (p *Parser) parseIfExpr() ast.Expr {
	p.skipWhitespace()
	from := p.pos
	to := p.scanToMacroEnd(false)
	expr := clearBackslash(p.scanner.Lit(from, to))
	return &ast.Text{
		Offset: from,
		Kind:   token.TEXT,
		Text:   expr,
	}
}

// #include <path>
// #include "path"
func (p *Parser) parseInclude(from token.Pos) ast.Stmt {
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
			if tok == token.GTR || tok == token.NEWLINE || tok == token.EOF {
				if tok == token.NEWLINE {
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
func (p *Parser) parseCmd(from token.Pos) ast.Stmt {
	node := &ast.MacroCmdStmt{
		Offset: from,
		Cmd:    p.lit,
		Kind:   p.tok,
	}
	to := p.scanToMacroEnd(false)
	node.Cmd = clearBackslash(p.scanner.Lit(from, to))
	return node
}

// #line number
// #line number "file"
// # number "file"
// # number
func (p *Parser) parseLine(from token.Pos) ast.Stmt {
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

// # any NEWLINE
func (p *Parser) parseInvalidStmt(from token.Pos) ast.Stmt {
	node := &ast.InvalidStmt{Offset: p.pos}
	p.errorf(p.pos, "invalid preprocessing directive #%s", p.lit)
	p.next()
	to := p.scanToMacroEnd(false)
	node.Text = p.scanner.Lit(from, to)
	return node
}

// #undef ident
func (p *Parser) parseUnDefine(from token.Pos) ast.Stmt {
	p.next()
	p.skipWhitespace()
	pos, _, name := p.expectedIdent()
	id := &ast.Ident{
		Offset: pos,
		Name:   name,
	}
	end := p.pos
	p.scanToMacroEnd(true)
	return &ast.UnDefineStmt{
		From: from,
		To:   end,
		Name: id,
	}
}

// define_statement    =
//    ( macro_prefix "#define" define_literal code_line ) .
func (p *Parser) parseDefine(from token.Pos) ast.Stmt {
	p.next()
	p.skipWhitespace()
	pos, _, name := p.expectedIdent()
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
		node.Body = p.parseMacroFuncBody()
		node.To = p.pos
		p.scanToMacroEnd(true)
		return node
	} else {
		node := &ast.ValDefineStmt{From: from, Name: id}
		p.skipWhitespace()
		node.Body = p.parseMacroTextBody()
		node.To = p.pos
		p.scanToMacroEnd(true)
		return node
	}
}

// 解析参数列表
func (p *Parser) parseIdentList() (params []*ast.Ident, err error) {
	params = []*ast.Ident{}
	p.skipWhitespace()
	for TokenNotIn(p.tok, token.RPAREN, token.NEWLINE, token.EOF) {
		if !isIdent(p.tok) {
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

// 解析宏函数定义体
// macro_body =
//    < text_line / macro_literal > .
func (p *Parser) parseMacroFuncBody() (node *ast.MacroLitArray) {
	node = &ast.MacroLitArray{}
	for !isMacroEnd(p.tok) {
		// 解析
		if TokenIn(p.tok, token.IDENT, token.FLOAT, token.INT, token.SHARP) || isIdent(p.tok) {
			node.Append(p.parseMacroExpr())
		} else {
			node.Append(p.parseText())
		}
	}
	return nilIfEmpty(node)
}

// 解析宏定义体
func (p *Parser) parseMacroTextBody() (node *ast.MacroLitArray) {
	node = &ast.MacroLitArray{}
	for !isMacroEnd(p.tok) {
		// 解析
		if isIdent(p.tok) {
			node.Append(p.parseMacroTermExpr())
		} else {
			node.Append(p.parseText())
		}
	}
	return nilIfEmpty(node)
}

// 解析文本语句
func (p *Parser) ParseTextStmt() (node *ast.MacroLitArray) {
	node = &ast.MacroLitArray{}
	for TokenNotIn(p.tok, token.EOF, token.MACRO) {
		// 解析
		if TokenIn(p.tok, token.IDENT) || p.tok.IsKeyword() {
			node.Append(p.parseMacroLitExpr(false))
		} else {
			node.Append(p.parseText())
		}
	}
	return nilIfEmpty(node)
}

// 解析宏表达式
func (p *Parser) parseMacroExpr() (node ast.MacroLiter) {
	expr := p.parseMacroTermExpr()
	for p.nextNotEmptyIs(token.DOUBLE_SHARP) {
		p.skipWhitespace()
		offs := p.pos
		p.next()
		p.skipWhitespace()
		y := p.parseMacroTermExpr()
		expr = &ast.BinaryExpr{
			X:      expr,
			Offset: offs,
			Op:     token.DOUBLE_SHARP,
			Y:      y,
		}
	}
	return expr
}

// 表达式终结符
func (p *Parser) parseMacroTermExpr() (node ast.MacroLiter) {
	pos, tok, name := p.pos, p.tok, p.lit
	if TokenIn(p.tok, token.FLOAT, token.INT) {
		lit := &ast.LitExpr{
			Offset: pos,
			Kind:   tok,
			Value:  name,
		}
		p.next()
		return lit
	} else if isIdent(p.tok) {
		return p.parseIdentExpr()
	}
	return p.parseMacroUnaryExpr()
}

// 表达式一元运算
func (p *Parser) parseMacroUnaryExpr() (node ast.MacroLiter) {
	from := p.pos
	p.next() // #
	off, _, name := p.expectedIdent()
	x := &ast.Ident{
		Offset: off,
		Name:   name,
	}
	return &ast.UnaryExpr{
		Offset: from,
		Op:     token.SHARP,
		X:      x,
	}
}

// 解析标识符字面量
// macro_liter =
//    identifier
//    / identifier ( "("  [ macro_argument ]  ")"  ) .
func (p *Parser) parseIdentExpr() (node ast.MacroLiter) {
	return p.parseMacroLitExpr(true)
}

// 解析标识符字面量
// macro_liter =
//    identifier
//    / identifier ( "("  [ macro_argument ]  ")"  ) .
func (p *Parser) parseMacroLitExpr(inMacro bool) (node ast.MacroLiter) {
	id := &ast.Ident{
		Offset: p.pos,
		Name:   p.lit,
	}
	p.next()
	if p.tryParenPair(inMacro) {
		p.skipWhitespace()
		lp, _, _ := p.expected(token.LPAREN)
		list := p.parseMacroArgument(inMacro)
		rp, _, _ := p.expected(token.RPAREN)
		return &ast.MacroCallExpr{
			From:      id.Pos(),
			To:        p.pos,
			LParen:    lp,
			Name:      id,
			RParen:    rp,
			ParamList: list,
		}
	}
	return id
}

// 查找完整的()
func (p *Parser) tryParenPair(inMacro bool) bool {
	pp := p.clone()
	defer p.reset(pp)
	p.skipWhitespace()
	if p.tok == token.LPAREN {
		dp := 1
		p.next()
		for !isMacroArgEnd(inMacro, p.tok) {
			if p.tok == token.LPAREN {
				dp++
			}
			if p.tok == token.RPAREN {
				dp--
				if dp <= 0 {
					return true
				}
			}
			p.next()
		}
	}
	return false
}

// 找到调用参数
// macro_argument =
//    macro_param_lit  <  "," macro_param_lit  >  .
func (p *Parser) parseMacroArgument(inMacro bool) (node *ast.MacroLitArray) {
	node = &ast.MacroLitArray{}
	for TokenNotIn(p.tok, token.RPAREN) && !isMacroArgEnd(inMacro, p.tok) {
		lit := p.parseMacroArgumentLit(inMacro)
		node.Append(lit)
		if p.tok == token.COMMA {
			p.next() // ,
		}
	}
	return nilIfEmpty(node)
}

// 参数字面量
// macro_param_lit =
//    < macro_item > .
func (p *Parser) parseMacroArgumentLit(inMacro bool) (node ast.MacroLiter) {
	list := &ast.MacroLitArray{}
	for TokenNotIn(p.tok, token.COMMA, token.RPAREN) && !isMacroArgEnd(inMacro, p.tok) {
		lit := p.parseMacroArgumentLitItem(inMacro)
		list.Append(lit)
	}
	if len(*list) == 1 {
		return (*list)[0]
	}
	return nilIfEmpty(list)
}

// macro_item =
//    identifier
//    / macro_call_expr
//    / "text not , " .
func (p *Parser) parseMacroArgumentLitItem(inMacro bool) (node ast.MacroLiter) {
	if isIdent(p.tok) {
		return p.parseMacroLitExpr(inMacro)
	}
	return p.parseText()
}

// 是否是参数结尾
func isMacroArgEnd(inMacro bool, tok token.Token) bool {
	if inMacro {
		return isMacroEnd(tok)
	}
	return TokenIn(tok, token.EOF, token.COMMENT, token.MACRO)
}

// 解析字面量
// 宏参数字面量不参与运算
func (p *Parser) parseText() (node ast.MacroLiter) {
	node = &ast.Text{
		Offset: p.pos,
		Kind:   p.tok,
		Text:   p.lit,
	}
	p.next()
	return
}

func (p *Parser) parseComment() ast.Stmt {
	return &ast.Comment{
		Offset: p.pos,
		Kind:   p.tok,
		Text:   p.lit,
	}
}

func (p *Parser) parseIf() ast.Stmt {
	node := &ast.IfStmt{From: p.pos}
	return node
}

// 解析宏表达式
func (p *Parser) ParseExpr() (expr ast.Expr) {
	return p.parseExprPrecedence(token.LowestPrec)
}

// 解析表达式
// 优先级运算解析
func (p *Parser) parseExprPrecedence(prec int) (expr ast.Expr) {
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
func (p *Parser) parseOpExpr(prec int) (expr ast.Expr) {
	if prec >= token.UnaryPrec {
		return p.parseUnaryExpr()
	} else {
		return p.parseExprPrecedence(prec + 1)
	}
}

// 	( ("-" / "~" / "defined" / "!" ) parseTermExpr )
func (p *Parser) parseUnaryExpr() ast.Expr {
	p.skipWhitespace()
	var expr *ast.UnaryExpr
	var last *ast.UnaryExpr
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
		last.X = p.parseTermExpr()
		return expr
	}
	return p.parseTermExpr()
}

// 	"(" expr ")"
//  / numeric_expression
//  / identifier
//  / macro_call_expr.
func (p *Parser) parseTermExpr() (expr ast.Expr) {
	p.skipWhitespace()
	switch p.tok {
	case token.LPAREN:
		lp, _, _ := p.expected(token.LPAREN)
		expr = p.ParseExpr()
		rp, _, _ := p.expected(token.RPAREN)
		return &ast.ParenExpr{
			Lparen: lp,
			X:      expr,
			Rparen: rp,
		}
	case token.INT, token.FLOAT, token.CHAR, token.STRING:
		return p.parseLiteralExpr()
	case token.IDENT:
		return p.parseIdentExpr()
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
func (p *Parser) parseLiteralExpr() (expr ast.Expr) {
	expr = &ast.LitExpr{
		Offset: p.pos,
		Kind:   p.tok,
		Value:  p.lit,
	}
	p.next()
	return
}

// ------------ end expr ----------------- //

// 获取下一个Token
func (p *Parser) next() (pos token.Pos, tok token.Token, lit string) {
	pos, tok, lit = p.pos, p.tok, p.lit
	p.pos, p.tok, p.lit = p.scanner.Scan()
	return
}

// 获取下一个标识符
func (p *Parser) expectedIdent() (pos token.Pos, tok token.Token, lit string) {
	pos, tok, lit = p.pos, p.tok, p.lit
	if isIdent(tok) {
		p.next()
		return
	}
	p.errorf(pos, "expected ident token, got %v", tok)
	return
}

// 是否为标识符
func isIdent(tok token.Token) bool {
	return tok == token.IDENT || (tok.IsKeyword() && tok != token.DEFINED)
}

func (p *Parser) expected(typ ...token.Token) (pos token.Pos, tok token.Token, lit string) {
	pos, tok, lit = p.pos, p.tok, p.lit
	if !TokenIn(tok, typ...) {
		p.errorf(pos, "expected %v token, got %v", typ, tok)
	}
	p.next()
	return
}

func TokenNotIn(tok token.Token, typ ...token.Token) bool {
	return !TokenIn(tok, typ...)
}

// 宏结束符
func isMacroEnd(tok token.Token) bool {
	return TokenIn(tok, token.EOF, token.NEWLINE, token.COMMENT)
}

func TokenIn(tok token.Token, typ ...token.Token) bool {
	for _, tt := range typ {
		if tok == tt {
			return true
		}
	}
	return false
}

// 跳过空白字符
// token.BACKSLASH_NEWLINE, token.BLOCK_COMMENT, token.TEXT
func (p *Parser) skipWhitespace() string {
	t := ""
	for (p.tok == token.TEXT && isEmptyText(p.lit)) || TokenIn(p.tok, token.BACKSLASH_NEWLINE, token.BLOCK_COMMENT) {
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
func isEmptyText(text string) bool {
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
func (p *Parser) scanToMacroEnd(needEmpty bool) (pos token.Pos) {
	for !isMacroEnd(p.tok) {
		cur := p.tok
		if needEmpty && !p.isMacroLineEmpty() && cur != token.BACKSLASH_NEWLINE {
			p.errorf(p.pos, "unexpected token %v", p.tok)
		}
		p.next()
	}
	if p.tok == token.NEWLINE {
		pos = p.pos
		p.next()
		return
	}
	return p.pos
}

// 当前为空
func (p *Parser) isMacroLineEmpty() bool {
	if p.tok == token.BLOCK_COMMENT {
		return true
	}
	if p.tok == token.TEXT && isEmptyText(p.lit) {
		return true
	}
	return false
}

// 尝试找到 token
func (p *Parser) tryNextNotEmpty(tok ...token.Token) (token.Token, bool) {
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
func (p *Parser) nextNotEmptyIs(tok ...token.Token) bool {
	_, ok := p.tryNextNotEmpty(tok...)
	return ok
}

// 检查下一个 token
func (p *Parser) tryNext(tok ...token.Token) (token.Token, bool) {
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
func (p *Parser) nextIs(tok token.Token) bool {
	_, ok := p.tryNext(tok)
	return ok
}

// 位置信息
func (p *Parser) FilePos() token.FilePos {
	return p.scanner.GetFilePos()
}

func (p *Parser) ErrorList() scanner.ErrorList {
	err := scanner.ErrorList{}
	err.Merge(p.errors)
	err.Merge(p.scanner.GetErr())
	err.Sort()
	return err
}

func (p *Parser) error(pos token.Pos, msg string) {
	p.errors.Add(p.scanner.GetFilePos().CreatePosition(pos), msg)
}

func (p *Parser) errorf(pos token.Pos, format string, args ...interface{}) {
	p.error(pos, fmt.Sprintf(format, args...))
}

// 保存状态
func (p *Parser) clone() *Parser {
	return &Parser{
		scanner: p.scanner.CloneWithoutSrc(),
		pos:     p.pos,
		tok:     p.tok,
		lit:     p.lit,
	}
}

// 回复状态
func (p *Parser) reset(state *Parser) {
	errs := p.errors
	src := p.scanner.GetSrc()
	*p = *state
	p.errors = errs
	p.scanner.SetSrc(src)
}

// 解析宏
func Parse(src []byte) (ast.Node, scanner.ErrorList) {
	p := &Parser{}
	p.Init(src)
	return p.Parse(), p.ErrorList()
}

// 解析表达式
func ParseExpr(src []byte, off token.Pos) (ast.Expr, scanner.ErrorList) {
	p := &Parser{}
	p.InitOffset(src, off)
	return p.ParseExpr(), p.ErrorList()
}

// 解析非宏定义区域文本(普通文本/标识符/调用)
func ParseTextStmt(src []byte, off token.Pos) (*ast.MacroLitArray, scanner.ErrorList) {
	p := &Parser{}
	p.InitOffset(src, off)
	return p.ParseTextStmt(), p.ErrorList()
}

func nilIfEmpty(array *ast.MacroLitArray) *ast.MacroLitArray {
	if len(*array) > 0 {
		return array
	}
	return nil
}
