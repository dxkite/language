package scanner

import (
	"dxkite.cn/language/macro/token"
	"fmt"
	"strconv"
	"testing"
)

func TestScanner_ScanToMacroEnd(t *testing.T) {
	tests := []struct {
		name       string
		code       []byte
		wantOffset int
		wantTok    token.Token
		wantLit    string
	}{
		{
			"test line-comment",
			[]byte("#error something // error"),
			0,
			token.ERROR,
			" something ",
		},
		{
			"test error",
			[]byte("#error something"),
			0,
			token.ERROR,
			" something",
		},
		{
			"test backslash",
			[]byte("#error something \\\n is error"),
			0,
			token.ERROR,
			" something \\\n is error",
		},
		{
			"test cr backslash",
			[]byte("#error something \\\r\r\r\n is error"),
			0,
			token.ERROR,
			" something \\\r\r\r\n is error",
		},
		{
			"test backslash not line end",
			[]byte("#error something \\ is error\nsomeone"),
			0,
			token.ERROR,
			" something \\ is error",
		},
		{
			"test block-comment",
			[]byte("#error something /* error */"),
			0,
			token.ERROR,
			" something ",
		},
		{
			"test nop id block-comment",
			[]byte("# something /* error */"),
			0,
			token.NOP,
			"something ",
		},
		{
			"test nop number block-comment",
			[]byte("# 0 123 435w2 /* error */"),
			0,
			token.NOP,
			"0 123 435w2 ",
		},
		{
			"test endif",
			[]byte("#endif"),
			0,
			token.ENDIF,
			"endif",
		},
		{
			"test endif",
			[]byte("#endif 1"),
			0,
			token.ENDIF,
			"endif",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scanner{}
			s.Init(tt.code)
			gotOffset, gotTok, gotLit := s.Scan()
			if gotOffset != tt.wantOffset {
				t.Errorf("Scan() gotOffset = %v, want %v", gotOffset, tt.wantOffset)
			}
			if gotTok != tt.wantTok {
				t.Errorf("Scan() gotTok = %v, want %v", gotTok, tt.wantTok)
			}
			if gotLit != tt.wantLit {
				t.Errorf("Scan() gotLit = %v, want %v", gotLit, tt.wantLit)
			}
		})
	}
}

