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

func testFile(name, src, ext string, t *testing.T) {
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
	it := Interpreter{}
	it.Eval(stmts, name, p.FilePos())
	pp := path.Join(".", src+".txt")
	// .c 为正常测试
	// .h 调试中
	if exists(pp) && ext == ".c" {
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
		if ext == ".c" || ext == ".h" {
			t.Run(p, func(t *testing.T) {
				testFile(name, p, ext, t)
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
