package parser

import (
	"bytes"
	"dxkite.cn/language/macro/ast"
	"dxkite.cn/language/macro/token"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseSimple(t *testing.T) {
	tests := []struct {
		name string
		src  []byte
		want ast.Node
	}{
		{
			"parse include",
			[]byte("#include <stdio.h>\n#include \"log.h\""),
			&ast.BlockStmt{
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
			},
		},
		{
			"parse error",
			[]byte("# error compile error\n#error error 1234 is \\\ndefined"),
			&ast.BlockStmt{
				&ast.MacroCmdStmt{
					Offset: 0,
					Kind:   token.ERROR,
					Cmd:    "# error compile error",
				},
				&ast.MacroCmdStmt{
					Offset: 22,
					Kind:   token.ERROR,
					Cmd:    "#error error 1234 is defined",
				},
			},
		},
		{
			"parse undef",
			[]byte("# error compile error\n#undef A"),
			&ast.BlockStmt{
				&ast.MacroCmdStmt{
					Offset: 0,
					Kind:   token.ERROR,
					Cmd:    "# error compile error",
				},
				&ast.UnDefineStmt{
					From: 22, To: 30,
					Name: &ast.Ident{
						Offset: 29,
						Name:   "A",
					},
				},
			},
		},
		{
			"parse line",
			[]byte("#line 99\n#line 98 \"test.c\"\n# 99\n# 1 \"<built-in>\""),
			&ast.BlockStmt{
				&ast.LineStmt{
					From: 0, To: 8,
					Line: "99",
				},
				&ast.LineStmt{
					From: 9, To: 26,
					Line: "98", Path: "\"test.c\"",
				},
				&ast.LineStmt{
					From: 27, To: 31,
					Line: "99",
				},
				&ast.LineStmt{
					From: 32, To: 48,
					Line: "1", Path: "\"<built-in>\"",
				},
			},
		},
		{
			"parse nop",
			[]byte("#line 98 \"test.c\"\n# a b c d e\n#error no op find"),
			&ast.BlockStmt{
				&ast.LineStmt{
					From: 0, To: 17,
					Line: "98", Path: "\"test.c\"",
				},
				&ast.InvalidStmt{
					Offset: 20,
					Text:   "# a b c d e",
				},
				&ast.MacroCmdStmt{
					Offset: 30,
					Kind:   token.ERROR,
					Cmd:    "#error no op find",
				},
			},
		},
		{
			"parse val define",
			[]byte("#define A 123'#bb\n#define B 123#123\n#define C 123'bb"),
			&ast.BlockStmt{
				&ast.ValDefineStmt{
					From: 0, To: 17,
					Name: &ast.Ident{
						Offset: 8,
						Name:   "A",
					},
					Body: &ast.MacroLitArray{
						&ast.Text{
							Offset: 10,
							Kind:   token.INT,
							Text:   "123",
						},
						&ast.Text{
							Offset: 13,
							Kind:   token.QUOTE,
							Text:   "'",
						},
						&ast.Text{
							Offset: 14,
							Kind:   token.SHARP,
							Text:   "#",
						},
						&ast.Ident{
							Offset: 15,
							Name:   "bb",
						},
					},
				},
				&ast.ValDefineStmt{
					From: 18, To: 35,
					Name: &ast.Ident{
						Offset: 26,
						Name:   "B",
					},
					Body: &ast.MacroLitArray{
						&ast.Text{
							Offset: 28,
							Kind:   token.INT,
							Text:   "123",
						},
						&ast.Text{
							Offset: 31,
							Kind:   token.SHARP,
							Text:   "#",
						},
						&ast.Text{
							Offset: 32,
							Kind:   token.INT,
							Text:   "123",
						},
					},
				},
				&ast.ValDefineStmt{
					From: 36, To: 52,
					Name: &ast.Ident{
						Offset: 44,
						Name:   "C",
					},
					Body: &ast.MacroLitArray{
						&ast.Text{
							Offset: 46,
							Kind:   token.INT,
							Text:   "123",
						},
						&ast.Text{
							Offset: 49,
							Kind:   token.QUOTE,
							Text:   "'",
						},
						&ast.Ident{
							Offset: 50,
							Name:   "bb",
						},
					},
				},
			},
		},
		{
			"parse func define",
			[]byte("#define A 123'#bb\n#define B(x, y) x#y\n#define C(x"),
			&ast.BlockStmt{
				&ast.ValDefineStmt{
					From: 0, To: 17,
					Name: &ast.Ident{
						Offset: 8,
						Name:   "A",
					},
					Body: &ast.MacroLitArray{
						&ast.Text{
							Offset: 10,
							Kind:   token.INT,
							Text:   "123",
						},
						&ast.Text{
							Offset: 13,
							Kind:   token.QUOTE,
							Text:   "'",
						},
						&ast.Text{
							Offset: 14,
							Kind:   token.SHARP,
							Text:   "#",
						},
						&ast.Ident{
							Offset: 15,
							Name:   "bb",
						},
					},
				},
				&ast.FuncDefineStmt{
					From: 18, To: 37,
					Name: &ast.Ident{
						Offset: 26,
						Name:   "B",
					},
					Lparen: 27,
					IdentList: []*ast.Ident{
						{
							Offset: 28,
							Name:   "x",
						}, {
							Offset: 31,
							Name:   "y",
						},
					},
					Rparen: 32,
					Body: &ast.MacroLitArray{
						&ast.Ident{
							Offset: 34,
							Name:   "x",
						},
						&ast.UnaryExpr{
							Offset: 35,
							Op:     token.SHARP,
							X: &ast.Ident{
								Offset: 36,
								Name:   "y",
							},
						},
					},
				},
				&ast.BadExpr{
					Offset: 38,
					Token:  token.DEFINE,
					Lit:    "#define C(x",
				},
			},
		},
		{
			"parse define ## macro function",
			[]byte("#define A(x, y) x##y\n#define B(x, y) x#y"),
			&ast.BlockStmt{
				&ast.FuncDefineStmt{
					From: 0, To: 20,
					Name: &ast.Ident{
						Offset: 8,
						Name:   "A",
					},
					Lparen: 9,
					IdentList: []*ast.Ident{
						{
							Offset: 10,
							Name:   "x",
						}, {
							Offset: 13,
							Name:   "y",
						},
					},
					Rparen: 14,
					Body: &ast.MacroLitArray{
						&ast.BinaryExpr{
							X: &ast.Ident{
								Offset: 16,
								Name:   "x",
							},
							Offset: 17,
							Op:     token.DOUBLE_SHARP,
							Y: &ast.Ident{
								Offset: 19,
								Name:   "y",
							},
						},
					},
				},
				&ast.FuncDefineStmt{
					From: 21, To: 40,
					Name: &ast.Ident{
						Offset: 29,
						Name:   "B",
					},
					Lparen: 30,
					IdentList: []*ast.Ident{
						{
							Offset: 31,
							Name:   "x",
						}, {
							Offset: 34,
							Name:   "y",
						},
					},
					Rparen: 35,
					Body: &ast.MacroLitArray{
						&ast.Ident{
							Offset: 37,
							Name:   "x",
						},
						&ast.UnaryExpr{
							Offset: 38,
							Op:     token.SHARP,
							X: &ast.Ident{
								Offset: 39,
								Name:   "y",
							},
						},
					},
				},
			},
		},
		{
			"parse define ## macro function many space",
			[]byte("#define A(  x, y  ) x##y\n#define B(  x, y  ) x   ## y"),
			&ast.BlockStmt{
				&ast.FuncDefineStmt{
					From: 0, To: 24,
					Name: &ast.Ident{
						Offset: 8,
						Name:   "A",
					},
					Lparen: 9,
					IdentList: []*ast.Ident{
						{
							Offset: 12,
							Name:   "x",
						}, {
							Offset: 15,
							Name:   "y",
						},
					},
					Rparen: 18,
					Body: &ast.MacroLitArray{
						&ast.BinaryExpr{
							X: &ast.Ident{
								Offset: 20,
								Name:   "x",
							},
							Offset: 21,
							Op:     token.DOUBLE_SHARP,
							Y: &ast.Ident{
								Offset: 23,
								Name:   "y",
							},
						},
					},
				},
				&ast.FuncDefineStmt{
					From: 25, To: 53,
					Name: &ast.Ident{
						Offset: 33,
						Name:   "B",
					},
					Lparen: 34,
					IdentList: []*ast.Ident{
						{
							Offset: 37,
							Name:   "x",
						}, {
							Offset: 40,
							Name:   "y",
						},
					},
					Rparen: 43,
					Body: &ast.MacroLitArray{
						&ast.BinaryExpr{
							X: &ast.Ident{
								Offset: 45,
								Name:   "x",
							},
							Offset: 49,
							Op:     token.DOUBLE_SHARP,
							Y: &ast.Ident{
								Offset: 52,
								Name:   "y",
							},
						},
					},
				},
			},
		},
		{
			"parse define single A",
			[]byte("#define A \n#define B(x,y) x##y"),
			&ast.BlockStmt{
				&ast.ValDefineStmt{
					From: 0, To: 10,
					Name: &ast.Ident{
						Offset: 8,
						Name:   "A",
					},
				},
				&ast.FuncDefineStmt{
					From: 11, To: 30,
					Name: &ast.Ident{
						Offset: 19,
						Name:   "B",
					},
					Lparen: 20,
					IdentList: []*ast.Ident{
						{
							Offset: 21,
							Name:   "x",
						}, {
							Offset: 23,
							Name:   "y",
						},
					},
					Rparen: 24,
					Body: &ast.MacroLitArray{
						&ast.BinaryExpr{
							X: &ast.Ident{
								Offset: 26,
								Name:   "x",
							},
							Offset: 27,
							Op:     token.DOUBLE_SHARP,
							Y: &ast.Ident{
								Offset: 29,
								Name:   "y",
							},
						},
					},
				},
			},
		},
		{
			"parse define single C call",
			[]byte("#define A \n#define B(x,y) C(x)x##y"),
			&ast.BlockStmt{
				&ast.ValDefineStmt{
					From: 0, To: 10,
					Name: &ast.Ident{
						Offset: 8,
						Name:   "A",
					},
				},
				&ast.FuncDefineStmt{
					From: 11, To: 34,
					Name: &ast.Ident{
						Offset: 19,
						Name:   "B",
					},
					Lparen: 20,
					IdentList: []*ast.Ident{
						{
							Offset: 21,
							Name:   "x",
						}, {
							Offset: 23,
							Name:   "y",
						},
					},
					Rparen: 24,
					Body: &ast.MacroLitArray{
						&ast.MacroCallExpr{
							From: 26, To: 30,
							Name: &ast.Ident{
								Offset: 26,
								Name:   "C",
							},
							Lparen: 27,
							ParamList: &ast.MacroLitArray{
								&ast.Ident{
									Offset: 28,
									Name:   "x",
								},
							},
							Rparen: 29,
						},
						&ast.BinaryExpr{
							X: &ast.Ident{
								Offset: 30,
								Name:   "x",
							},
							Offset: 31,
							Op:     token.DOUBLE_SHARP,
							Y: &ast.Ident{
								Offset: 33,
								Name:   "y",
							},
						},
					},
				},
			},
		},
		{
			"parse define single C call space",
			[]byte("#define A \n#define B(x,y) C( x 123)x##y"),
			&ast.BlockStmt{
				&ast.ValDefineStmt{
					From: 0, To: 10,
					Name: &ast.Ident{
						Offset: 8,
						Name:   "A",
					},
				},
				&ast.FuncDefineStmt{
					From: 11, To: 39,
					Name: &ast.Ident{
						Offset: 19,
						Name:   "B",
					},
					Lparen: 20,
					IdentList: []*ast.Ident{
						{
							Offset: 21,
							Name:   "x",
						}, {
							Offset: 23,
							Name:   "y",
						},
					},
					Rparen: 24,
					Body: &ast.MacroLitArray{
						&ast.MacroCallExpr{
							From: 26, To: 35,
							Name: &ast.Ident{
								Offset: 26,
								Name:   "C",
							},
							Lparen: 27,
							ParamList: &ast.MacroLitArray{
								&ast.MacroLitArray{
									&ast.Text{
										Offset: 28,
										Kind:   token.TEXT,
										Text:   " ",
									},
									&ast.Ident{
										Offset: 29,
										Name:   "x",
									},
									&ast.Text{
										Offset: 30,
										Kind:   token.TEXT,
										Text:   " ",
									},
									&ast.Text{
										Offset: 31,
										Kind:   token.INT,
										Text:   "123",
									},
								},
							},
							Rparen: 34,
						},
						&ast.BinaryExpr{
							X: &ast.Ident{
								Offset: 35,
								Name:   "x",
							},
							Offset: 36,
							Op:     token.DOUBLE_SHARP,
							Y: &ast.Ident{
								Offset: 38,
								Name:   "y",
							},
						},
					},
				},
			},
		},
		{
			"parse define call nested",
			[]byte("#define A \n#define B(x,y) C(D(x) 123,E()F(g)))x##y"),
			&ast.BlockStmt{
				&ast.ValDefineStmt{
					From: 0, To: 10,
					Name: &ast.Ident{
						Offset: 8,
						Name:   "A",
					},
				},
				&ast.FuncDefineStmt{
					From: 11, To: 50,
					Name: &ast.Ident{
						Offset: 19,
						Name:   "B",
					},
					Lparen: 20,
					IdentList: []*ast.Ident{
						{
							Offset: 21,
							Name:   "x",
						}, {
							Offset: 23,
							Name:   "y",
						},
					},
					Rparen: 24,
					Body: &ast.MacroLitArray{
						&ast.MacroCallExpr{
							From: 26, To: 45,
							Name: &ast.Ident{
								Offset: 26,
								Name:   "C",
							},
							Lparen: 27,
							ParamList: &ast.MacroLitArray{
								&ast.MacroLitArray{
									&ast.MacroCallExpr{
										From: 28, To: 32,
										Name: &ast.Ident{
											Offset: 28,
											Name:   "D",
										},
										Lparen: 29,
										ParamList: &ast.MacroLitArray{
											&ast.Ident{
												Offset: 30,
												Name:   "x",
											},
										},
										Rparen: 31,
									},
									&ast.Text{
										Offset: 32,
										Kind:   token.TEXT,
										Text:   " ",
									},
									&ast.Text{
										Offset: 33,
										Kind:   token.INT,
										Text:   "123",
									},
								},
								&ast.MacroLitArray{
									&ast.MacroCallExpr{
										From: 37, To: 40,
										Name: &ast.Ident{
											Offset: 37,
											Name:   "E",
										},
										Lparen: 38,
										Rparen: 39,
									},
									&ast.MacroCallExpr{
										From: 40, To: 44,
										Name: &ast.Ident{
											Offset: 40,
											Name:   "F",
										},
										Lparen: 41,
										ParamList: &ast.MacroLitArray{
											&ast.Ident{
												Offset: 42,
												Name:   "g",
											},
										},
										Rparen: 43,
									},
								},
							},
							Rparen: 44,
						},
						&ast.Text{
							Offset: 45,
							Kind:   token.RPAREN,
							Text:   ")",
						},
						&ast.BinaryExpr{
							X: &ast.Ident{
								Offset: 46,
								Name:   "x",
							},
							Offset: 47,
							Op:     token.DOUBLE_SHARP,
							Y: &ast.Ident{
								Offset: 49,
								Name:   "y",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := Parse(tt.src); !reflect.DeepEqual(got, tt.want) {
				gotS, _ := json.Marshal(got)
				wantS, _ := json.Marshal(tt.want)
				t.Errorf("Parse() = \ngot \t%s\nwant\t%s", gotS, wantS)
			}
		})
	}
}

func Test_clearBackslash(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			"end lf",
			"aabbcc\\\n",
			"aabbcc",
		},
		{
			"mul lf",
			"aab\\\nbcc\\\n",
			"aabbcc",
		},
		{
			"single backslash",
			"aab\\\nbcc\\",
			"aabbcc\\",
		},
		{
			"single backslash end",
			"aab\\bcc\\\nee\\",
			"aab\\bccee\\",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clearBackslash(tt.text); got != tt.want {
				t.Errorf("clearBackslash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parser_parseLiteralExpr(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		wantExpr ast.MacroLiter
	}{
		{
			"number",
			"1",
			&ast.LitExpr{
				Offset: 0,
				Kind:   token.INT,
				Value:  "1",
			},
		},
		{
			"number space",
			"  1",
			&ast.LitExpr{
				Offset: 2,
				Kind:   token.INT,
				Value:  "1",
			},
		},
		{
			"float",
			"  1.1e10",
			&ast.LitExpr{
				Offset: 2,
				Kind:   token.FLOAT,
				Value:  "1.1e10",
			},
		},
		{
			"string",
			`  "1.1e10"`,
			&ast.LitExpr{
				Offset: 2,
				Kind:   token.STRING,
				Value:  `"1.1e10"`,
			},
		},
		{
			"char",
			"  'a'",
			&ast.LitExpr{
				Offset: 2,
				Kind:   token.CHAR,
				Value:  "'a'",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{}
			p.Init([]byte(tt.code))
			if gotExpr := p.parseExpr(); !reflect.DeepEqual(gotExpr, tt.wantExpr) {
				gotS, _ := json.Marshal(gotExpr)
				wantS, _ := json.Marshal(tt.wantExpr)
				t.Errorf("parseExpr() = %v, want %v", string(gotS), string(wantS))
			}
		})
	}
}

func Test_parser_parseUnaryExpr(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		wantExpr ast.MacroLiter
	}{
		{
			"not a",
			"!a",
			&ast.UnaryExpr{
				Offset: 0,
				Op:     token.LNOT,
				X: &ast.Ident{
					Offset: 1,
					Name:   "a",
				},
			},
		},
		{
			"not10",
			"!10",
			&ast.UnaryExpr{
				Offset: 0,
				Op:     token.LNOT,
				X: &ast.LitExpr{
					Offset: 1,
					Kind:   token.INT,
					Value:  "10",
				},
			},
		},
		{
			"not defined A",
			"!defined A",
			&ast.UnaryExpr{
				Offset: 0,
				Op:     token.LNOT,
				X: &ast.UnaryExpr{
					Offset: 1,
					Op:     token.DEFINED,
					X: &ast.Ident{
						Offset: 9,
						Name:   "A",
					},
				},
			},
		},
		{
			"not define",
			"!define",
			&ast.UnaryExpr{
				Offset: 0,
				Op:     token.LNOT,
				X: &ast.Ident{
					Offset: 1,
					Name:   "define",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{}
			p.Init([]byte(tt.code))
			if gotExpr := p.parseExpr(); !reflect.DeepEqual(gotExpr, tt.wantExpr) {
				gotS, _ := json.Marshal(gotExpr)
				wantS, _ := json.Marshal(tt.wantExpr)
				t.Errorf("parseExpr() = \ngot \t%s\nwant\t%s", string(gotS), string(wantS))
			}
		})
	}
}

func Test_parser_parseExpr(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		wantExpr ast.MacroLiter
	}{
		{
			"a*a/a%a+b",
			"a*a/a%a+b",
			&ast.BinaryExpr{
				X: &ast.BinaryExpr{
					X: &ast.BinaryExpr{
						X: &ast.BinaryExpr{
							X: &ast.Ident{
								Offset: 0,
								Name:   "a",
							},
							Offset: 1,
							Op:     token.MUL,
							Y: &ast.Ident{
								Offset: 2,
								Name:   "a",
							},
						},
						Offset: 3,
						Op:     token.QUO,
						Y: &ast.Ident{
							Offset: 4,
							Name:   "a",
						},
					},
					Offset: 5,
					Op:     token.REM,
					Y: &ast.Ident{
						Offset: 6,
						Name:   "a",
					},
				},
				Offset: 7,
				Op:     13,
				Y: &ast.Ident{
					Offset: 8,
					Name:   "b",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotExpr, errs := ParseExpr([]byte(tt.code), 0); !reflect.DeepEqual(gotExpr, tt.wantExpr) {
				for _, err := range errs {
					fmt.Println(err)
				}
				if len(errs) > 0 {
					t.Error(errs)
				}
				gotS, _ := json.Marshal(gotExpr)
				wantS, _ := json.Marshal(tt.wantExpr)
				t.Errorf("parseExpr() = \ngot \t%s\nwant\t%s", string(gotS), string(wantS))
			}
		})
	}
}

func encode(x interface{}) ([]byte, error) {
	b := &bytes.Buffer{}
	e := json.NewEncoder(b)
	if err := e.Encode(x); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func testFile(name, src string, t *testing.T) {
	code, err := ioutil.ReadFile(src)
	if err != nil {
		t.Error(err)
	}
	stmts, errors := Parse(code)
	for _, err := range errors {
		// 194571
		fmt.Println(name, err)
	}
	if len(errors) > 0 {
		_, _ = encode(stmts)
		//fmt.Println(string(a))
		t.Error(errors)
	}
}

func TestParse(t *testing.T) {
	if err := filepath.Walk("testdata/", func(p string, info os.FileInfo, err error) error {
		ext := filepath.Ext(p)
		name := filepath.Base(p)
		if ext == ".c" {
			t.Run(p, func(t *testing.T) {
				testFile(name, p, t)
			})
		}
		return nil
	}); err != nil {
		t.Error(err)
	}
}
