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
	ILLEGAL       Token = iota
	EOF                 // end of file
	BLOCK_COMMENT       // comment
	COMMENT             // comment
	TEXT                // code text

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

	MACRO        //#
	SHARP        //#
	DOUBLE_SHARP //##
	QUOTE        // ’
	DOUBLE_QUOTE // "

	LPAREN    // (
	COMMA     // ,
	RPAREN    // )
	LF        // \n
	BACKSLASH // \
	EQU       // =

	operator_end

	keyword_beg
	INCLUDE
	IF
	IFDEF
	IFNODEF
	ELSE
	ELSEIF
	ENDIF
	UNDEF
	LINE
	ERROR
	DEFINED
	DEFINE
	keyword_end
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",

	EOF:           "EOF",
	BLOCK_COMMENT: "BLOCK_COMMENT",
	COMMENT:       "COMMENT",
	TEXT:          "TEXT",

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
	EQU: "=",
	LSS: "<",
	GTR: ">",

	NEQ:          "!=",
	LEQ:          "<=",
	GEQ:          ">=",
	MACRO:        "#",
	SHARP:        "#",
	DOUBLE_SHARP: "##",
	QUOTE:        "'",
	DOUBLE_QUOTE: "\"",
	BACKSLASH:    "\\",
	LF:           "\\n",
	LPAREN:       "(",
	COMMA:        ",",
	RPAREN:       ")",

	INCLUDE: "#include",
	IF:      "#if",
	IFDEF:   "#ifdef",
	IFNODEF: "#ifndef",
	ELSE:    "#else",
	ELSEIF:  "#elif",
	ENDIF:   "#endif",
	UNDEF:   "#undef",
	LINE:    "#line",
	ERROR:   "#error",
	DEFINED: "defined",
	DEFINE:  "#define",
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
	ILLEGAL:       "ILLEGAL",
	EOF:           "EOF",
	BLOCK_COMMENT: "BLOCK_COMMENT",
	COMMENT:       "COMMENT",
	TEXT:          "TEXT",
	IDENT:         "IDENT",
	INT:           "INT",
	FLOAT:         "FLOAT",
	CHAR:          "CHAR",
	STRING:        "STRING",
	ADD:           "ADD",
	SUB:           "SUB",
	MUL:           "MUL",
	QUO:           "QUO",
	REM:           "REM",
	AND:           "AND",
	OR:            "OR",
	XOR:           "XOR",
	SHL:           "SHL",
	SHR:           "SHR",
	NOT:           "NOT",
	LAND:          "LAND",
	LOR:           "LOR",
	LNOT:          "LNOT",
	EQL:           "EQL",
	LSS:           "LSS",
	GTR:           "GTR",
	NEQ:           "NEQ",
	LEQ:           "LEQ",
	GEQ:           "GEQ",
	MACRO:         "MACRO",
	SHARP:         "SHARP",
	DOUBLE_SHARP:  "DOUBLE_SHARP",
	QUOTE:         "QUOTE",
	DOUBLE_QUOTE:  "DOUBLE_QUOTE",
	BACKSLASH:     "BACKSLASH",
	LF:            "LF",
	EQU:           "EQU",
	LPAREN:        "LPAREN",
	COMMA:         "COMMA",
	RPAREN:        "RPAREN",
	INCLUDE:       "INCLUDE",
	IF:            "IF",
	IFDEF:         "IFDEF",
	IFNODEF:       "IFNODEF",
	UNDEF:         "UNDEF",
	ELSE:          "ELSE",
	ELSEIF:        "ELSEIF",
	ENDIF:         "ENDIF",
	LINE:          "LINE",
	ERROR:         "ERROR",
	DEFINED:       "DEFINED",
	DEFINE:        "DEFINE",
}

type TokenName Token
type Pos int

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

const (
	LowestPrec = 0  // 最低优先级
	UnaryPrec  = 11 // 最高优先级
)

// 优先级
func (op Token) Precedence() int {
	switch op {
	case OR:
		return 1
	case AND:
		return 2
	case LOR:
		return 2
	case XOR:
		return 4
	case LAND:
		return 5
	case EQL, NEQ:
		return 6
	case LSS, LEQ, GTR, GEQ:
		return 7
	case SHL, SHR:
		return 8
	case ADD, SUB:
		return 9
	case MUL, QUO, REM:
		return 10
	}
	return LowestPrec
}
