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
STR(a 123 41'2/24"51 12, 'b')
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
#undef A` + "#define a \r\r\r\r\n#define b\\\r\r\r\r\n")

func TestScanner_Scan(t *testing.T) {
	s := &Scanner{}
	s.Init(code)
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
		{361, token.IDENT, "STR"}, {364, token.LPAREN, "("}, {365, token.IDENT, "a"}, {366, token.TEXT, " "}, {367, token.INT, "123"}, {370, token.TEXT, " "}, {371, token.INT, "41"}, {373, token.QUOTE, "'"}, {374, token.INT, "2"}, {375, token.QUO, "/"}, {376, token.INT, "24"}, {378, token.DOUBLE_QUOTE, "\""}, {379, token.INT, "51"}, {381, token.TEXT, " "}, {382, token.INT, "12"}, {384, token.COMMA, ","}, {385, token.TEXT, " "}, {386, token.CHAR, "'b'"}, {389, token.RPAREN, ")"}, {390, token.NEWLINE, "\n"},
		{391, token.MACRO, "#"}, {392, token.IFNDEF, "ifndef"}, {398, token.TEXT, " "}, {399, token.IDENT, "A"}, {400, token.NEWLINE, "\n"},
		{401, token.MACRO, "#"}, {402, token.ERROR, "error"}, {407, token.TEXT, " "}, {408, token.IDENT, "missing"}, {415, token.TEXT, " "}, {416, token.IDENT, "a"}, {417, token.TEXT, " "}, {418, token.IDENT, "config"}, {424, token.NEWLINE, "\n"},
		{425, token.MACRO, "#"}, {426, token.ENDIF, "endif"}, {431, token.NEWLINE, "\n"},
		{432, token.SHR, ">>"}, {434, token.GEQ, ">="}, {436, token.OR, "|"}, {437, token.LNOT, "!"}, {438, token.TEXT, " "}, {439, token.EQU, "="}, {440, token.TEXT, " "}, {441, token.NEQ, "!="}, {443, token.TEXT, " "}, {444, token.LEQ, "<="}, {446, token.TEXT, " "}, {447, token.SHL, "<<"}, {449, token.TEXT, " "}, {450, token.MUL, "*"}, {451, token.TEXT, " "}, {452, token.REM, "%"}, {453, token.TEXT, " "}, {454, token.LAND, "&&"}, {456, token.TEXT, " "}, {457, token.AND, "&"}, {458, token.TEXT, " "}, {459, token.LOR, "||"}, {461, token.TEXT, " "}, {462, token.OR, "|"}, {463, token.TEXT, "☺"}, {466, token.IDENT, "中文"}, {472, token.TEXT, " "}, {473, token.EQL, "=="}, {475, token.NEWLINE, "\n"},
		{476, token.MACRO, "#"}, {477, token.IF, "if"}, {479, token.TEXT, " "}, {480, token.LNOT, "!"}, {481, token.DEFINED, "defined"}, {488, token.TEXT, " "}, {489, token.IDENT, "A"}, {490, token.NEWLINE, "\n"},
		{491, token.MACRO, "#"}, {492, token.ELSEIF, "elif"}, {496, token.TEXT, " "}, {497, token.FLOAT, "1f"}, {499, token.TEXT, " "}, {500, token.ADD, "+"}, {501, token.TEXT, " "}, {502, token.IDENT, "A"}, {503, token.TEXT, "  "}, {505, token.GTR, ">"}, {506, token.TEXT, " "}, {507, token.INT, "2020uL"}, {513, token.NEWLINE, "\n"},
		{514, token.MACRO, "#"}, {515, token.IFDEF, "ifdef"}, {520, token.TEXT, " "}, {521, token.IDENT, "B"}, {522, token.NEWLINE, "\n"},
		{523, token.MACRO, "#"}, {524, token.ELSE, "else"}, {528, token.NEWLINE, "\n"},
		{529, token.MACRO, "#"}, {530, token.ELSEIF, "elif"}, {534, token.NEWLINE, "\n"},
		{535, token.MACRO, "#"}, {536, token.DEFINE, "define"}, {542, token.TEXT, " "}, {543, token.IDENT, "$A"}, {545, token.TEXT, " "}, {546, token.INT, "123"}, {549, token.NEWLINE, "\n"},
		{550, token.MACRO, "#"}, {551, token.TEXT, " "}, {552, token.IDENT, "hello"}, {557, token.TEXT, " "}, {558, token.IDENT, "world"}, {563, token.NEWLINE, "\n"},
		{564, token.MACRO, "#"}, {565, token.LINE, "line"}, {569, token.TEXT, " "}, {570, token.INT, "30"}, {572, token.TEXT, " "}, {573, token.STRING, "\"test.c\""}, {581, token.TEXT, " "}, {582, token.INT, "012345678"}, {591, token.NEWLINE, "\n"},
		{592, token.MACRO, "#"}, {593, token.ERROR, "error"}, {598, token.TEXT, " "}, {599, token.IDENT, "b"}, {600, token.TEXT, " "}, {601, token.IDENT, "is"}, {603, token.TEXT, " \\"}, {605, token.NEWLINE, "\n"},
		{606, token.TEXT, "\t"}, {607, token.DEFINED, "defined"}, {614, token.NEWLINE, "\n"},
		{615, token.MACRO, "#"}, {616, token.ERROR, "error"}, {621, token.TEXT, " "}, {622, token.IDENT, "b"}, {623, token.TEXT, " "}, {624, token.IDENT, "is"}, {626, token.TEXT, " \\ "}, {629, token.INT, "12"}, {631, token.TEXT, " "}, {632, token.NEWLINE, "\n"},
		{633, token.TEXT, "\t"}, {634, token.DEFINED, "defined"}, {641, token.NEWLINE, "\n"},
		{642, token.MACRO, "#"}, {643, token.ENDIF, "endif"}, {648, token.NEWLINE, "\n"},
		{649, token.MACRO, "#"}, {650, token.IF, "if"}, {652, token.TEXT, " "}, {653, token.INT, "123BCDE123"}, {663, token.NEWLINE, "\n"},
		{664, token.IDENT, "int"}, {667, token.TEXT, " "}, {668, token.IDENT, "ch"}, {670, token.TEXT, " "}, {671, token.EQU, "="}, {672, token.TEXT, " "}, {673, token.INT, "123"}, {676, token.TEXT, ";"}, {677, token.NEWLINE, "\n"},
		{678, token.MACRO, "#"}, {679, token.ENDIF, "endif"}, {684, token.NEWLINE, "\n"},
		{685, token.MACRO, "#"}, {686, token.DEFINE, "define"}, {692, token.TEXT, " "}, {693, token.IDENT, "$A"}, {695, token.LPAREN, "("}, {696, token.IDENT, "a"}, {697, token.RPAREN, ")"}, {698, token.TEXT, " "}, {699, token.INT, "123"}, {702, token.SHARP, "#"}, {703, token.IDENT, "a"}, {704, token.NEWLINE, "\n"},
		{705, token.MACRO, "#"}, {706, token.DEFINE, "define"}, {712, token.TEXT, " "}, {713, token.IDENT, "B"}, {714, token.LPAREN, "("}, {715, token.IDENT, "v"}, {716, token.RPAREN, ")"}, {717, token.TEXT, " "}, {718, token.IDENT, "$A"}, {720, token.LPAREN, "("}, {721, token.INT, "123v"}, {725, token.TEXT, " "}, {726, token.IDENT, "awd"}, {729, token.RPAREN, ")"}, {730, token.NEWLINE, "\n"},
		{731, token.IDENT, "B"}, {732, token.LPAREN, "("}, {733, token.IDENT, "awd"}, {736, token.RPAREN, ")"}, {737, token.NEWLINE, "\n"},
		{738, token.INT, "123B"}, {742, token.NEWLINE, "\n"},
		{743, token.MACRO, "#"}, {744, token.UNDEF, "undef"}, {749, token.TEXT, " "}, {750, token.IDENT, "A"}, {751, token.EOF, ""},
	}

	errors := ErrorList{
		&Error{token.Position{37, 2, 29}, "'p' exponent requires hexadecimal mantissa"},
	}

	litCode := ""

	for _, tt := range tests {
		gotOffset, gotTok, gotLit := s.Scan()
		if gotTok != tt.tok {
			fmt.Printf("=== offset:%v \ttok:token.%-8s lit:%v\n", gotOffset, token.TokenName(gotTok), strconv.QuoteToGraphic(gotLit))
			t.Fatalf("Scan() gotTok = %v, want %v", gotTok, tt.tok)
		}
		if gotLit != tt.lit {
			fmt.Printf("=== offset:%v \ttok:token.%-8s lit:%v\n", gotOffset, token.TokenName(gotTok), strconv.QuoteToGraphic(gotLit))
			t.Fatalf("Scan() gotLit = %v, want %v", gotLit, tt.lit)
		}
		litCode += gotLit
	}

	if !reflect.DeepEqual(s.Err, errors) {
		t.Fatalf("Scan() Error report failed\n")
	}

	if string(code) != litCode {
		t.Fatalf("Scan() lit code missing\nwant:\n%s\ngot:\n%s\n", string(code), litCode)
	}
}

func TestTokenName(t *testing.T) {
	s := &Scanner{}
	s.Init(code)
	for {
		gotOffset, gotTok, gotLit := s.Scan()
		fmt.Printf("=== offset:%v \ttok:token.%-8v lit:%v\n", gotOffset, token.TokenName(gotTok), strconv.QuoteToGraphic(gotLit))
		if gotTok == token.EOF || gotTok == token.ILLEGAL {
			fmt.Print("\n")
			break
		}
	}

	s.Err.Sort()
	for _, err := range s.Err {
		fmt.Println("error", err)
	}

	s.Init(code)
	for {
		gotOffset, gotTok, gotLit := s.Scan()
		fmt.Printf("{%v,token.%v,%v},", gotOffset, token.TokenName(gotTok), strconv.QuoteToGraphic(gotLit))
		if gotTok == token.NEWLINE {
			fmt.Print("\n")
		}
		if gotTok == token.EOF || gotTok == token.ILLEGAL {
			fmt.Print("\n")
			break
		}
	}

	for _, err := range s.Err {
		fmt.Printf("&Error{Position{%d,%d,%d},%s},\n", err.Pos.Offset, err.Pos.Line, err.Pos.Column, strconv.QuoteToGraphic(err.Msg))
	}
}
