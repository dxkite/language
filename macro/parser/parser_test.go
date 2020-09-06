package parser

import (
	"dxkite.cn/language/macro/ast"
	"dxkite.cn/language/macro/token"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		src  []byte
		want ast.Node
	}{
		{
			"parse include",
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
		{
			"parse error",
			[]byte("# error compile error\n#error error 1234 is \\ \ndefined"),
			&ast.BlockStmt{Stmts: []ast.Stmt{
				&ast.ErrorStmt{
					Offset: 0,
					Msg:    "# error compile error",
				},
				&ast.ErrorStmt{
					Offset: 22,
					Msg:    "#error error 1234 is \\ \ndefined",
				},
			}},
		},
		{
			"parse line",
			[]byte("#line 99\n#line 98 \"test.c\""),
			&ast.BlockStmt{Stmts: []ast.Stmt{
				&ast.LineStmt{
					From: 0, To: 8,
					Line: "99",
				},
				&ast.LineStmt{
					From: 9, To: 26,
					Line: "98", Path: "\"test.c\"",
				},
			}},
		},
		{
			"parse nop",
			[]byte("#line 98 \"test.c\"\n# a b c d e\n#error no op find"),
			&ast.BlockStmt{Stmts: []ast.Stmt{
				&ast.LineStmt{
					From: 0, To: 17,
					Line: "98", Path: "\"test.c\"",
				},
				&ast.NopStmt{
					Offset: 18,
					Text:   "# a b c d e",
				},
				&ast.ErrorStmt{
					Offset: 30,
					Msg:    "#error no op find",
				},
			}},
		},
		{
			"parse simple define",
			[]byte("#define A 123'#bb\n#define A 123'bb"),
			&ast.BlockStmt{Stmts: []ast.Stmt{
				&ast.DefineStmt{
					From: 0, To: 17,
					Name: &ast.Ident{
						Offset: 8,
						Name:   "A",
					},
					LitList: []ast.MacroLiter{
						&ast.Text{
							Offset: 10,
							Text:   "123'",
						},
						&ast.UnaryExpr{
							Offset: 14,
							Op:     token.SHARP,
							X: &ast.Ident{
								Offset: 15,
								Name:   "bb",
							},
						},
					},
				},
				&ast.DefineStmt{
					From: 18, To: 34,
					Name: &ast.Ident{
						Offset: 26,
						Name:   "A",
					},
					LitList: []ast.MacroLiter{
						&ast.Text{
							Offset: 28,
							Text:   "123'",
						},
						&ast.Ident{
							Offset: 32,
							Name:   "bb",
						},
					},
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
