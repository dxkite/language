package parser

import (
	"dxkite.cn/language/macro/ast"
	"dxkite.cn/language/macro/scanner"
	"dxkite.cn/language/macro/token"
	"errors"
	"fmt"
	"strings"
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
		switch p.tok {
		case token.INCLUDE:
			node = p.parseInclude()
		case token.ERROR:
			node = p.parseError()
		case token.LINE:
			node = p.parseLine()
		case token.NOP:
			node = p.parseNop()
		case token.DEFINE:
			node = p.parseDefine()
		case token.UNDEF:
			node = p.parseUnDefine()
		case token.COMMENT, token.BLOCK_COMMENT:
			node = p.parseComment()
		case token.IF:
			node = p.parseIf()
		default:
			node = &ast.BadExpr{Offset: p.pos, Token: p.tok, Lit: p.lit}
			p.next()
		}
		if node != nil {
			p.root.Add(node)
		}
	}
}

// #include <path>
// #include "path"
func (p *parser) parseInclude() ast.Stmt {
	from := p.pos
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
func (p *parser) parseError() ast.Stmt {
	node := &ast.ErrorStmt{
		Offset: p.pos,
		Msg:    p.lit,
	}
	p.next()
	p.scanToMacroEnd(true)
	return node
}

// #line number
// #line number "file"
func (p *parser) parseLine() ast.Stmt {
	offs := p.pos
	node := &ast.LineStmt{From: offs}
	p.next()
	p.skipWhitespace()
	node.To, _, node.Line = p.expected(token.INT)
	node.To += token.Pos(len(node.Line))
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
func (p *parser) parseNop() ast.Stmt {
	node := &ast.NopStmt{Offset: p.pos}
	node.Text = p.lit
	p.next()
	node.Text += p.scanToMacroEnd(false)
	return node
}

func (p *parser) parseUnDefine() ast.Stmt {
	start := p.pos
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
		From: start,
		To:   end,
		Name: id,
	}
}

// define_statement    =
//    ( macro_prefix "#define" define_literal code_line ) .
func (p *parser) parseDefine() ast.Stmt {
	start := p.pos
	p.next()
	p.skipWhitespace()
	pos, _, name := p.expected(token.IDENT)
	id := &ast.Ident{
		Offset: pos,
		Name:   name,
	}
	if p.tok == token.LPAREN {
		node := &ast.FuncDefineStmt{From: start, Name: id}
		lp, _, _ := p.expected(token.LPAREN)
		list, err := p.parseIdentList()
		if err != nil {
			return &ast.BadExpr{
				Offset: start,
				Token:  token.DEFINE,
				Lit:    p.scanner.Lit(start, p.pos),
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
		node := &ast.ValDefineStmt{From: start, Name: id}
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
	} else {
		return p.parseMacroCallExprOrIdent()
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
	for p.tok == token.TEXT && len(strings.TrimSpace(p.lit)) == 0 {
		t += p.lit
		p.next()
	}
	return t
}

func (p *parser) scanToMacroEnd(needEmpty bool) string {
	lit := ""
	for !isMacroEnd(p.tok) {
		if needEmpty && !p.curIsComment() && !p.curIsEmpty() {
			p.errorf(p.pos, "unexpected token %v after %v", p.tok, token.INCLUDE)
		}
		lit += p.lit
		p.next()
	}
	if p.tok == token.LF {
		p.next()
	}
	return lit
}

// 当前为空
func (p *parser) curIsEmpty() bool {
	return p.tok == token.TEXT && len(strings.TrimSpace(p.lit)) == 0
}

// 当前为空
func (p *parser) curIsComment() bool {
	return p.tok == token.COMMENT || p.tok == token.BLOCK_COMMENT
}

func (p *parser) error(pos token.Pos, msg string) {
	fmt.Println("error", pos, msg)
}

func (s *parser) errorf(pos token.Pos, format string, args ...interface{}) {
	s.error(pos, fmt.Sprintf(format, args...))
}

// 保存状态
func (s *parser) clone() *parser {
	return &parser{
		scanner: s.scanner.CloneWithoutSrc(),
		errors:  s.errors,
		pos:     s.pos,
		tok:     s.tok,
		lit:     s.lit,
	}
}

// 回复状态
func (s *parser) reset(state *parser) {
	root := s.root
	src := s.scanner.GetSrc()
	*s = *state
	s.root = root
	s.scanner.SetSrc(src)
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
