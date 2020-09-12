package interpreter

import (
	"bytes"
	"dxkite.cn/language/macro/ast"
	"dxkite.cn/language/macro/parser"
	"dxkite.cn/language/macro/token"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"
)

func Test_interpreter_evalLitExpr(t *testing.T) {
	it := interpreter{}
	i := it.evalLitExpr(&ast.LitExpr{
		Offset: 0,
		Kind:   token.INT,
		Value:  "0x123ABCDE",
	})
	if i != int64(0x123ABCDE) {
		t.Error("error parse token.INT 0x123ABCEE")
	}
	f := it.evalLitExpr(&ast.LitExpr{
		Offset: 0,
		Kind:   token.FLOAT,
		Value:  "0x123ABCEEp10",
	})
	if f != 0x123ABCEEp10 {
		t.Error("error parse token.FLOAT 0x123ABCEE")
	}
}

/*
func TestGenBinaryExpr(t *testing.T) {
	tf := []token.Token{
		token.ADD,
		token.MUL,
		token.SUB,
		token.QUO,
		token.LSS, // <
		token.GTR, // >
		token.LEQ, // <=
		token.GEQ, // >=
		token.EQL, // ==
		token.NEQ, // !=
	}
	ti := []token.Token{
		token.REM, // %
		token.AND, // &
		token.OR,  // |
		token.XOR, // ^
		token.SHL, // <<
		token.SHR, // >>
	}
	tb := []token.Token{
		token.LAND, // &&
		token.LOR,  // ||
	}
	ftpl := `case %s:
		if ixo {
			if iyo {
				return ix %s iy
			} else {
				return float64(ix) %s fy
			}
		} else {
			if iyo {
				return fx %s float64(iy)
			} else {
				return fx %s fy
			}
		}
	`
	btpl := `case %s:
		if ixo {
			if iyo {
				return (ix > 0) %s ( iy > 0)
			} else {
				return (ix > 0) %s  (fy >0 )
			}
		} else {
			if iyo {
				return (fx> 0)  %s (iy > 0)
			} else {
				return (fx> 0)  %s (fy >0 )
			}
		}
	`
	itpl := `case %s:
		if ixo {
			if iyo {
				return ix %s iy
			} else {
				return ix %s int64(fy)
			}
		} else {
			if iyo {
				return int64(fx) %s iy
			} else {
				return int64(fx) %s int64(fy)
			}
		}
	`
	for _, tok := range tf {
		tn := "token." + token.Name(tok).String()
		op := tok.String()
		fmt.Printf(ftpl, tn, op, op, op, op)
	}
	for _, tok := range ti {
		tn := "token." + token.Name(tok).String()
		op := tok.String()
		fmt.Printf(itpl, tn, op, op, op, op)
	}
	for _, tok := range tb {
		tn := "token." + token.Name(tok).String()
		op := tok.String()
		fmt.Printf(btpl, tn, op, op, op, op)
	}
}

*/

func exists(p string) bool {
	_, err := os.Stat(p)
	if err != nil {
		if os.IsExist(err) { // 根据错误类型进行判断
			return true
		}
		return false
	}
	return true
}

func testFile(name, src string, t *testing.T) {
	code, err := ioutil.ReadFile(src)
	if err != nil {
		t.Error(err)
	}
	p := parser.Parser{}
	p.Init(code)
	stmts := p.Parse()
	for _, err := range p.ErrorList() {
		// 194571
		fmt.Println(name, err)
	}
	if len(p.ErrorList()) > 0 {
		t.Error(p.ErrorList())
	}
	it := interpreter{}
	it.Eval(stmts, src, p.FilePos())
	pp := path.Join(".", src+".txt")
	if exists(pp) {
		txt, err := ioutil.ReadFile(pp)
		if err != nil {
			t.Error(err)
		}
		if bytes.Equal(txt, it.src.Bytes()) == false {
			t.Fatalf("macro eval error\nwant:\n%s\ngot:\n%s\n",
				strconv.QuoteToGraphic(string(txt)), strconv.QuoteToGraphic(it.src.String()))
		}
	} else {
		fmt.Println("write eval file", pp)
		_ = ioutil.WriteFile(pp, it.src.Bytes(), os.ModePerm)
	}
}

func TestEval(t *testing.T) {
	if err := filepath.Walk("testdata/", func(p string, info os.FileInfo, err error) error {
		ext := filepath.Ext(p)
		name := filepath.Base(p)
		if ext == ".src" {
			t.Run(p, func(t *testing.T) {
				testFile(name, p, t)
			})
		}
		return nil
	}); err != nil {
		t.Error(err)
	}
}
