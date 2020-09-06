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
	p.root = &ast.BlockStmt{Stmts: []ast.Stmt{}}
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

// define_statement    =
//    ( macro_prefix "#define" define_literal code_line ) .
func (p *parser) parseDefine() ast.Stmt {
	start := p.pos
	node := &ast.DefineStmt{From: p.pos}
	p.next()
	p.skipWhitespace()
	pos, _, name := p.expected(token.IDENT)
	node.Name = &ast.Ident{
		Offset: pos,
		Name:   name,
	}
	node.LitList = []ast.MacroLiter{}
	if p.tok == token.LPAREN {
		p.expected(token.LPAREN)
		list, err := p.parseIdentList()
		if err != nil {
			return &ast.BadExpr{
				Offset: start,
				Token:  token.DEFINE,
				Lit:    p.scanner.Lit(start, p.pos),
			}
		}
		p.expected(token.RPAREN)
		node.IdentList = list
	}
	p.skipWhitespace()
	var prev ast.Expr
	for p.tok != token.EOF && p.tok != token.LF {
		lit := p.parseMacroLit()
		// 合并text节点
		p, pok := prev.(*ast.Text)
		t, tok := lit.(*ast.Text)
		if pok && tok {
			p.Append(t)
		} else {
			node.LitList = append(node.LitList, lit)
			prev = lit
		}
		node.To = lit.End()
	}
	p.scanToMacroEnd(true)
	return node
}

func (p *parser) parseIdentList() (params []*ast.Ident, err error) {
	params = []*ast.Ident{}
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

//
func (p *parser) parseMacroLit() (node ast.MacroLiter) {
	pos, tok, name := p.pos, p.tok, p.lit
	switch tok {
	case token.SHARP:
		p.next()
		offs, typ, lit := p.next()
		if typ == token.IDENT {
			node = &ast.UnaryExpr{
				Offset: pos,
				Op:     token.SHARP,
				X: &ast.Ident{
					Offset: offs,
					Name:   lit,
				},
			}
		} else {
			node = &ast.Text{
				Offset: pos,
				Text:   name + lit,
			}
		}
	case token.IDENT:
		node = p.parseMacroCall()
	default:
		node = &ast.Text{
			Offset: pos,
			Text:   name,
		}
		p.next()
	}
	return
}

func (p *parser) parseMacroCall() (node ast.MacroLiter) {
	name := &ast.Ident{
		Offset: p.pos,
		Name:   p.lit,
	}
	p.expected(token.IDENT)
	p.skipWhitespace()
	if p.tok == token.LPAREN {
		node = &ast.MacroCallExpr {
			From: p.pos, To: name.End(),
			Name: name,
		}
		return
	}
	return name
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

func Parse(src []byte) ast.Node {
	p := &parser{}
	p.init(src)
	p.parse()
	return p.root
}
