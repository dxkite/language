package interpreter

import (
	"bytes"
	"dxkite.cn/language/macro/parser"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"
)

//func Test_interpreter_evalLitExpr(t *testing.T) {
//	it := interpreter{}
//	i := it.evalLitExpr(&ast.LitExpr{
//		Offset: 0,
//		Kind:   token.INT,
//		Value:  "0x123ABCDE",
//	})
//	if i != int64(0x123ABCDE) {
//		t.Error("error parse token.INT 0x123ABCEE")
//	}
//	f := it.evalLitExpr(&ast.LitExpr{
//		Offset: 0,
//		Kind:   token.FLOAT,
//		Value:  "0x123ABCEEp10",
//	})
//	if f != 0x123ABCEEp10 {
//		t.Error("error parse token.FLOAT 0x123ABCEE")
//	}
//}

//
//func TestGenBinaryExpr(t *testing.T) {
//	tf := []token.Token{
//		token.ADD,
//		token.MUL,
//		token.SUB,
//		token.QUO,
//		token.LSS, // <
//		token.GTR, // >
//		token.LEQ, // <=
//		token.GEQ, // >=
//		token.EQL, // ==
//		token.NEQ, // !=
//	}
//	//ti := []token.Token{
//	//	token.REM, // %
//	//	token.AND, // &
//	//	token.OR,  // |
//	//	token.XOR, // ^
//	//	token.SHL, // <<
//	//	token.SHR, // >>
//	//}
//	//tb := []token.Token{
//	//	token.LAND, // &&
//	//	token.LOR,  // ||
//	//}
//	ftpl := `case %s:
//		return xx %s yy
//	`
//	for _, tok := range tf {
//		tn := "token." + token.Name(tok).String()
//		op := tok.String()
//		fmt.Printf(ftpl, tn, op)
//	}
//	//for _, tok := range ti {
//	//	tn := "token." + token.Name(tok).String()
//	//	op := tok.String()
//	//	fmt.Printf(itpl, tn, op, op, op, op)
//	//}
//	//for _, tok := range tb {
//	//	tn := "token." + token.Name(tok).String()
//	//	op := tok.String()
//	//	fmt.Printf(btpl, tn, op, op, op, op)
//	//}
//}

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
	it.Eval(stmts, name, p.FilePos())
	pp := path.Join(".", src+".txt")
	if exists(pp) {
		txt, err := ioutil.ReadFile(pp)
		if err != nil {
			t.Error(err)
		}
		if bytes.Equal(txt, it.src.Bytes()) == false {
			t.Fatalf("macro evalStmt error\nwant:\n%s\ngot:\n%s\n",
				strconv.QuoteToGraphic(string(txt)), strconv.QuoteToGraphic(it.src.String()))
		}
	} else {
		fmt.Println("write evalStmt file", pp)
		_ = ioutil.WriteFile(pp, it.src.Bytes(), os.ModePerm)
	}
}

func TestEval(t *testing.T) {
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

func Test_charValue(t *testing.T) {
	tests := []struct {
		ch   string
		want uint8
	}{
		{
			"'A'",
			'A',
		},
		{
			"'\\0'",
			0,
		},
		{
			"'\\xAB'",
			0xAB,
		},
		{
			"'\\301'",
			0301,
		},
		{
			"'\\''",
			'\'',
		},
		{
			"'\\\"'",
			'"',
		},
	}
	for _, tt := range tests {
		t.Run(tt.ch, func(t *testing.T) {
			if got := charValue(tt.ch); got != tt.want {
				t.Errorf("charValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
