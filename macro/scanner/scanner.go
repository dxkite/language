package scanner

import (
	"dxkite.cn/language/macro/token"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// 词法扫描
type Scanner struct {
	src          []byte    // source
	ch           rune      // current character
	offset       int       // character offset
	rdOffset     int       // reading offset (position after current character)
	isLineStart  bool      // line start
	line, column int       // position
	Err          ErrorList // error list
}

// Init
func (s *Scanner) Init(src []byte) {
	s.src = src
	s.ch = ' '
	s.offset = 0
	s.rdOffset = 0
	s.line = 1
	s.column = 0
	s.Err.Reset()
	s.next()
	s.isLineStart = true
}

func (s *Scanner) next() {
	s.isLineStart = false
	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset
		if s.ch == '\n' {
			s.isLineStart = true
			s.line++
			s.column = 0
		}
		r, w := rune(s.src[s.rdOffset]), 1
		switch {
		case r == 0:
			s.error(s.offset, "illegal character NUL")
		case r >= utf8.RuneSelf:
			// not ASCII
			r, w = utf8.DecodeRune(s.src[s.rdOffset:])
			if r == utf8.RuneError && w == 1 {
				s.error(s.offset, "illegal UTF-8 encoding")
			}
		}
		s.rdOffset += w
		s.column += w
		s.ch = r
	} else {
		s.offset = len(s.src)
		s.ch = -1 // eof
	}
}

func (s *Scanner) Scan() (offset token.Pos, tok token.Token, lit string) {
	offset = token.Pos(s.offset)
	switch ch := s.ch; {
	case isLetter(ch):
		tok = token.IDENT
		lit = s.scanIdentifier()
		switch lit {
		case "defined":
			tok = token.DEFINED
		case "include":
			tok = token.INCLUDE
		case "if":
			tok = token.IF
		case "ifdef":
			tok = token.IFDEF
		case "ifndef":
			tok = token.IFNODEF
		case "else":
			tok = token.ELSE
		case "elif":
			tok = token.ELSEIF
		case "endif":
			tok = token.ENDIF
		case "undef":
			tok = token.UNDEF
		case "line":
			tok = token.LINE
		case "define":
			tok = token.DEFINE
		case "error":
			tok = token.ERROR
		}
	case isDecimal(ch):
		tok, lit = s.scanNumber()
	case ch == '\'' && s.tryChar():
		c := s.scanChar()
		tok = token.CHAR
		lit = c
	case ch == '"' && s.tryString():
		c := s.scanString()
		tok = token.STRING
		lit = c
	case ch == '#' && s.isLineStart:
		tok = token.MACRO
		lit = "#"
		s.next()
	default:
		s.next()
		switch ch {
		case '/':
			if s.ch == '*' || s.ch == '/' {
				tok, lit = s.scanComment()
				return
			} else {
				tok = token.QUO
				lit = string(ch)
			}
		case '(':
			tok = token.LPAREN
			lit = "("
		case ')':
			tok = token.RPAREN
			lit = ")"
		case ',':
			tok = token.COMMA
			lit = ","
		case '+':
			tok = token.ADD
			lit = string(ch)
		case '-':
			tok = token.SUB
			lit = string(ch)
		case '*':
			tok = token.MUL
			lit = string(ch)
		case '%':
			tok = token.REM
			lit = string(ch)
		case '&':
			if s.ch == '&' {
				s.next()
				tok = token.LAND
				lit = "&&"
			} else {
				tok = token.AND
				lit = string(ch)
			}
		case '|':
			if s.ch == '|' {
				s.next()
				tok = token.LOR
				lit = "||"
			} else {
				tok = token.OR
				lit = string(ch)
			}
		case '=':
			if s.ch == '=' {
				s.next()
				tok = token.EQL
				lit = "=="
			} else {
				tok = token.EQU
				lit = string(ch)
			}
		case '^':
			tok = token.XOR
			lit = string(ch)
		case '~':
			tok = token.NOT
			lit = string(ch)
		case '<':
			if s.ch == '=' {
				s.next()
				tok = token.LEQ
				lit = "<="
			} else if s.ch == '<' {
				s.next()
				tok = token.SHL
				lit = "<<"
			} else {
				tok = token.LSS
				lit = string(ch)
			}
		case '>':
			if s.ch == '=' {
				s.next()
				tok = token.GEQ
				lit = ">="
			} else if s.ch == '>' {
				s.next()
				tok = token.SHR
				lit = ">>"
			} else {
				tok = token.GTR
				lit = string(ch)
			}
		case '!':
			if s.ch == '=' {
				s.next()
				tok = token.NEQ
				lit = "!="
			} else {
				tok = token.LNOT
				lit = string(ch)
			}
		case '\\':
			tok = token.BACKSLASH
			lit = string(ch)
		case '\n':
			tok = token.LF
			lit = string(ch)
		case '#':
			if s.ch == '#' {
				s.next()
				tok = token.DOUBLE_SHARP
				lit = "##"
			} else {
				tok = token.SHARP
				lit = "#"
			}
		case '\'':
			tok = token.QUOTE
			lit = string(ch)
		case '"':
			tok = token.DOUBLE_QUOTE
			lit = string(ch)
		case -1:
			tok = token.EOF
		default:
			for s.isEndOfText() == false {
				s.next()
			}
			tok = token.TEXT
			lit = string(s.src[offset:s.offset])
		}
	}
	return
}

