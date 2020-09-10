package interpreter

import (
	"dxkite.cn/language/macro/ast"
	"dxkite.cn/language/macro/token"
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
