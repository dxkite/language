package ast

import "dxkite.cn/language/macro/token"

// 节点
type Node interface {
	Pos() token.Pos
	End() token.Pos
}

// 表达式
type Expr interface {
	Node
	exprNode()
}

// 语句
type Stmt interface {
	Node
	stmtNode()
}

// 文件包含类型
type IncludeType int

const (
	IncludeInner IncludeType = iota // 系统内
	IncludeOuter                    // 相对
)

type (
	// 标识符
	Ident struct {
		Offset token.Pos // 标识符位置
		Name   string    // 名称
	}
	// 普通文本（宏中的代码块）
	Text struct {
		Offset token.Pos // 标识符位置
		Text   string    // 文本内容
	}
	// 注释
	Comment struct {
		Offset token.Pos // 标识符位置
		Text   string    // 文本内容
	}

	// 语句块
	BlockStmt struct {
		Stmts []Stmt // 定义的语句
	}

	// 宏定义
	DefineStmt struct {
		From, To   token.Pos // 标识符位置
		Name       *Ident    // 定义的标识符
		ParamToken []*Ident  // 定义的参数
		Stmts      []Stmt    // 定义的语句
	}

	// 文件包含语句
	IncludeStmt struct {
		From, To token.Pos   // 标识符位置
		Path     string      // 文件路径
		Type     IncludeType // 文件包含类型
	}

	// 宏调用
	MacroCallExpr struct {
		From, To token.Pos // 标识符位置
		Name     *Ident    // 定义的标识符
		Params   []Stmt    // 调用的参数
	}

	// 数值/字符串字面量
	// token.INT token.FLOAT token.STRING token.CHAR
	LitExpr struct {
		Offset token.Pos // 标识符位置
		Kind   token.Token
		Value  string // 值
	}

	// 报错表达式
	ErrorStmt struct {
		Offset token.Pos // 标识符位置
		Msg    string    // 报错文本内容
	}

	// 无操作宏
	NopStmt struct {
		Offset token.Pos // 标识符位置
		Text   string    // 报错文本内容
	}

	// if语句
	IfStmt struct {
		From, To token.Pos // 标识符位置
		X        Expr      // 条件
		Then     []Stmt    // 正确
		Else     []Stmt    // 错误
	}

	// 一元运算
	UnaryExpr struct {
		Offset token.Pos   // 标识符位置
		Op     token.Token // 操作类型
		X      Expr        // 操作的表达式
	}

	// 二元运算
	BinaryExpr struct {
		X      Expr        // 左值
		Offset token.Pos   // 标识符位置
		Op     token.Token // 操作类型
		Y      Expr        // r右值
	}
)

//------ Node
func (t *Ident) Pos() token.Pos { return t.Offset }
func (t *Ident) End() token.Pos { return token.Pos(int(t.Offset) + len(t.Name)) }
func (*Ident) exprNode()        {}

func (t *Text) Pos() token.Pos { return t.Offset }
func (t *Text) End() token.Pos { return token.Pos(int(t.Offset) + len(t.Text)) }
func (*Text) stmtNode()        {}

// 判断是否为空文本节点
func (t Text) IsEmpty() bool {
	for _, b := range t.Text {
		switch b {
		case ' ', '\t', '\r':
		default:
			return false
		}
	}
	return true
}

func (t *Comment) Pos() token.Pos { return t.Offset }
func (t *Comment) End() token.Pos { return token.Pos(int(t.Offset) + len(t.Text)) }
func (*Comment) stmtNode()        {}

func (t *BlockStmt) Pos() token.Pos {
	if len(t.Stmts) > 0 {
		return t.Stmts[0].Pos()
	}
	return 0
}
func (t *BlockStmt) End() token.Pos {
	l := len(t.Stmts)
	if l > 0 {
		return t.Stmts[l-1].End()
	}
	return 0
}

func (t *BlockStmt) Add(stmt Stmt) {
	t.Stmts = append(t.Stmts, stmt)
}

func (*BlockStmt) stmtNode() {}

func (t *DefineStmt) Pos() token.Pos { return t.From }
func (t *DefineStmt) End() token.Pos { return t.To }
func (*DefineStmt) stmtNode()        {}

// 是否有参数
func (t *DefineStmt) hasParam() bool { return len(t.ParamToken) > 0 }

func (t *IncludeStmt) Pos() token.Pos { return t.From }
func (t *IncludeStmt) End() token.Pos { return t.To }
func (*IncludeStmt) stmtNode()        {}

func (t *MacroCallExpr) Pos() token.Pos { return t.From }
func (t *MacroCallExpr) End() token.Pos { return t.To }
func (*MacroCallExpr) exprNode()        {}

func (t *LitExpr) Pos() token.Pos { return t.Offset }
func (t *LitExpr) End() token.Pos { return token.Pos(int(t.Offset) + len(t.Value)) }
func (*LitExpr) exprNode()        {}

func (t *ErrorStmt) Pos() token.Pos { return t.Offset }
func (t *ErrorStmt) End() token.Pos { return token.Pos(int(t.Offset) + len(t.Msg)) }
func (*ErrorStmt) stmtNode()        {}

func (t *NopStmt) Pos() token.Pos { return t.Offset }
func (t *NopStmt) End() token.Pos { return token.Pos(int(t.Offset) + len(t.Text)) }
func (*NopStmt) stmtNode()        {}

func (t *IfStmt) Pos() token.Pos { return t.From }
func (t *IfStmt) End() token.Pos { return t.To }
func (*IfStmt) stmtNode()        {}

func (t *UnaryExpr) Pos() token.Pos { return t.Offset }
func (t *UnaryExpr) End() token.Pos { return t.X.End() }
func (*UnaryExpr) exprNode()        {}

func (t *BinaryExpr) Pos() token.Pos { return t.X.Pos() }
func (t *BinaryExpr) End() token.Pos { return t.Y.End() }
func (*BinaryExpr) exprNode()        {}
