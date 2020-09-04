package scanner

import (
	"dxkite.cn/language/macro/token"
	"fmt"
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
	for {
		gotOffset, gotTok, gotLit := s.Scan()
		fmt.Printf("== get --> \n\toffset:%v\n\ttok:%v\n\tlit:%v\n", gotOffset, gotTok, gotLit)
		if gotTok == token.EOF || gotTok == token.ILLEGAL {
			break
		}
	}
}
