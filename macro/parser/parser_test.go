package parser

import (
	"dxkite.cn/language/macro/ast"
	"reflect"
	"testing"
)

func TestParse_Include(t *testing.T) {

	tests := []struct {
		name string
		src  []byte
		want ast.Node
	}{
		{
			"include",
			[]byte("#include <stdio.h>\n#include \"log.h\""),
			&ast.BlockStmt{Stmts: []ast.Stmt{
				&ast.IncludeStmt{
					From: 0,
					To:   17,
					Path: "<stdio.h>",
					Type: ast.IncludeInner,
				},
				&ast.IncludeStmt{
					From: 19,
					To:   28,
					Path: "\"log.h\"",
					Type: ast.IncludeOuter,
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Parse(tt.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParse_Error(t *testing.T) {
	code := `#include <stdio.h>
#include "log.h"
# error compile error
#error error 1234 is \ 
defined`
	tests := []struct {
		name string
		src  []byte
		want ast.Node
	}{
		{
			"error",
			[]byte(code),
			&ast.BlockStmt{Stmts: []ast.Stmt{
				&ast.IncludeStmt{
					From: 0,
					To:   17,
					Path: "<stdio.h>",
					Type: ast.IncludeInner,
				},
				&ast.IncludeStmt{
					From: 19,
					To:   28,
					Path: "\"log.h\"",
					Type: ast.IncludeOuter,
				},
				&ast.ErrorStmt{
					Offset: 36,
					Msg:    "# error compile error",
				},
				&ast.ErrorStmt{
					Offset: 58,
					Msg:    "#error error 1234 is \\ \ndefined",
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Parse(tt.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