func lower(ch rune) rune     { return ('a' - 'A') | ch } // returns lower-case ch iff ch is ASCII letter
func isDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }

func isHex(ch rune) bool { return '0' <= ch && ch <= '9' || 'a' <= lower(ch) && lower(ch) <= 'f' }
func isLetter(ch rune) bool {
	return 'a' <= lower(ch) && lower(ch) <= 'z' || ch == '_' || ch == '$' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}
func isDigit(ch rune) bool {
	return isDecimal(ch) || ch >= utf8.RuneSelf && unicode.IsDigit(ch)
}

// 扫描字符串
func (s *Scanner) scanString() string {
	offs := s.offset
	s.next()
	for s.ch > 0 && s.ch != '"' {
		if s.ch == '\\' && s.peek() == '"' {
			s.next()
			s.next()
		}
		s.next()
	}
	s.next()
	return string(s.src[offs:s.offset])
}

// 获取字面量
func (s *Scanner) Lit(start, stop token.Pos) string {
	return string(s.src[start:stop])
}

// 尝试解析字符串
func (s *Scanner) tryString() bool {
	ss := s.save()
	defer s.reset(ss)
	s.next()
	for s.ch > 0 && s.ch != '"' {
		if s.ch == '\n' {
			break
		}
		if s.ch == '\\' && s.peek() == '"' {
			s.next()
			s.next()
		}
		s.next()
	}
	if s.ch < 0 || s.ch == '\n' {
		return false
	}
	return true
}

// 扫描单个字符
func (s *Scanner) scanChar() string {
	offs := s.offset
	s.next() // '
	n := 0
	for s.ch > 0 && s.ch != '\'' {
		n++
		if s.ch == '\\' && s.peek() == '\'' {
			s.next()
		}
		s.next()
		if n > 1 {
			s.error(s.offset, "error char")
			break
		}
	}
	s.next()
	return string(s.src[offs:s.offset])
}

// 尝试解析字符
func (s *Scanner) tryChar() bool {
	ss := s.save()
	defer s.reset(ss)
	s.next()
	n := 0
	for s.ch > 0 && s.ch != '\'' {
		n++
		if s.ch == '\\' && s.peek() == '\'' {
			s.next()
		}
		s.next()
		if n > 1 {
			break
		}
	}
	if s.ch < 0 || n > 1 || n == 0 {
		return false
	}
	return true
}

// 扫描标识符
func (s *Scanner) scanIdentifier() string {
	offs := s.offset
	for isLetter(s.ch) || isDigit(s.ch) {
		s.next()
	}
	return string(s.src[offs:s.offset])
}

