package ast

import "fmt"

type Visitor interface {
	Visit(node Node) (w Visitor)
}

// Walk
func Walk(v Visitor, node Node) {
	if node == nil {
		return
	}
	if v = v.Visit(node); v == nil {
		return
	}
	switch n := node.(type) {
	case *BadExpr, *Text, *Ident, *Comment:
	case *BlockStmt:
		for _, stmt := range *n {
			Walk(v, stmt)
		}
	case *ValDefineStmt:
		v.Visit(n)
	case *UnDefineStmt:
		v.Visit(n)
	case *FuncDefineStmt:
		v.Visit(n)
		v.Visit(n.Body)
	case *IncludeStmt:
		v.Visit(n)
	case *MacroCallExpr:
		v.Visit(n)
		Walk(v, n.ParamList)
	case *MacroLitArray:
		for _, item := range *n {
			Walk(v, item)
		}
	case *LitExpr:
		// non-visit
	case *MacroCmdStmt, *LineStmt, *InvalidStmt:
		v.Visit(n)
	case *IfStmt:
		v.Visit(n)
		Walk(v, n.Then)
		Walk(v, n.Else)
	case *ElseIfStmt:
		v.Visit(n)
		Walk(v, n.Then)
		Walk(v, n.Else)
	case *IfDefStmt:
		v.Visit(n)
		Walk(v, n.Then)
		Walk(v, n.Else)
	case *IfNoDefStmt:
		v.Visit(n)
		Walk(v, n.Then)
		Walk(v, n.Else)
	case *UnaryExpr:
		Walk(v, n.X)
	case *BinaryExpr:
		Walk(v, n.X)
		Walk(v, n.Y)
	default:
		panic(fmt.Sprintf("ast.Walk: unexpected node type %T", n))
	}
	v.Visit(nil)
}

type inspector func(Node) bool

func (f inspector) Visit(node Node) Visitor {
	if f(node) {
		return f
	}
	return nil
}

func Inspect(node Node, f func(Node) bool) {
	Walk(inspector(f), node)
}
