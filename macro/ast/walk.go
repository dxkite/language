package ast

type Visitor interface {
	Visit(node Node) (w Visitor)
}

// Walk
func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}
	switch n := node.(type) {
	case *BadExpr, *Text, *Ident, *Comment, *IncludeStmt, *LitExpr, *ErrorStmt, *LineStmt:
		v.Visit(n)

	}
}
