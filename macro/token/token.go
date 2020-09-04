package token

import "strconv"

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
	LF           // \n

	LPAREN // (
	COMMA  // ,
	RPAREN // )
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