func TestScanner_Scan(t *testing.T) {
	s := &Scanner{}
	s.Init([]byte(`-1.6e+10 +0xbbp-4
123.45 0xABC.EF"dxkite" "personal.h"'4"56'
"dxkite" "personal.h"
12 / 123'4"56'#abc##1234
#include <stdio.h>
# include "personal.h"
#  error something error
int main() { // comment
	MAX /*
block comment
*/
	0b1010102345
	/* some comment */printf("hello world",12 342); /* some comment */
}
STR(a 123 41'2/24"51 12, b)
#ifndef A
#error missing a config
#endif
>>>=|!
#if !defined A
#elif 1f + A  > 2020uL
#ifdef B
#error b is defined
#endif`))
	tests := []struct {
		offset int
		tok    token.Token
		lit    string
	}{
		{0, token.SUB, "-"}, {1, token.FLOAT, "1.6e+10"}, {8, token.TEXT, " "}, {9, token.ADD, "+"}, {10, token.FLOAT, "0xbbp-4"}, {17, token.LF, "\n"},
		{18, token.FLOAT, "123.45"}, {24, token.TEXT, " "}, {25, token.FLOAT, "0xABC.EF"}, {33, token.STRING, "\"dxkite\""}, {41, token.TEXT, " "}, {42, token.STRING, "\"personal.h\""}, {54, token.TEXT, "'"}, {55, token.INT, "4"}, {56, token.TEXT, "\""}, {57, token.INT, "56"}, {59, token.TEXT, "'"}, {60, token.LF, "\n"},
		{61, token.STRING, "\"dxkite\""}, {69, token.TEXT, " "}, {70, token.STRING, "\"personal.h\""}, {82, token.LF, "\n"},
		{83, token.INT, "12"}, {85, token.TEXT, " "}, {86, token.QUO, "/"}, {87, token.TEXT, " "}, {88, token.INT, "123"}, {91, token.TEXT, "'"}, {92, token.INT, "4"}, {93, token.TEXT, "\""}, {94, token.INT, "56"}, {96, token.TEXT, "'"}, {97, token.SHARP, "#"}, {98, token.IDENT, "abc"}, {101, token.DOUBLE_SHARP, "##"}, {103, token.INT, "1234"}, {107, token.LF, "\n"},
		{108, token.INCLUDE, "#include"}, {116, token.TEXT, " "}, {117, token.LSS, "<"}, {118, token.IDENT, "stdio"}, {123, token.TEXT, "."}, {124, token.IDENT, "h"}, {125, token.GTR, ">"}, {126, token.LF, "\n"},
		{127, token.INCLUDE, "# include"}, {136, token.TEXT, " "}, {137, token.STRING, "\"personal.h\""}, {149, token.LF, "\n"},
		{150, token.ERROR, "#  error something error"}, {174, token.LF, "\n"},
		{175, token.IDENT, "int"}, {178, token.TEXT, " "}, {179, token.IDENT, "main"}, {183, token.LPAREN, "("}, {184, token.RPAREN, ")"}, {185, token.TEXT, " { "}, {188, token.COMMENT, "// comment"}, {198, token.LF, "\n"},
		{199, token.TEXT, "\t"}, {200, token.IDENT, "MAX"}, {203, token.TEXT, " "}, {204, token.COMMENT, "/*\nblock comment\n*/"}, {223, token.LF, "\n"},
		{224, token.TEXT, "\t"}, {225, token.INT, "0b101010"}, {233, token.INT, "2345"}, {237, token.LF, "\n"},
		{238, token.TEXT, "\t"}, {239, token.COMMENT, "/* some comment */"}, {257, token.IDENT, "printf"}, {263, token.LPAREN, "("}, {264, token.STRING, "\"hello world\""}, {277, token.COMMA, ","}, {278, token.INT, "12"}, {280, token.TEXT, " "}, {281, token.INT, "342"}, {284, token.RPAREN, ")"}, {285, token.TEXT, "; "}, {287, token.COMMENT, "/* some comment */"}, {305, token.LF, "\n"},
		{306, token.TEXT, "}"}, {307, token.LF, "\n"},
		{308, token.IDENT, "STR"}, {311, token.LPAREN, "("}, {312, token.IDENT, "a"}, {313, token.TEXT, " "}, {314, token.INT, "123"}, {317, token.TEXT, " "}, {318, token.INT, "41"}, {320, token.TEXT, "'"}, {321, token.INT, "2"}, {322, token.QUO, "/"}, {323, token.INT, "24"}, {325, token.TEXT, "\""}, {326, token.INT, "51"}, {328, token.TEXT, " "}, {329, token.INT, "12"}, {331, token.COMMA, ","}, {332, token.TEXT, " "}, {333, token.IDENT, "b"}, {334, token.RPAREN, ")"}, {335, token.LF, "\n"},
		{336, token.IF_NO_DEFINE, "#ifndef"}, {343, token.TEXT, " "}, {344, token.IDENT, "A"}, {345, token.LF, "\n"},
		{346, token.ERROR, "#error missing a config"}, {369, token.LF, "\n"},
		{370, token.ENDIF, "#endif"}, {376, token.LF, "\n"},
		{377, token.SHR, ">>"}, {379, token.GEQ, ">="}, {381, token.OR, "|"}, {382, token.LNOT, "!"}, {383, token.LF, "\n"},
		{384, token.IF, "#if"}, {387, token.TEXT, " "}, {388, token.LNOT, "!"}, {389, token.DEFINED, "defined"}, {396, token.TEXT, " "}, {397, token.IDENT, "A"}, {398, token.LF, "\n"},
		{399, token.ELSEIF, "#elif"}, {404, token.TEXT, " "}, {405, token.INT, "1f"}, {407, token.TEXT, " "}, {408, token.ADD, "+"}, {409, token.TEXT, " "}, {410, token.IDENT, "A"}, {411, token.TEXT, "  "}, {413, token.GTR, ">"}, {414, token.TEXT, " "}, {415, token.INT, "2020uL"}, {421, token.LF, "\n"},
		{422, token.IF_DEFINE, "#ifdef"}, {428, token.TEXT, " "}, {429, token.IDENT, "B"}, {430, token.LF, "\n"},
		{431, token.ERROR, "#error b is defined"}, {450, token.LF, "\n"},
		{451, token.ENDIF, "#endif"}, {457, token.EOF, ""},
	}

	//for  {
	//	gotOffset, gotTok, gotLit := s.Scan()
	//	fmt.Printf("\t{%v,token.%v,%v},", gotOffset, gotTok, strconv.QuoteToGraphic(gotLit))
	//	if gotTok == token.LF {
	//		fmt.Print("\n")
	//	}
	//	if gotTok == token.EOF || gotTok == token.ILLEGAL {
	//		fmt.Print("\n")
	//		break
	//	}
	//}

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
