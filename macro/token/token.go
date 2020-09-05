package token

import (
	"strconv"
	"unicode"
)

// Token is the set of lexical tokens of the macro language
type Token int

// The list of tokens.
const (
	// Special tokens
	ILLEGAL Token = iota
	EOF           // end of file
	COMMENT       // comment
	TEXT          // code text

	literal_beg
	// Identifiers and basic type literals
	// (these tokens stand for classes of literals)
	IDENT  // defined
	INT    // 12345
	FLOAT  // 123.45
	CHAR   // 'a'
	STRING // "abc"
	literal_end

	operator_beg
	// Operators and delimiters
	ADD // +
	SUB // -
	MUL // *
	QUO // /
	REM // %

	AND // &
	OR  // |
	XOR // ^
	SHL // <<
	SHR // >>
	NOT // ~

	LAND // &&
	LOR  // ||
	LNOT // !

	EQL // ==
	LSS // <
	GTR // >

	NEQ // !=
	LEQ // <=
	GEQ // >=

	SHARP        //#
	DOUBLE_SHARP //##

	LPAREN // (
	COMMA  // ,
	RPAREN // )
	LF     // \n

	operator_end

	keyword_beg
	INCLUDE
	IF
	IF_DEFINE
	IF_NO_DEFINE
	ELSE
	ELSEIF
	ENDIF
	LINE
	ERROR
	DEFINED
	NOP
	keyword_end
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",

	EOF:     "EOF",
	COMMENT: "COMMENT",
	TEXT:    "TEXT",

	IDENT:  "IDENT",
	INT:    "INT",
	FLOAT:  "FLOAT",
	CHAR:   "CHAR",
	STRING: "STRING",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	QUO: "/",
	REM: "%",

	AND: "&",
	OR:  "|",
	XOR: "^",
	SHL: "<<",
	SHR: ">>",
	NOT: "~",

	LAND: "&&",
	LOR:  "||",
	LNOT: "!",

	EQL: "==",
	LSS: "<",
	GTR: ">",

	NEQ:          "!=",
	LEQ:          "<=",
	GEQ:          ">=",
	SHARP:        "#",
	DOUBLE_SHARP: "##",
	LF:           "\\n",
	LPAREN:       "(",
	COMMA:        ",",
	RPAREN:       ")",

	INCLUDE:      "#include",
	IF:           "#if",
	IF_DEFINE:    "#ifdef",
	IF_NO_DEFINE: "#ifndef",
	ELSE:         "#else",
	ELSEIF:       "#elif",
	ENDIF:        "#endif",
	LINE:         "#line",
	ERROR:        "#error",
	NOP:          "#",
	DEFINED:      "defined",
}

func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

// 是否是操作符
func (tok Token) IsOperator() bool { return operator_beg < tok && tok < operator_end }

// 是否是关键字
func (tok Token) IsKeyword() bool { return keyword_beg < tok && tok < keyword_end }

var tokenName = [...]string{
	ILLEGAL:      "ILLEGAL",
	EOF:          "EOF",
	COMMENT:      "COMMENT",
	TEXT:         "TEXT",
	IDENT:        "IDENT",
	INT:          "INT",
	FLOAT:        "FLOAT",
	CHAR:         "CHAR",
	STRING:       "STRING",
	ADD:          "ADD",
	SUB:          "SUB",
	MUL:          "MUL",
	QUO:          "QUO",
	REM:          "REM",
	AND:          "AND",
	OR:           "OR",
	XOR:          "XOR",
	SHL:          "SHL",
	SHR:          "SHR",
	NOT:          "NOT",
	LAND:         "LAND",
	LOR:          "LOR",
	LNOT:         "LNOT",
	EQL:          "EQL",
	LSS:          "LSS",
	GTR:          "GTR",
	NEQ:          "NEQ",
	LEQ:          "LEQ",
	GEQ:          "GEQ",
	SHARP:        "SHARP",
	DOUBLE_SHARP: "DOUBLE_SHARP",
	LF:           "LF",
	LPAREN:       "LPAREN",
	COMMA:        "COMMA",
	RPAREN:       "RPAREN",
	INCLUDE:      "INCLUDE",
	IF:           "IF",
	IF_DEFINE:    "IF_DEFINE",
	IF_NO_DEFINE: "IF_NO_DEFINE",
	ELSE:         "ELSE",
	ELSEIF:       "ELSEIF",
	ENDIF:        "ENDIF",
	LINE:         "LINE",
	ERROR:        "ERROR",
	NOP:          "NOP",
	DEFINED:      "DEFINED",
}

type TokenName Token

func (tok TokenName) String() string {
	s := ""
	if 0 <= tok && tok < TokenName(len(tokens)) {
		s = tokenName[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token)
	for i := keyword_beg + 1; i < keyword_end; i++ {
		keywords[tokens[i]] = i
	}
}

// 是否是关键字
func IsKeyword(name string) bool {
	_, ok := keywords[name]
	return ok
}

// 是否是标识符
func IsIdentifier(name string) bool {
	for i, c := range name {
		if !unicode.IsLetter(c) && c != '_' && (i == 0 || !unicode.IsDigit(c)) {
			return false
		}
	}
	return name != "" && !IsKeyword(name)
}
