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
# hello world
#line 30 "test.c" 012345678
#error b is \ 
	defined
#error b is \ 12 
	defined
#endif`)

func TestScanner_Scan(t *testing.T) {
	s := &Scanner{}
	s.Init(code)
	tests := []struct {
		offset int
		tok    token.Token
		lit    string
	}{
		{0, token.CHAR, "'a'"}, {3, token.TEXT, " "}, {4, token.CHAR, "'\\''"}, {8, token.LF, "\n"},
		{9, token.TEXT, "''"}, {11, token.CHAR, "'\\''"}, {15, token.SUB, "-"}, {16, token.FLOAT, "1.6e+10"}, {23, token.TEXT, " "}, {24, token.ADD, "+"}, {25, token.FLOAT, "0xbbp-4"}, {32, token.TEXT, " "}, {33, token.FLOAT, "0123p13"}, {40, token.LF, "\n"},
		{41, token.FLOAT, "123.45"}, {47, token.TEXT, " "}, {48, token.FLOAT, "0xABC.EF"}, {56, token.STRING, "\"dxkite\""}, {64, token.TEXT, " "}, {65, token.STRING, "\"personal.h\""}, {77, token.TEXT, "'"}, {78, token.INT, "4"}, {79, token.TEXT, "\""}, {80, token.INT, "56"}, {82, token.TEXT, "'"}, {83, token.LF, "\n"},
		{84, token.STRING, "\"dxkite\""}, {92, token.TEXT, " "}, {93, token.STRING, "\"personal.h\""}, {105, token.LF, "\n"},
		{106, token.INT, "12"}, {108, token.TEXT, " "}, {109, token.QUO, "/"}, {110, token.TEXT, " "}, {111, token.INT, "123"}, {114, token.TEXT, "'"}, {115, token.INT, "4"}, {116, token.TEXT, "\""}, {117, token.INT, "56"}, {119, token.TEXT, "'"}, {120, token.SHARP, "#"}, {121, token.IDENT, "abc"}, {124, token.DOUBLE_SHARP, "##"}, {126, token.INT, "1234"}, {130, token.LF, "\n"},
		{131, token.INCLUDE, "#include"}, {139, token.TEXT, " "}, {140, token.LSS, "<"}, {141, token.IDENT, "stdio"}, {146, token.TEXT, "."}, {147, token.IDENT, "h"}, {148, token.GTR, ">"}, {149, token.LF, "\n"},
		{150, token.INCLUDE, "# include"}, {159, token.TEXT, " "}, {160, token.STRING, "\"personal.h\""}, {172, token.LF, "\n"},
		{173, token.ERROR, "#  error something error"}, {197, token.LF, "\n"},
		{198, token.IDENT, "int"}, {201, token.TEXT, " "}, {202, token.IDENT, "main"}, {206, token.LPAREN, "("}, {207, token.RPAREN, ")"}, {208, token.TEXT, " \" {      "}, {218, token.COMMENT, "// comment"}, {228, token.LF, "\n"},
		{229, token.TEXT, "\t"}, {230, token.IDENT, "MAX"}, {233, token.TEXT, " "}, {234, token.COMMENT, "/*\nblock comment\n*/"}, {253, token.LF, "\n"},
		{254, token.TEXT, "\t"}, {255, token.INT, "0b101010"}, {263, token.INT, "2345"}, {267, token.XOR, "^"}, {268, token.INT, "0b1"}, {271, token.INT, "24"}, {273, token.LF, "\n"},
		{274, token.TEXT, "\t"}, {275, token.NOT, "~"}, {276, token.INT, "22"}, {278, token.LF, "\n"},
		{279, token.TEXT, "\t"}, {280, token.COMMENT, "/* some comment */"}, {298, token.IDENT, "printf"}, {304, token.LPAREN, "("}, {305, token.STRING, "\"hello \\\" world\""}, {321, token.COMMA, ","}, {322, token.INT, "12"}, {324, token.TEXT, " "}, {325, token.INT, "342"}, {328, token.COMMA, ","}, {329, token.TEXT, " "}, {330, token.CHAR, "'\\''"}, {334, token.RPAREN, ")"}, {335, token.TEXT, "; "}, {337, token.COMMENT, "/* some comment */"}, {355, token.LF, "\n"},
		{356, token.TEXT, "} ''"}, {360, token.LF, "\n"},
		{361, token.IDENT, "STR"}, {364, token.LPAREN, "("}, {365, token.IDENT, "a"}, {366, token.TEXT, " "}, {367, token.INT, "123"}, {370, token.TEXT, " "}, {371, token.INT, "41"}, {373, token.TEXT, "'"}, {374, token.INT, "2"}, {375, token.QUO, "/"}, {376, token.INT, "24"}, {378, token.TEXT, "\""}, {379, token.INT, "51"}, {381, token.TEXT, " "}, {382, token.INT, "12"}, {384, token.COMMA, ","}, {385, token.TEXT, " "}, {386, token.CHAR, "'b'"}, {389, token.RPAREN, ")"}, {390, token.LF, "\n"},
		{391, token.IF_NO_DEFINE, "#ifndef"}, {398, token.TEXT, " "}, {399, token.IDENT, "A"}, {400, token.LF, "\n"},
		{401, token.ERROR, "#error missing a config"}, {424, token.LF, "\n"},
		{425, token.ENDIF, "#endif"}, {431, token.LF, "\n"},
		{432, token.SHR, ">>"}, {434, token.GEQ, ">="}, {436, token.OR, "|"}, {437, token.LNOT, "!"}, {438, token.TEXT, " "}, {439, token.LSS, "="}, {440, token.TEXT, " "}, {441, token.NEQ, "!="}, {443, token.TEXT, " "}, {444, token.LEQ, "<="}, {446, token.TEXT, " "}, {447, token.SHL, "<<"}, {449, token.TEXT, " "}, {450, token.SUB, "*"}, {451, token.TEXT, " "}, {452, token.REM, "%"}, {453, token.TEXT, " "}, {454, token.LAND, "&&"}, {456, token.TEXT, " "}, {457, token.AND, "&"}, {458, token.TEXT, " "}, {459, token.LOR, "||"}, {461, token.TEXT, " "}, {462, token.OR, "|"}, {463, token.TEXT, "☺"}, {466, token.IDENT, "中文"}, {472, token.LF, "\n"},
		{473, token.IF, "#if"}, {476, token.TEXT, " "}, {477, token.LNOT, "!"}, {478, token.DEFINED, "defined"}, {485, token.TEXT, " "}, {486, token.IDENT, "A"}, {487, token.LF, "\n"},
		{488, token.ELSEIF, "#elif"}, {493, token.TEXT, " "}, {494, token.INT, "1f"}, {496, token.TEXT, " "}, {497, token.ADD, "+"}, {498, token.TEXT, " "}, {499, token.IDENT, "A"}, {500, token.TEXT, "  "}, {502, token.GTR, ">"}, {503, token.TEXT, " "}, {504, token.INT, "2020uL"}, {510, token.LF, "\n"},
		{511, token.IF_DEFINE, "#ifdef"}, {517, token.TEXT, " "}, {518, token.IDENT, "B"}, {519, token.LF, "\n"},
		{520, token.ELSE, "#else"}, {525, token.LF, "\n"},
		{526, token.ELSEIF, "#elif"}, {531, token.LF, "\n"},
		{532, token.NOP, "# hello"}, {539, token.TEXT, " "}, {540, token.IDENT, "world"}, {545, token.LF, "\n"},
		{546, token.LINE, "#line"}, {551, token.TEXT, " "}, {552, token.INT, "30"}, {554, token.TEXT, " "}, {555, token.STRING, "\"test.c\""}, {563, token.TEXT, " "}, {564, token.INT, "01234567"}, {572, token.INT, "8"}, {573, token.LF, "\n"},
		{574, token.ERROR, "#error b is \\ \n\tdefined"}, {597, token.LF, "\n"},
		{598, token.ERROR, "#error b is \\ 12 "}, {615, token.LF, "\n"},
		{616, token.TEXT, "\t"}, {617, token.DEFINED, "defined"}, {624, token.LF, "\n"},
		{625, token.ENDIF, "#endif"}, {631, token.EOF, ""},
	}

	errors := ErrorList{
		&Error{Position{37, 2, 29}, "'p' exponent requires hexadecimal mantissa"},
	}

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
	}

	if !reflect.DeepEqual(s.Err, errors) {
		t.Fatalf("Scan() Error report failed")
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
