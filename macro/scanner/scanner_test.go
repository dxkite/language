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
>>>=|! = != <= << * % && & || |☺中文
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
123B`)

func TestScanner_Scan(t *testing.T) {
	s := &Scanner{}
	s.Init(code)
	tests := []struct {
		offset int
		tok    token.Token
		lit    string
	}{
		{0, token.CHAR, "'a'"}, {3, token.TEXT, " "}, {4, token.CHAR, "'\\''"}, {8, token.LF, "\n"},
		{9, token.QUOTE, "'"}, {10, token.QUOTE, "'"}, {11, token.CHAR, "'\\''"}, {15, token.SUB, "-"}, {16, token.FLOAT, "1.6e+10"}, {23, token.TEXT, " "}, {24, token.ADD, "+"}, {25, token.FLOAT, "0xbbp-4"}, {32, token.TEXT, " "}, {33, token.FLOAT, "0123p13"}, {40, token.LF, "\n"},
		{41, token.FLOAT, "123.45"}, {47, token.TEXT, " "}, {48, token.FLOAT, "0xABC.EF"}, {56, token.STRING, "\"dxkite\""}, {64, token.TEXT, " "}, {65, token.STRING, "\"personal.h\""}, {77, token.QUOTE, "'"}, {78, token.INT, "4"}, {79, token.DOUBLE_QUOTE, "\""}, {80, token.INT, "56"}, {82, token.QUOTE, "'"}, {83, token.LF, "\n"},
		{84, token.STRING, "\"dxkite\""}, {92, token.TEXT, " "}, {93, token.STRING, "\"personal.h\""}, {105, token.LF, "\n"},
		{106, token.INT, "12"}, {108, token.TEXT, " "}, {109, token.QUO, "/"}, {110, token.TEXT, " "}, {111, token.INT, "123"}, {114, token.QUOTE, "'"}, {115, token.INT, "4"}, {116, token.DOUBLE_QUOTE, "\""}, {117, token.INT, "56"}, {119, token.QUOTE, "'"}, {120, token.SHARP, "#"}, {121, token.IDENT, "abc"}, {124, token.DOUBLE_SHARP, "##"}, {126, token.INT, "1234"}, {130, token.LF, "\n"},
		{131, token.INCLUDE, "#include"}, {139, token.TEXT, " "}, {140, token.LSS, "<"}, {141, token.IDENT, "stdio"}, {146, token.TEXT, "."}, {147, token.IDENT, "h"}, {148, token.GTR, ">"}, {149, token.LF, "\n"},
		{150, token.INCLUDE, "# include"}, {159, token.TEXT, " "}, {160, token.STRING, "\"personal.h\""}, {172, token.LF, "\n"},
		{173, token.ERROR, "#  error something error"}, {197, token.LF, "\n"},
		{198, token.IDENT, "int"}, {201, token.TEXT, " "}, {202, token.IDENT, "main"}, {206, token.LPAREN, "("}, {207, token.RPAREN, ")"}, {208, token.TEXT, " "}, {209, token.DOUBLE_QUOTE, "\""}, {210, token.TEXT, " {      "}, {218, token.COMMENT, "// comment"}, {228, token.LF, "\n"},
		{229, token.TEXT, "\t"}, {230, token.IDENT, "MAX"}, {233, token.TEXT, " "}, {234, token.BLOCK_COMMENT, "/*\nblock comment\n*/"}, {253, token.LF, "\n"},
		{254, token.TEXT, "\t"}, {255, token.INT, "0b1010102345"}, {267, token.XOR, "^"}, {268, token.INT, "0b124"}, {273, token.LF, "\n"},
		{274, token.TEXT, "\t"}, {275, token.NOT, "~"}, {276, token.INT, "22"}, {278, token.LF, "\n"},
		{279, token.TEXT, "\t"}, {280, token.BLOCK_COMMENT, "/* some comment */"}, {298, token.IDENT, "printf"}, {304, token.LPAREN, "("}, {305, token.STRING, "\"hello \\\" world\""}, {321, token.COMMA, ","}, {322, token.INT, "12"}, {324, token.TEXT, " "}, {325, token.INT, "342"}, {328, token.COMMA, ","}, {329, token.TEXT, " "}, {330, token.CHAR, "'\\''"}, {334, token.RPAREN, ")"}, {335, token.TEXT, "; "}, {337, token.BLOCK_COMMENT, "/* some comment */"}, {355, token.LF, "\n"},
		{356, token.TEXT, "} "}, {358, token.QUOTE, "'"}, {359, token.QUOTE, "'"}, {360, token.LF, "\n"},
		{361, token.IDENT, "STR"}, {364, token.LPAREN, "("}, {365, token.IDENT, "a"}, {366, token.TEXT, " "}, {367, token.INT, "123"}, {370, token.TEXT, " "}, {371, token.INT, "41"}, {373, token.QUOTE, "'"}, {374, token.INT, "2"}, {375, token.QUO, "/"}, {376, token.INT, "24"}, {378, token.DOUBLE_QUOTE, "\""}, {379, token.INT, "51"}, {381, token.TEXT, " "}, {382, token.INT, "12"}, {384, token.COMMA, ","}, {385, token.TEXT, " "}, {386, token.CHAR, "'b'"}, {389, token.RPAREN, ")"}, {390, token.LF, "\n"},
		{391, token.IF_NO_DEFINE, "#ifndef"}, {398, token.TEXT, " "}, {399, token.IDENT, "A"}, {400, token.LF, "\n"},
		{401, token.ERROR, "#error missing a config"}, {424, token.LF, "\n"},
		{425, token.ENDIF, "#endif"}, {431, token.LF, "\n"},
		{432, token.SHR, ">>"}, {434, token.GEQ, ">="}, {436, token.OR, "|"}, {437, token.LNOT, "!"}, {438, token.TEXT, " "}, {439, token.LSS, "="}, {440, token.TEXT, " "}, {441, token.NEQ, "!="}, {443, token.TEXT, " "}, {444, token.LEQ, "<="}, {446, token.TEXT, " "}, {447, token.SHL, "<<"}, {449, token.TEXT, " "}, {450, token.SUB, "*"}, {451, token.TEXT, " "}, {452, token.REM, "%"}, {453, token.TEXT, " "}, {454, token.LAND, "&&"}, {456, token.TEXT, " "}, {457, token.AND, "&"}, {458, token.TEXT, " "}, {459, token.LOR, "||"}, {461, token.TEXT, " "}, {462, token.OR, "|"}, {463, token.TEXT, "☺"}, {466, token.IDENT, "中文"}, {472, token.LF, "\n"},
		{473, token.IF, "#if"}, {476, token.TEXT, " "}, {477, token.LNOT, "!"}, {478, token.DEFINED, "defined"}, {485, token.TEXT, " "}, {486, token.IDENT, "A"}, {487, token.LF, "\n"},
		{488, token.ELSEIF, "#elif"}, {493, token.TEXT, " "}, {494, token.FLOAT, "1f"}, {496, token.TEXT, " "}, {497, token.ADD, "+"}, {498, token.TEXT, " "}, {499, token.IDENT, "A"}, {500, token.TEXT, "  "}, {502, token.GTR, ">"}, {503, token.TEXT, " "}, {504, token.INT, "2020uL"}, {510, token.LF, "\n"},
		{511, token.IF_DEFINE, "#ifdef"}, {517, token.TEXT, " "}, {518, token.IDENT, "B"}, {519, token.LF, "\n"},
		{520, token.ELSE, "#else"}, {525, token.LF, "\n"},
		{526, token.ELSEIF, "#elif"}, {531, token.LF, "\n"},
		{532, token.DEFINE, "#define"}, {539, token.TEXT, " "}, {540, token.IDENT, "$A"}, {542, token.TEXT, " "}, {543, token.INT, "123"}, {546, token.LF, "\n"},
		{547, token.NOP, "# hello"}, {554, token.TEXT, " "}, {555, token.IDENT, "world"}, {560, token.LF, "\n"},
		{561, token.LINE, "#line"}, {566, token.TEXT, " "}, {567, token.INT, "30"}, {569, token.TEXT, " "}, {570, token.STRING, "\"test.c\""}, {578, token.TEXT, " "}, {579, token.INT, "012345678"}, {588, token.LF, "\n"},
		{589, token.ERROR, "#error b is \\ \n\tdefined"}, {612, token.LF, "\n"},
		{613, token.ERROR, "#error b is \\ 12 "}, {630, token.LF, "\n"},
		{631, token.TEXT, "\t"}, {632, token.DEFINED, "defined"}, {639, token.LF, "\n"},
		{640, token.ENDIF, "#endif"}, {646, token.LF, "\n"},
		{647, token.IF, "#if"}, {650, token.TEXT, " "}, {651, token.INT, "123BCDE123"}, {661, token.LF, "\n"},
		{662, token.IDENT, "int"}, {665, token.TEXT, " "}, {666, token.IDENT, "ch"}, {668, token.TEXT, " "}, {669, token.LSS, "="}, {670, token.TEXT, " "}, {671, token.INT, "123"}, {674, token.TEXT, ";"}, {675, token.LF, "\n"},
		{676, token.ENDIF, "#endif"}, {682, token.LF, "\n"},
		{683, token.DEFINE, "#define"}, {690, token.TEXT, " "}, {691, token.IDENT, "$A"}, {693, token.LPAREN, "("}, {694, token.IDENT, "a"}, {695, token.RPAREN, ")"}, {696, token.TEXT, " "}, {697, token.INT, "123"}, {700, token.SHARP, "#"}, {701, token.IDENT, "a"}, {702, token.LF, "\n"},
		{703, token.DEFINE, "#define"}, {710, token.TEXT, " "}, {711, token.IDENT, "B"}, {712, token.LPAREN, "("}, {713, token.IDENT, "v"}, {714, token.RPAREN, ")"}, {715, token.TEXT, " "}, {716, token.IDENT, "$A"}, {718, token.LPAREN, "("}, {719, token.INT, "123v"}, {723, token.TEXT, " "}, {724, token.IDENT, "awd"}, {727, token.RPAREN, ")"}, {728, token.LF, "\n"},
		{729, token.IDENT, "B"}, {730, token.LPAREN, "("}, {731, token.IDENT, "awd"}, {734, token.RPAREN, ")"}, {735, token.LF, "\n"},
		{736, token.INT, "123B"}, {740, token.EOF, ""},
	}

	errors := ErrorList{
		&Error{Position{37, 2, 29}, "'p' exponent requires hexadecimal mantissa"},
	}

	litCode := ""

	for _, tt := range tests {
		gotOffset, gotTok, gotLit := s.Scan()
		if gotTok != tt.tok {
			fmt.Printf("=== offset:%v \ttok:token.%-8v lit:%v\n", gotOffset, gotTok, strconv.QuoteToGraphic(gotLit))
			t.Fatalf("Scan() gotTok = %v, want %v", gotTok, tt.tok)
		}
		if gotLit != tt.lit {
			fmt.Printf("=== offset:%v \ttok:token.%-8v lit:%v\n", gotOffset, gotTok, strconv.QuoteToGraphic(gotLit))
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
		if gotTok == token.LF {
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
