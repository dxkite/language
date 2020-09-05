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
			node = &ast.ErrorStmt{
				Offset: p.pos,
				Msg:    p.lit,
			}
			p.next()
			p.scanToMacroEnd()
		}
		if node != nil {
			p.root.Add(node)
		}
	}
}

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
	p.scanToMacroEnd()
	return &ast.IncludeStmt{
		From: from,
		To:   pos,
		Path: path,
		Type: typ,
	}
}

// 获取下一个Token
func (p *parser) next() (pos token.Pos, tok token.Token, lit string) {
	pos, tok, lit = p.peek()
	p.pos, p.tok, p.lit = p.scanner.Scan()
	return
}

// 获取下一个Token
func (p parser) peek() (pos token.Pos, tok token.Token, lit string) {
	pos, tok, lit = p.pos, p.tok, p.lit
	return
}

func (p *parser) skipWhitespace() {
	for p.tok == token.TEXT && len(strings.TrimSpace(p.lit)) == 0 {
		p.next()
	}
	return
}

func (p *parser) scanToMacroEnd() {
	for p.tok != token.LF && p.tok != token.EOF {
		if p.tok != token.TEXT || (p.tok == token.TEXT && len(strings.TrimSpace(p.lit)) >= 0) {
			p.errorf(p.pos, "unexpected token %v after %v", p.tok, token.INCLUDE)
		}
		p.next()
	}
	if p.tok == token.LF {
		p.next()
	}
	return
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
