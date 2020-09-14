package scanner

import (
	"dxkite.cn/language/macro/token"
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

var code = []byte(`'a' '\''
'''\''-1.6e+10 +0xbbp-4 0123p13
123.45 0xABC.EF"dxkite" "personal.h"'4"56'
"dxkite" "personal.h"
12 / 123'4"56'#abc##1234
#include <stdio.h>
# include "personal.h"
#  error something error
int main() " {      // comment
	MAX /*
block comment
*/
	0b1010102345^0b124
	~22
	/* some comment */printf("hello \" world",12 342, '\''); /* some comment */
} ''
STR(a 123 41'2/24"51 12, 'b'. '\0', '\100', '\'')
#ifndef A
#error missing a config
#endif
>>>=|! = != <= << * % && & || |☺中文 ==
#if !defined A
#elif 1f + A  > 2020uL
#ifdef B
#else
#elif
#define $A 123
# hello world
#line 30 "test.c" 012345678
#error b is \
	defined
#error b is \ 12 
	defined
#endif
#if 123BCDE123
int ch = 123;
#endif
#define $A(a) 123#a
#define B(v) $A(123v awd)
B(awd)
123B
"a\0a" "\xaabbcc" "\xFFbbcc\t\""
#undef A` + "#define a \r\r\r\r\n#define b\\\r\r\r\r\n" + "#abc //aaa\r\r\r\r\n")

func TestScanner_Scan(t *testing.T) {
	s := &scanner{}
	s.init(code)
	tests := []struct {
		offset int
		tok    token.Token
		lit    string
	}{
		{0, token.CHAR, "'a'"}, {3, token.TEXT, " "}, {4, token.CHAR, "'\\''"}, {8, token.NEWLINE, "\n"},
		{9, token.QUOTE, "'"}, {10, token.QUOTE, "'"}, {11, token.CHAR, "'\\''"}, {15, token.SUB, "-"}, {16, token.FLOAT, "1.6e+10"}, {23, token.TEXT, " "}, {24, token.ADD, "+"}, {25, token.FLOAT, "0xbbp-4"}, {32, token.TEXT, " "}, {33, token.FLOAT, "0123p13"}, {40, token.NEWLINE, "\n"},
		{41, token.FLOAT, "123.45"}, {47, token.TEXT, " "}, {48, token.FLOAT, "0xABC.EF"}, {56, token.STRING, "\"dxkite\""}, {64, token.TEXT, " "}, {65, token.STRING, "\"personal.h\""}, {77, token.QUOTE, "'"}, {78, token.INT, "4"}, {79, token.DOUBLE_QUOTE, "\""}, {80, token.INT, "56"}, {82, token.QUOTE, "'"}, {83, token.NEWLINE, "\n"},
		{84, token.STRING, "\"dxkite\""}, {92, token.TEXT, " "}, {93, token.STRING, "\"personal.h\""}, {105, token.NEWLINE, "\n"},
		{106, token.INT, "12"}, {108, token.TEXT, " "}, {109, token.QUO, "/"}, {110, token.TEXT, " "}, {111, token.INT, "123"}, {114, token.QUOTE, "'"}, {115, token.INT, "4"}, {116, token.DOUBLE_QUOTE, "\""}, {117, token.INT, "56"}, {119, token.QUOTE, "'"}, {120, token.SHARP, "#"}, {121, token.IDENT, "abc"}, {124, token.DOUBLE_SHARP, "##"}, {126, token.INT, "1234"}, {130, token.NEWLINE, "\n"},
		{131, token.MACRO, "#"}, {132, token.INCLUDE, "include"}, {139, token.TEXT, " "}, {140, token.LSS, "<"}, {141, token.IDENT, "stdio"}, {146, token.TEXT, "."}, {147, token.IDENT, "h"}, {148, token.GTR, ">"}, {149, token.NEWLINE, "\n"},
		{150, token.MACRO, "#"}, {151, token.TEXT, " "}, {152, token.INCLUDE, "include"}, {159, token.TEXT, " "}, {160, token.STRING, "\"personal.h\""}, {172, token.NEWLINE, "\n"},
		{173, token.MACRO, "#"}, {174, token.TEXT, "  "}, {176, token.ERROR, "error"}, {181, token.TEXT, " "}, {182, token.IDENT, "something"}, {191, token.TEXT, " "}, {192, token.ERROR, "error"}, {197, token.NEWLINE, "\n"},
		{198, token.IDENT, "int"}, {201, token.TEXT, " "}, {202, token.IDENT, "main"}, {206, token.LPAREN, "("}, {207, token.RPAREN, ")"}, {208, token.TEXT, " "}, {209, token.DOUBLE_QUOTE, "\""}, {210, token.TEXT, " {      "}, {218, token.COMMENT, "// comment"}, {228, token.NEWLINE, "\n"},
		{229, token.TEXT, "\t"}, {230, token.IDENT, "MAX"}, {233, token.TEXT, " "}, {234, token.BLOCK_COMMENT, "/*\nblock comment\n*/"}, {253, token.NEWLINE, "\n"},
		{254, token.TEXT, "\t"}, {255, token.INT, "0b1010102345"}, {267, token.XOR, "^"}, {268, token.INT, "0b124"}, {273, token.NEWLINE, "\n"},
		{274, token.TEXT, "\t"}, {275, token.NOT, "~"}, {276, token.INT, "22"}, {278, token.NEWLINE, "\n"},
		{279, token.TEXT, "\t"}, {280, token.BLOCK_COMMENT, "/* some comment */"}, {298, token.IDENT, "printf"}, {304, token.LPAREN, "("}, {305, token.STRING, "\"hello \\\" world\""}, {321, token.COMMA, ","}, {322, token.INT, "12"}, {324, token.TEXT, " "}, {325, token.INT, "342"}, {328, token.COMMA, ","}, {329, token.TEXT, " "}, {330, token.CHAR, "'\\''"}, {334, token.RPAREN, ")"}, {335, token.TEXT, "; "}, {337, token.BLOCK_COMMENT, "/* some comment */"}, {355, token.NEWLINE, "\n"},
		{356, token.TEXT, "} "}, {358, token.QUOTE, "'"}, {359, token.QUOTE, "'"}, {360, token.NEWLINE, "\n"},
		{361, token.IDENT, "STR"}, {364, token.LPAREN, "("}, {365, token.IDENT, "a"}, {366, token.TEXT, " "}, {367, token.INT, "123"}, {370, token.TEXT, " "}, {371, token.INT, "41"}, {373, token.QUOTE, "'"}, {374, token.INT, "2"}, {375, token.QUO, "/"}, {376, token.INT, "24"}, {378, token.DOUBLE_QUOTE, "\""}, {379, token.INT, "51"}, {381, token.TEXT, " "}, {382, token.INT, "12"}, {384, token.COMMA, ","}, {385, token.TEXT, " "}, {386, token.CHAR, "'b'"}, {389, token.TEXT, ". "}, {391, token.CHAR, "'\\0'"}, {395, token.COMMA, ","}, {396, token.TEXT, " "}, {397, token.CHAR, "'\\100'"}, {403, token.COMMA, ","}, {404, token.TEXT, " "}, {405, token.CHAR, "'\\''"}, {409, token.RPAREN, ")"}, {410, token.NEWLINE, "\n"},
		{411, token.MACRO, "#"}, {412, token.IFNDEF, "ifndef"}, {418, token.TEXT, " "}, {419, token.IDENT, "A"}, {420, token.NEWLINE, "\n"},
		{421, token.MACRO, "#"}, {422, token.ERROR, "error"}, {427, token.TEXT, " "}, {428, token.IDENT, "missing"}, {435, token.TEXT, " "}, {436, token.IDENT, "a"}, {437, token.TEXT, " "}, {438, token.IDENT, "config"}, {444, token.NEWLINE, "\n"},
		{445, token.MACRO, "#"}, {446, token.ENDIF, "endif"}, {451, token.NEWLINE, "\n"},
		{452, token.SHR, ">>"}, {454, token.GEQ, ">="}, {456, token.OR, "|"}, {457, token.LNOT, "!"}, {458, token.TEXT, " "}, {459, token.EQU, "="}, {460, token.TEXT, " "}, {461, token.NEQ, "!="}, {463, token.TEXT, " "}, {464, token.LEQ, "<="}, {466, token.TEXT, " "}, {467, token.SHL, "<<"}, {469, token.TEXT, " "}, {470, token.MUL, "*"}, {471, token.TEXT, " "}, {472, token.REM, "%"}, {473, token.TEXT, " "}, {474, token.LAND, "&&"}, {476, token.TEXT, " "}, {477, token.AND, "&"}, {478, token.TEXT, " "}, {479, token.LOR, "||"}, {481, token.TEXT, " "}, {482, token.OR, "|"}, {483, token.TEXT, "☺"}, {486, token.IDENT, "中文"}, {492, token.TEXT, " "}, {493, token.EQL, "=="}, {495, token.NEWLINE, "\n"},
		{496, token.MACRO, "#"}, {497, token.IF, "if"}, {499, token.TEXT, " "}, {500, token.LNOT, "!"}, {501, token.DEFINED, "defined"}, {508, token.TEXT, " "}, {509, token.IDENT, "A"}, {510, token.NEWLINE, "\n"},
		{511, token.MACRO, "#"}, {512, token.ELSEIF, "elif"}, {516, token.TEXT, " "}, {517, token.FLOAT, "1f"}, {519, token.TEXT, " "}, {520, token.ADD, "+"}, {521, token.TEXT, " "}, {522, token.IDENT, "A"}, {523, token.TEXT, "  "}, {525, token.GTR, ">"}, {526, token.TEXT, " "}, {527, token.INT, "2020uL"}, {533, token.NEWLINE, "\n"},
		{534, token.MACRO, "#"}, {535, token.IFDEF, "ifdef"}, {540, token.TEXT, " "}, {541, token.IDENT, "B"}, {542, token.NEWLINE, "\n"},
		{543, token.MACRO, "#"}, {544, token.ELSE, "else"}, {548, token.NEWLINE, "\n"},
		{549, token.MACRO, "#"}, {550, token.ELSEIF, "elif"}, {554, token.NEWLINE, "\n"},
		{555, token.MACRO, "#"}, {556, token.DEFINE, "define"}, {562, token.TEXT, " "}, {563, token.IDENT, "$A"}, {565, token.TEXT, " "}, {566, token.INT, "123"}, {569, token.NEWLINE, "\n"},
		{570, token.MACRO, "#"}, {571, token.TEXT, " "}, {572, token.IDENT, "hello"}, {577, token.TEXT, " "}, {578, token.IDENT, "world"}, {583, token.NEWLINE, "\n"},
		{584, token.MACRO, "#"}, {585, token.LINE, "line"}, {589, token.TEXT, " "}, {590, token.INT, "30"}, {592, token.TEXT, " "}, {593, token.STRING, "\"test.c\""}, {601, token.TEXT, " "}, {602, token.INT, "012345678"}, {611, token.NEWLINE, "\n"},
		{612, token.MACRO, "#"}, {613, token.ERROR, "error"}, {618, token.TEXT, " "}, {619, token.IDENT, "b"}, {620, token.TEXT, " "}, {621, token.IDENT, "is"}, {623, token.TEXT, " "}, {624, token.BACKSLASH_NEWLINE, "\\\n"}, {626, token.TEXT, "\t"}, {627, token.DEFINED, "defined"}, {634, token.NEWLINE, "\n"},
		{635, token.MACRO, "#"}, {636, token.ERROR, "error"}, {641, token.TEXT, " "}, {642, token.IDENT, "b"}, {643, token.TEXT, " "}, {644, token.IDENT, "is"}, {646, token.TEXT, " \\ "}, {649, token.INT, "12"}, {651, token.TEXT, " "}, {652, token.NEWLINE, "\n"},
		{653, token.TEXT, "\t"}, {654, token.DEFINED, "defined"}, {661, token.NEWLINE, "\n"},
		{662, token.MACRO, "#"}, {663, token.ENDIF, "endif"}, {668, token.NEWLINE, "\n"},
		{669, token.MACRO, "#"}, {670, token.IF, "if"}, {672, token.TEXT, " "}, {673, token.INT, "123BCDE123"}, {683, token.NEWLINE, "\n"},
		{684, token.IDENT, "int"}, {687, token.TEXT, " "}, {688, token.IDENT, "ch"}, {690, token.TEXT, " "}, {691, token.EQU, "="}, {692, token.TEXT, " "}, {693, token.INT, "123"}, {696, token.TEXT, ";"}, {697, token.NEWLINE, "\n"},
		{698, token.MACRO, "#"}, {699, token.ENDIF, "endif"}, {704, token.NEWLINE, "\n"},
		{705, token.MACRO, "#"}, {706, token.DEFINE, "define"}, {712, token.TEXT, " "}, {713, token.IDENT, "$A"}, {715, token.LPAREN, "("}, {716, token.IDENT, "a"}, {717, token.RPAREN, ")"}, {718, token.TEXT, " "}, {719, token.INT, "123"}, {722, token.SHARP, "#"}, {723, token.IDENT, "a"}, {724, token.NEWLINE, "\n"},
		{725, token.MACRO, "#"}, {726, token.DEFINE, "define"}, {732, token.TEXT, " "}, {733, token.IDENT, "B"}, {734, token.LPAREN, "("}, {735, token.IDENT, "v"}, {736, token.RPAREN, ")"}, {737, token.TEXT, " "}, {738, token.IDENT, "$A"}, {740, token.LPAREN, "("}, {741, token.INT, "123v"}, {745, token.TEXT, " "}, {746, token.IDENT, "awd"}, {749, token.RPAREN, ")"}, {750, token.NEWLINE, "\n"},
		{751, token.IDENT, "B"}, {752, token.LPAREN, "("}, {753, token.IDENT, "awd"}, {756, token.RPAREN, ")"}, {757, token.NEWLINE, "\n"},
		{758, token.INT, "123B"}, {762, token.NEWLINE, "\n"},
		{763, token.STRING, "\"a\\0a\""}, {769, token.TEXT, " "}, {770, token.STRING, "\"\\xaabbcc\""}, {780, token.TEXT, " "}, {781, token.STRING, "\"\\xFFbbcc\\t\\\"\""}, {795, token.NEWLINE, "\n"},
		{796, token.MACRO, "#"}, {797, token.UNDEF, "undef"}, {802, token.TEXT, " "}, {803, token.IDENT, "A"}, {804, token.SHARP, "#"}, {805, token.DEFINE, "define"}, {811, token.TEXT, " "}, {812, token.IDENT, "a"}, {813, token.TEXT, " "}, {814, token.NEWLINE, "\r\r\r\r\n"},
		{819, token.MACRO, "#"}, {820, token.DEFINE, "define"}, {826, token.TEXT, " "}, {827, token.IDENT, "b"}, {828, token.BACKSLASH_NEWLINE, "\\\r\r\r\r\n"}, {834, token.MACRO, "#"}, {835, token.IDENT, "abc"}, {838, token.TEXT, " "}, {839, token.COMMENT, "//aaa"}, {844, token.NEWLINE, "\r\r\r\r\n"},
		{849, token.EOF, ""},
	}

	errors := ErrorList{
		&Error{token.Position{37, 2, 28}, "'p' exponent requires hexadecimal mantissa"},
	}

	litCode := ""

	for _, tt := range tests {
		gotOffset, gotTok, gotLit := s.Scan()
		if gotTok != tt.tok {
			fmt.Printf("=== offset:%v \ttok:token.%-8s lit:%v\n", gotOffset, token.Name(gotTok), strconv.QuoteToGraphic(gotLit))
			t.Fatalf("Scan() gotTok = %v, want %v", gotTok, tt.tok)
		}
		if gotLit != tt.lit {
			fmt.Printf("=== offset:%v \ttok:token.%-8s lit:%v\n", gotOffset, token.Name(gotTok), strconv.QuoteToGraphic(gotLit))
			t.Fatalf("Scan() gotLit = %v, want %v", gotLit, tt.lit)
		}
		litCode += gotLit
	}

	if !reflect.DeepEqual(s.err, errors) {
		t.Fatalf("Scan() Error report failed\n")
	}

	if string(code) != litCode {
		t.Fatalf("Scan() lit code missing\nwant:\n%s\ngot:\n%s\n", string(code), litCode)
	}
}

func TestNewOffsetScanner(t *testing.T) {
	s := NewOffsetScanner(code, 100)
	for {
		gotOffset, gotTok, gotLit := s.Scan()
		fmt.Printf("=== offset:%v \ttok:token.%-8v lit:%v\n", gotOffset, token.Name(gotTok), strconv.QuoteToGraphic(gotLit))
		if gotTok == token.EOF || gotTok == token.ILLEGAL {
			fmt.Print("\n")
			break
		}
	}
}

func TestTokenName(t *testing.T) {
	s := &scanner{}
	s.init(code)
	for {
		gotOffset, gotTok, gotLit := s.Scan()
		fmt.Printf("=== offset:%v \ttok:token.%-8v lit:%v\n", gotOffset, token.Name(gotTok), strconv.QuoteToGraphic(gotLit))
		if gotTok == token.EOF || gotTok == token.ILLEGAL {
			fmt.Print("\n")
			break
		}
	}

	s.err.Sort()
	for _, err := range s.err {
		fmt.Println("error", err)
	}

	s.init(code)
	for {
		gotOffset, gotTok, gotLit := s.Scan()
		fmt.Printf("{%v,token.%v,%v},", gotOffset, token.Name(gotTok), strconv.QuoteToGraphic(gotLit))
		if gotTok == token.NEWLINE {
			fmt.Print("\n")
		}
		if gotTok == token.EOF || gotTok == token.ILLEGAL {
			fmt.Print("\n")
			break
		}
	}

	for _, err := range s.err {
		fmt.Printf("&Error{Position{%d,%d,%d},%s},\n", err.Pos.Offset, err.Pos.Line, err.Pos.Column, strconv.QuoteToGraphic(err.Msg))
	}
}
