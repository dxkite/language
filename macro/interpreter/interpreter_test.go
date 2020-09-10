package interpreter

import (
	"dxkite.cn/language/macro/ast"
	"dxkite.cn/language/macro/token"
	"fmt"
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
		token.EQL,  // ==
		token.NEQ,  // !=
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
		tn := "token." + token.TokenName(tok).String()
		op := tok.String()
		fmt.Printf(ftpl, tn, op, op, op, op)
	}
	for _, tok := range ti {
		tn := "token." + token.TokenName(tok).String()
		op := tok.String()
		fmt.Printf(itpl, tn, op, op, op, op)
	}
	for _, tok := range tb {
		tn := "token." + token.TokenName(tok).String()
		op := tok.String()
		fmt.Printf(btpl, tn, op, op, op, op)
	}
}
