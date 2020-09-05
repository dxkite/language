package scanner

import (
	"dxkite.cn/language/macro/token"
	"fmt"
	"strconv"
	"testing"
)

var code = []byte(`'a' '\''
'''\''-1.6e+10 +0xbbp-4
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
		{0,token.CHAR,"'a'"},{3,token.TEXT," "},{4,token.CHAR,"'\\''"},{8,token.LF,"\n"},
		{9,token.TEXT,"''"},{11,token.CHAR,"'\\''"},{15,token.SUB,"-"},{16,token.FLOAT,"1.6e+10"},{23,token.TEXT," "},{24,token.ADD,"+"},{25,token.FLOAT,"0xbbp-4"},{32,token.LF,"\n"},
		{33,token.FLOAT,"123.45"},{39,token.TEXT," "},{40,token.FLOAT,"0xABC.EF"},{48,token.STRING,"\"dxkite\""},{56,token.TEXT," "},{57,token.STRING,"\"personal.h\""},{69,token.TEXT,"'"},{70,token.INT,"4"},{71,token.TEXT,"\""},{72,token.INT,"56"},{74,token.TEXT,"'"},{75,token.LF,"\n"},
		{76,token.STRING,"\"dxkite\""},{84,token.TEXT," "},{85,token.STRING,"\"personal.h\""},{97,token.LF,"\n"},
		{98,token.INT,"12"},{100,token.TEXT," "},{101,token.QUO,"/"},{102,token.TEXT," "},{103,token.INT,"123"},{106,token.TEXT,"'"},{107,token.INT,"4"},{108,token.TEXT,"\""},{109,token.INT,"56"},{111,token.TEXT,"'"},{112,token.SHARP,"#"},{113,token.IDENT,"abc"},{116,token.DOUBLE_SHARP,"##"},{118,token.INT,"1234"},{122,token.LF,"\n"},
		{123,token.INCLUDE,"#include"},{131,token.TEXT," "},{132,token.LSS,"<"},{133,token.IDENT,"stdio"},{138,token.TEXT,"."},{139,token.IDENT,"h"},{140,token.GTR,">"},{141,token.LF,"\n"},
		{142,token.INCLUDE,"# include"},{151,token.TEXT," "},{152,token.STRING,"\"personal.h\""},{164,token.LF,"\n"},
		{165,token.ERROR,"#  error something error"},{189,token.LF,"\n"},
		{190,token.IDENT,"int"},{193,token.TEXT," "},{194,token.IDENT,"main"},{198,token.LPAREN,"("},{199,token.RPAREN,")"},{200,token.TEXT," \" {      "},{210,token.COMMENT,"// comment"},{220,token.LF,"\n"},
		{221,token.TEXT,"\t"},{222,token.IDENT,"MAX"},{225,token.TEXT," "},{226,token.COMMENT,"/*\nblock comment\n*/"},{245,token.LF,"\n"},
		{246,token.TEXT,"\t"},{247,token.INT,"0b101010"},{255,token.INT,"2345"},{259,token.XOR,"^"},{260,token.INT,"0b1"},{263,token.INT,"24"},{265,token.LF,"\n"},
		{266,token.TEXT,"\t"},{267,token.NOT,"~"},{268,token.INT,"22"},{270,token.LF,"\n"},
		{271,token.TEXT,"\t"},{272,token.COMMENT,"/* some comment */"},{290,token.IDENT,"printf"},{296,token.LPAREN,"("},{297,token.STRING,"\"hello \\\" world\""},{313,token.COMMA,","},{314,token.INT,"12"},{316,token.TEXT," "},{317,token.INT,"342"},{320,token.COMMA,","},{321,token.TEXT," "},{322,token.CHAR,"'\\''"},{326,token.RPAREN,")"},{327,token.TEXT,"; "},{329,token.COMMENT,"/* some comment */"},{347,token.LF,"\n"},
		{348,token.TEXT,"} ''"},{352,token.LF,"\n"},
		{353,token.IDENT,"STR"},{356,token.LPAREN,"("},{357,token.IDENT,"a"},{358,token.TEXT," "},{359,token.INT,"123"},{362,token.TEXT," "},{363,token.INT,"41"},{365,token.TEXT,"'"},{366,token.INT,"2"},{367,token.QUO,"/"},{368,token.INT,"24"},{370,token.TEXT,"\""},{371,token.INT,"51"},{373,token.TEXT," "},{374,token.INT,"12"},{376,token.COMMA,","},{377,token.TEXT," "},{378,token.CHAR,"'b'"},{381,token.RPAREN,")"},{382,token.LF,"\n"},
		{383,token.IF_NO_DEFINE,"#ifndef"},{390,token.TEXT," "},{391,token.IDENT,"A"},{392,token.LF,"\n"},
		{393,token.ERROR,"#error missing a config"},{416,token.LF,"\n"},
		{417,token.ENDIF,"#endif"},{423,token.LF,"\n"},
		{424,token.SHR,">>"},{426,token.GEQ,">="},{428,token.OR,"|"},{429,token.LNOT,"!"},{430,token.TEXT," "},{431,token.LSS,"="},{432,token.TEXT," "},{433,token.NEQ,"!="},{435,token.TEXT," "},{436,token.LEQ,"<="},{438,token.TEXT," "},{439,token.SHL,"<<"},{441,token.TEXT," "},{442,token.SUB,"*"},{443,token.TEXT," "},{444,token.REM,"%"},{445,token.TEXT," "},{446,token.LAND,"&&"},{448,token.TEXT," "},{449,token.AND,"&"},{450,token.TEXT," "},{451,token.LOR,"||"},{453,token.TEXT," "},{454,token.OR,"|"},{455,token.TEXT,"☺"},{458,token.IDENT,"中文"},{464,token.LF,"\n"},
		{465,token.IF,"#if"},{468,token.TEXT," "},{469,token.LNOT,"!"},{470,token.DEFINED,"defined"},{477,token.TEXT," "},{478,token.IDENT,"A"},{479,token.LF,"\n"},
		{480,token.ELSEIF,"#elif"},{485,token.TEXT," "},{486,token.INT,"1f"},{488,token.TEXT," "},{489,token.ADD,"+"},{490,token.TEXT," "},{491,token.IDENT,"A"},{492,token.TEXT,"  "},{494,token.GTR,">"},{495,token.TEXT," "},{496,token.INT,"2020uL"},{502,token.LF,"\n"},
		{503,token.IF_DEFINE,"#ifdef"},{509,token.TEXT," "},{510,token.IDENT,"B"},{511,token.LF,"\n"},
		{512,token.ELSE,"#else"},{517,token.LF,"\n"},
		{518,token.ELSEIF,"#elif"},{523,token.LF,"\n"},
		{524,token.NOP,"# hello"},{531,token.TEXT," "},{532,token.IDENT,"world"},{537,token.LF,"\n"},
		{538,token.LINE,"#line"},{543,token.TEXT," "},{544,token.INT,"30"},{546,token.TEXT," "},{547,token.STRING,"\"test.c\""},{555,token.TEXT," "},{556,token.INT,"01234567"},{564,token.INT,"8"},{565,token.LF,"\n"},
		{566,token.ERROR,"#error b is \\ \n\tdefined"},{589,token.LF,"\n"},
		{590,token.ERROR,"#error b is \\ 12 "},{607,token.LF,"\n"},
		{608,token.TEXT,"\t"},{609,token.DEFINED,"defined"},{616,token.LF,"\n"},
		{617,token.ENDIF,"#endif"},{623,token.EOF,""},
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
		fmt.Printf("=== offset:%v \ttok:token.%-8v, lit:%v\n", gotOffset, token.TokenName(gotTok), strconv.QuoteToGraphic(gotLit))
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
