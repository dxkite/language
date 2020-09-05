package scanner

import (
	"dxkite.cn/language/macro/token"
	"fmt"
	"strconv"
	"testing"
)

var code = []byte(`'\''-1.6e+10 +0xbbp-4
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
		{0, token.CHAR, "'\\''"}, {4, token.SUB, "-"}, {5, token.FLOAT, "1.6e+10"}, {12, token.TEXT, " "}, {13, token.ADD, "+"}, {14, token.FLOAT, "0xbbp-4"}, {21, token.LF, "\n"},
		{22, token.FLOAT, "123.45"}, {28, token.TEXT, " "}, {29, token.FLOAT, "0xABC.EF"}, {37, token.STRING, "\"dxkite\""}, {45, token.TEXT, " "}, {46, token.STRING, "\"personal.h\""}, {58, token.TEXT, "'"}, {59, token.INT, "4"}, {60, token.TEXT, "\""}, {61, token.INT, "56"}, {63, token.TEXT, "'"}, {64, token.LF, "\n"},
		{65, token.STRING, "\"dxkite\""}, {73, token.TEXT, " "}, {74, token.STRING, "\"personal.h\""}, {86, token.LF, "\n"},
		{87, token.INT, "12"}, {89, token.TEXT, " "}, {90, token.QUO, "/"}, {91, token.TEXT, " "}, {92, token.INT, "123"}, {95, token.TEXT, "'"}, {96, token.INT, "4"}, {97, token.TEXT, "\""}, {98, token.INT, "56"}, {100, token.TEXT, "'"}, {101, token.SHARP, "#"}, {102, token.IDENT, "abc"}, {105, token.DOUBLE_SHARP, "##"}, {107, token.INT, "1234"}, {111, token.LF, "\n"},
		{112, token.INCLUDE, "#include"}, {120, token.TEXT, " "}, {121, token.LSS, "<"}, {122, token.IDENT, "stdio"}, {127, token.TEXT, "."}, {128, token.IDENT, "h"}, {129, token.GTR, ">"}, {130, token.LF, "\n"},
		{131, token.INCLUDE, "# include"}, {140, token.TEXT, " "}, {141, token.STRING, "\"personal.h\""}, {153, token.LF, "\n"},
		{154, token.ERROR, "#  error something error"}, {178, token.LF, "\n"},
		{179, token.IDENT, "int"}, {182, token.TEXT, " "}, {183, token.IDENT, "main"}, {187, token.LPAREN, "("}, {188, token.RPAREN, ")"}, {189, token.TEXT, " \" {      "}, {199, token.COMMENT, "// comment"}, {209, token.LF, "\n"},
		{210, token.TEXT, "\t"}, {211, token.IDENT, "MAX"}, {214, token.TEXT, " "}, {215, token.COMMENT, "/*\nblock comment\n*/"}, {234, token.LF, "\n"},
		{235, token.TEXT, "\t"}, {236, token.INT, "0b101010"}, {244, token.INT, "2345"}, {248, token.XOR, "^"}, {249, token.INT, "0b1"}, {252, token.INT, "24"}, {254, token.LF, "\n"},
		{255, token.TEXT, "\t~"}, {257, token.INT, "22"}, {259, token.LF, "\n"},
		{260, token.TEXT, "\t"}, {261, token.COMMENT, "/* some comment */"}, {279, token.IDENT, "printf"}, {285, token.LPAREN, "("}, {286, token.STRING, "\"hello \\\" world\""}, {302, token.COMMA, ","}, {303, token.INT, "12"}, {305, token.TEXT, " "}, {306, token.INT, "342"}, {309, token.COMMA, ","}, {310, token.TEXT, " "}, {311, token.CHAR, "'\\''"}, {315, token.RPAREN, ")"}, {316, token.TEXT, "; "}, {318, token.COMMENT, "/* some comment */"}, {336, token.LF, "\n"},
		{337, token.TEXT, "} "}, {339, token.CHAR, "''"}, {341, token.LF, "\n"},
		{342, token.IDENT, "STR"}, {345, token.LPAREN, "("}, {346, token.IDENT, "a"}, {347, token.TEXT, " "}, {348, token.INT, "123"}, {351, token.TEXT, " "}, {352, token.INT, "41"}, {354, token.TEXT, "'"}, {355, token.INT, "2"}, {356, token.QUO, "/"}, {357, token.INT, "24"}, {359, token.TEXT, "\""}, {360, token.INT, "51"}, {362, token.TEXT, " "}, {363, token.INT, "12"}, {365, token.COMMA, ","}, {366, token.TEXT, " '"}, {368, token.IDENT, "b"}, {369, token.TEXT, "'"}, {370, token.RPAREN, ")"}, {371, token.LF, "\n"},
		{372, token.IF_NO_DEFINE, "#ifndef"}, {379, token.TEXT, " "}, {380, token.IDENT, "A"}, {381, token.LF, "\n"},
		{382, token.ERROR, "#error missing a config"}, {405, token.LF, "\n"},
		{406, token.ENDIF, "#endif"}, {412, token.LF, "\n"},
		{413, token.SHR, ">>"}, {415, token.GEQ, ">="}, {417, token.OR, "|"}, {418, token.LNOT, "!"}, {419, token.TEXT, " "}, {420, token.LSS, "="}, {421, token.TEXT, " "}, {422, token.NEQ, "!="}, {424, token.TEXT, " "}, {425, token.LEQ, "<="}, {427, token.TEXT, " "}, {428, token.SHL, "<<"}, {430, token.TEXT, " "}, {431, token.SUB, "*"}, {432, token.TEXT, " "}, {433, token.REM, "%"}, {434, token.TEXT, " "}, {435, token.LAND, "&&"}, {437, token.TEXT, " "}, {438, token.AND, "&"}, {439, token.TEXT, " "}, {440, token.LOR, "||"}, {442, token.TEXT, " "}, {443, token.OR, "|"}, {444, token.TEXT, "☺"}, {447, token.IDENT, "中文"}, {453, token.LF, "\n"},
		{454, token.IF, "#if"}, {457, token.TEXT, " "}, {458, token.LNOT, "!"}, {459, token.DEFINED, "defined"}, {466, token.TEXT, " "}, {467, token.IDENT, "A"}, {468, token.LF, "\n"},
		{469, token.ELSEIF, "#elif"}, {474, token.TEXT, " "}, {475, token.INT, "1f"}, {477, token.TEXT, " "}, {478, token.ADD, "+"}, {479, token.TEXT, " "}, {480, token.IDENT, "A"}, {481, token.TEXT, "  "}, {483, token.GTR, ">"}, {484, token.TEXT, " "}, {485, token.INT, "2020uL"}, {491, token.LF, "\n"},
		{492, token.IF_DEFINE, "#ifdef"}, {498, token.TEXT, " "}, {499, token.IDENT, "B"}, {500, token.LF, "\n"},
		{501, token.ELSE, "#else"}, {506, token.LF, "\n"},
		{507, token.ELSEIF, "#elif"}, {512, token.LF, "\n"},
		{513, token.NOP, "# hello"}, {520, token.TEXT, " "}, {521, token.IDENT, "world"}, {526, token.LF, "\n"},
		{527, token.LINE, "#line"}, {532, token.TEXT, " "}, {533, token.INT, "30"}, {535, token.TEXT, " "}, {536, token.STRING, "\"test.c\""}, {544, token.TEXT, " "}, {545, token.INT, "01234567"}, {553, token.INT, "8"}, {554, token.LF, "\n"},
		{555, token.ERROR, "#error b is \\ \n\tdefined"}, {578, token.LF, "\n"},
		{579, token.ERROR, "#error b is \\ 12 "}, {596, token.LF, "\n"},
		{597, token.TEXT, "\t"}, {598, token.DEFINED, "defined"}, {605, token.LF, "\n"},
		{606, token.ENDIF, "#endif"}, {612, token.EOF, ""},
	}

	for _, tt := range tests {
		gotOffset, gotTok, gotLit := s.Scan()
		if gotTok != tt.tok {
			fmt.Printf("=== offset:%v \ttok:token.%-8v, lit:%v\n", gotOffset, gotTok, strconv.QuoteToGraphic(gotLit))
			t.Fatalf("Scan() gotTok = %v, want %v", gotTok, tt.tok)
		}
		if gotLit != tt.lit {
			fmt.Printf("=== offset:%v \ttok:token.%-8v, lit:%v\n", gotOffset, gotTok, strconv.QuoteToGraphic(gotLit))
			t.Fatalf("Scan() gotLit = %v, want %v", gotLit, tt.lit)
		}
	}
}

func TestTokenName(t *testing.T) {
	s := &Scanner{}
	s.Init(code)

	for {
		gotOffset, gotTok, gotLit := s.Scan()
		fmt.Printf("=== offset:%v \ttok:token.%-8v, lit:%v\n", gotOffset, gotTok, strconv.QuoteToGraphic(gotLit))
		if gotTok == token.EOF || gotTok == token.ILLEGAL {
			fmt.Print("\n")
			break
		}
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
}