// 扫描代码注释
func (s *Scanner) scanComment() (tok token.Token, lit string) {
	offs := s.offset - 1
	tok = token.COMMENT
	if s.ch == '/' {
		s.next()
		for s.ch != '\n' && s.ch >= 0 {
			s.next()
		}
		goto exit
	}
	if s.ch == '*' {
		tok = token.BLOCK_COMMENT
		s.next()
		for s.ch != '*' && s.peek() != '/' {
			if s.ch < 0 {
				s.error(s.offset, "comment not terminated")
				goto exit
			}
			s.next()
		}
		s.next() // *
		s.next() // /
	}
exit:
	lit = string(s.src[offs:s.offset])
	return
}

// 扫描数字
func (s *Scanner) scanNumber() (tok token.Token, lit string) {
	offs := s.offset
	tok = token.INT
	u := false

	base := 10 // number base

	if s.ch == '0' {
		s.next()
		switch lower(s.ch) {
		case 'x':
			s.next()
			base = 16
		case 'b':
			s.next()
			base = 2
		default:
			base = 8
		}
	}
	s.scanNumberBase(base)
	if s.ch == '.' {
		tok = token.FLOAT
		s.next()
		s.scanNumberBase(base)
	}

	if e := lower(s.ch); e == 'e' || e == 'p' {
		switch {
		case e == 'e' && base != 10 && base != 8:
			s.errorf(s.offset, "%q exponent requires decimal mantissa", s.ch)
		case e == 'p' && base != 16:
			s.errorf(s.offset, "%q exponent requires hexadecimal mantissa", s.ch)
		}
		s.next()
		tok = token.FLOAT
		if s.ch == '+' || s.ch == '-' {
			s.next()
		}
		s.scanNumberBase(10)
	}

	for isLetter(s.ch) || isDecimal(s.ch) {
		if lower(s.ch) == 'u' {
			u = true
		}
		if !u && lower(s.ch) == 'f' {
			tok = token.FLOAT
		}
		s.next()
	}
	lit = string(s.src[offs:s.offset])
	return
}

func (s *Scanner) scanNumberBase(base int) {
	if base <= 10 {
		for isDecimal(s.ch) {
			s.next()
		}
	} else {
		for isHex(s.ch) {
			s.next()
		}
	}
	return
}

func (s *Scanner) isEndOfText() bool {
	if s.ch < 0 || isLetter(s.ch) || isDecimal(s.ch) || strings.Contains("\\/'\"(),+-*%&|=^~<>!\n#", string(s.ch)) {
		return true
	}
	return false
}

func (s *Scanner) peek() byte {
	if s.rdOffset < len(s.src) {
		return s.src[s.rdOffset]
	}
	return 0
}

// 复制一份，除了代码
func (s *Scanner) CloneWithoutSrc() *Scanner {
	return &Scanner{
		ch:          s.ch,
		offset:      s.offset,
		rdOffset:    s.rdOffset,
		isLineStart: s.isLineStart,
		line:        s.line,
		column:      s.column,
		Err:         s.Err,
	}
}

// 获取源码
func (s *Scanner) GetSrc() []byte {
	return s.src
}

func (s *Scanner) SetSrc(src []byte) {
	s.src = src
}

func (s *Scanner) save() *Scanner {
	return &Scanner{
		ch:          s.ch,
		offset:      s.offset,
		rdOffset:    s.rdOffset,
		isLineStart: s.isLineStart,
		line:        s.line,
		column:      s.column,
		Err:         s.Err,
	}
}

func (s *Scanner) reset(state *Scanner) {
	src := s.src
	*s = *state
	s.src = src
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\r' {
		s.next()
	}
}

func (s *Scanner) error(offs int, msg string) {
	p := Position{Offset: offs, Line: s.line, Column: s.column}
	s.Err.Add(p, msg)
}

func (s *Scanner) errorf(offs int, format string, args ...interface{}) {
	s.error(offs, fmt.Sprintf(format, args...))
}
