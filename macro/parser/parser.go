package parser

import (
	"dxkite.cn/language/macro/ast"
	"dxkite.cn/language/macro/scanner"
	"dxkite.cn/language/macro/token"
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
		case token.DEFINED:
			node = p.parseDefine()
		case token.COMMENT, token.BLOCK_COMMENT:
			node = p.parseComment()
		case token.IF:
			node = p.parseIf()
		default:
			node = &ast.BadToken{Offset: p.pos, Token: p.tok, Lit: p.lit}
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
	p.next()
	node.To += token.Pos(len(node.Line))
	p.skipWhitespace()
	to, tok, path := p.expected(token.STRING, token.LF)
	if tok == token.STRING {
		p.next()
		node.Path = path
		node.To = to + token.Pos(len(path))
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

func (p *parser) parseDefine() ast.Stmt {
	node := &ast.DefineStmt{From: p.pos}
	return node
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
	return
}

func tokenIn(tok token.Token, typ []token.Token) bool {
	for _, tt := range typ {
		if tok == tt {
			return true
		}
	}
	return false
}

func (p *parser) skipWhitespace() {
	for p.tok == token.TEXT && len(strings.TrimSpace(p.lit)) == 0 {
		p.next()
	}
	return
}

func (p *parser) scanToMacroEnd(needEmpty bool) string {
	lit := ""
	for p.tok != token.LF && p.tok != token.EOF {
		isEmptyText := p.tok == token.TEXT && len(strings.TrimSpace(p.lit)) >= 0
		isComment := p.tok == token.COMMENT || p.tok == token.BLOCK_COMMENT
		if needEmpty && !isEmptyText && !isComment {
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
