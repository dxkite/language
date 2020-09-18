package ast

import (
	"dxkite.cn/language/macro/token"
)

// 节点
type Node interface {
	Pos() token.Pos
	End() token.Pos
}

// 宏字面量
type MacroLiter interface {
	Node
	litNode()
}

// 语句
type Stmt interface {
	Node
	stmtNode()
}

// 定义语句
type DefineStmt interface {
	Stmt
	defineNode()
}

// 条件语句
type CondStmt interface {
	Stmt
	SetFromTO(from, to token.Pos)
	SetTrueStmt(t Stmt)
	SetFalseStmt(f Stmt)
}

// 文件包含类型
type IncludeType int

const (
	IncludeInner IncludeType = iota // 系统内
	IncludeOuter                    // 相对
)

type (
	// 错误表达式
	BadExpr struct {
		Offset token.Pos // 标识符位置
		Token  token.Token
		Lit    string
	}

	// 标识符
	Ident struct {
		Offset token.Pos // 标识符位置
		Name   string    // 名称
	}
	// 普通文本（宏中的代码块）
	Text struct {
		Offset token.Pos   // 标识符位置
		Kind   token.Token // 类型
		Text   string      // 文本内容
	}
	// 注释
	Comment struct {
		Offset token.Pos   // 标识符位置
		Kind   token.Token // 注释类型
		Text   string      // 文本内容
	}

	// 语句块
	BlockStmt []Stmt

	// 值定义
	ValDefineStmt struct {
		From, To token.Pos      // 标识符位置
		Name     *Ident         // 定义的标识符
		Body     *MacroLitArray // 定义的语句
	}

	// 取消定义指令
	UnDefineStmt struct {
		From, To token.Pos // 标识符位置
		Name     *Ident    // 定义的标识符
	}

	// 函数定义
	FuncDefineStmt struct {
		From, To  token.Pos      // 标识符位置
		Name      *Ident         // 定义的标识符
		Lparen    token.Pos      // (
		IdentList []*Ident       // 定义的参数
		Rparen    token.Pos      // )
		Body      *MacroLitArray // 定义的语句
	}

	// 文件包含语句
	IncludeStmt struct {
		From, To token.Pos   // 标识符位置
		Path     string      // 文件路径
		Type     IncludeType // 文件包含类型
	}

	// 宏调用
	MacroCallExpr struct {
		From, To  token.Pos      // 标识符位置
		Name      *Ident         // 定义的标识符
		Lparen    token.Pos      // (
		ParamList *MacroLitArray // 调用的参数列表
		Rparen    token.Pos      // )
	}

	// 括号表达式
	ParenExpr struct {
		Lparen token.Pos  // "("
		X      MacroLiter // 表达式值
		Rparen token.Pos  // ")"
	}

	// 字面量数组
	// 参数特用
	MacroLitArray []MacroLiter

	// 数值/字符串字面量
	// token.INT token.FLOAT token.STRING token.CHAR
	LitExpr struct {
		Offset token.Pos // 标识符位置
		Kind   token.Token
		Value  string // 值
	}

	// 报错表达式
	MacroCmdStmt struct {
		Offset token.Pos   // 标识符位置
		Kind   token.Token // 类型
		Cmd    string      // 报错文本内容
	}

	// 行语句
	LineStmt struct {
		From, To token.Pos // 标识符位置
		Line     string    // 文件行
		Path     string    // 文件名
	}

	// 非法宏
	InvalidStmt struct {
		Offset token.Pos // 标识符位置
		Text   string    // 报错文本内容
	}

	// if语句
	IfStmt struct {
		From, To token.Pos  // 标识符位置
		X        MacroLiter // 条件
		Then     Stmt       // 正确
		Else     Stmt       // 错误
	}

	ElseIfStmt struct {
		From, To token.Pos  // 标识符位置
		X        MacroLiter // 条件
		Then     Stmt       // 正确
		Else     Stmt       // 错误
	}

	// ifdef语句
	IfDefStmt struct {
		From, To token.Pos // 标识符位置
		Name     *Ident    // 定义的标识符
		Then     Stmt      // 正确
		Else     Stmt      // 错误
	}

	// ifndef语句
	IfNoDefStmt struct {
		From, To token.Pos // 标识符位置
		Name     *Ident    // 定义的标识符
		Then     Stmt      // 正确
		Else     Stmt      // 错误
	}

	// 一元运算
	UnaryExpr struct {
		Offset token.Pos   // 标识符位置
		Op     token.Token // 操作类型
		X      MacroLiter  // 操作的表达式
	}

	// 二元运算
	BinaryExpr struct {
		X      MacroLiter  // 左值
		Offset token.Pos   // 操作符位置
		Op     token.Token // 操作类型
		Y      MacroLiter  // 右值
	}
)

//------ Node
func (t *BadExpr) Pos() token.Pos { return t.Offset }
func (t *BadExpr) End() token.Pos { return token.Pos(int(t.Offset) + len(t.Lit)) }
func (*BadExpr) stmtNode()        {}
func (*BadExpr) litNode()         {}

func (t *Ident) Pos() token.Pos { return t.Offset }
func (t *Ident) End() token.Pos { return token.Pos(int(t.Offset) + len(t.Name)) }
func (*Ident) litNode()         {}

func (t *Text) Pos() token.Pos { return t.Offset }
func (t *Text) End() token.Pos { return token.Pos(int(t.Offset) + len(t.Text)) }
func (*Text) stmtNode()        {}
func (*Text) litNode()         {}

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

// 尾部添加
func (t *Text) Append(m *Text) {
	t.Text += m.Text
}

func (t *Comment) Pos() token.Pos { return t.Offset }
func (t *Comment) End() token.Pos { return token.Pos(int(t.Offset) + len(t.Text)) }
func (*Comment) stmtNode()        {}

func (t *BlockStmt) Pos() token.Pos {
	if len(*t) > 0 {
		return (*t)[0].Pos()
	}
	return 0
}
func (t *BlockStmt) End() token.Pos {
	l := len(*t)
	if l > 0 {
		return (*t)[l-1].End()
	}
	return 0
}

func (t *BlockStmt) Add(stmt Stmt) {
	*t = append(*t, stmt)
}

func (*BlockStmt) stmtNode() {}

func (t *ValDefineStmt) Pos() token.Pos { return t.From }
func (t *ValDefineStmt) End() token.Pos { return t.To }
func (*ValDefineStmt) stmtNode()        {}
func (*ValDefineStmt) defineNode()      {}

func (t *UnDefineStmt) Pos() token.Pos { return t.From }
func (t *UnDefineStmt) End() token.Pos { return t.To }
func (*UnDefineStmt) stmtNode()        {}

func (t *FuncDefineStmt) Pos() token.Pos { return t.From }
func (t *FuncDefineStmt) End() token.Pos { return t.To }
func (*FuncDefineStmt) stmtNode()        {}
func (*FuncDefineStmt) defineNode()      {}

// 是否有参数
func (t *FuncDefineStmt) hasParam() bool { return len(t.IdentList) > 0 }

func (t *IncludeStmt) Pos() token.Pos { return t.From }
func (t *IncludeStmt) End() token.Pos { return t.To }
func (*IncludeStmt) stmtNode()        {}

func (t *MacroCallExpr) Pos() token.Pos { return t.From }
func (t *MacroCallExpr) End() token.Pos { return t.To }
func (*MacroCallExpr) litNode()         {}

func (t *ParenExpr) Pos() token.Pos { return t.Lparen }
func (t *ParenExpr) End() token.Pos { return t.Rparen }
func (*ParenExpr) litNode()         {}

func (t *LitExpr) Pos() token.Pos { return t.Offset }
func (t *LitExpr) End() token.Pos { return token.Pos(int(t.Offset) + len(t.Value)) }
func (*LitExpr) litNode()         {}

func (t *MacroCmdStmt) Pos() token.Pos { return t.Offset }
func (t *MacroCmdStmt) End() token.Pos { return token.Pos(int(t.Offset) + len(t.Cmd)) }
func (*MacroCmdStmt) stmtNode()        {}

func (t *LineStmt) Pos() token.Pos { return t.From }
func (t *LineStmt) End() token.Pos { return t.To }
func (*LineStmt) stmtNode()        {}

func (t *InvalidStmt) Pos() token.Pos { return t.Offset }
func (t *InvalidStmt) End() token.Pos { return token.Pos(int(t.Offset) + len(t.Text)) }
func (*InvalidStmt) stmtNode()        {}

func (stmt *IfStmt) Pos() token.Pos { return stmt.From }
func (stmt *IfStmt) End() token.Pos { return stmt.To }
func (*IfStmt) stmtNode()           {}
func (stmt *IfStmt) SetFromTO(from, to token.Pos) {
	stmt.From = from
	stmt.To = to
}
func (stmt *IfStmt) SetTrueStmt(t Stmt) {
	stmt.Then = t
}
func (stmt *IfStmt) SetFalseStmt(f Stmt) {
	stmt.Else = f
}
func (stmt *IfDefStmt) Pos() token.Pos { return stmt.From }
func (stmt *IfDefStmt) End() token.Pos { return stmt.To }
func (*IfDefStmt) stmtNode()           {}
func (stmt *IfDefStmt) SetFromTO(from, to token.Pos) {
	stmt.From = from
	stmt.To = to
}
func (stmt *IfDefStmt) SetTrueStmt(t Stmt) {
	stmt.Then = t
}
func (stmt *IfDefStmt) SetFalseStmt(f Stmt) {
	stmt.Else = f
}
func (stmt *IfNoDefStmt) Pos() token.Pos { return stmt.From }
func (stmt *IfNoDefStmt) End() token.Pos { return stmt.To }
func (*IfNoDefStmt) stmtNode()           {}
func (stmt *IfNoDefStmt) SetFromTO(from, to token.Pos) {
	stmt.From = from
	stmt.To = to
}
func (stmt *IfNoDefStmt) SetTrueStmt(t Stmt) {
	stmt.Then = t
}
func (stmt *IfNoDefStmt) SetFalseStmt(f Stmt) {
	stmt.Else = f
}
func (stmt *ElseIfStmt) Pos() token.Pos { return stmt.From }
func (stmt *ElseIfStmt) End() token.Pos { return stmt.To }
func (*ElseIfStmt) stmtNode()           {}
func (stmt *ElseIfStmt) SetFromTO(from, to token.Pos) {
	stmt.From = from
	stmt.To = to
}
func (stmt *ElseIfStmt) SetTrueStmt(t Stmt) {
	stmt.Then = t
}
func (stmt *ElseIfStmt) SetFalseStmt(f Stmt) {
	stmt.Else = f
}
func (t *UnaryExpr) Pos() token.Pos { return t.Offset }
func (t *UnaryExpr) End() token.Pos { return t.X.End() }
func (*UnaryExpr) litNode()         {}

func (t *BinaryExpr) Pos() token.Pos { return t.X.Pos() }
func (t *BinaryExpr) End() token.Pos { return t.Y.End() }
func (*BinaryExpr) litNode()         {}

func (t MacroLitArray) Pos() token.Pos {
	if len(t) > 0 {
		return t[0].Pos()
	}
	return 0
}
func (t MacroLitArray) End() token.Pos {
	l := len(t)
	if l > 0 {
		return t[l-1].End()
	}
	return 0
}

func (t *MacroLitArray) Append(liter MacroLiter) {
	*t = append(*t, liter)
}

func (*MacroLitArray) litNode()  {}
func (*MacroLitArray) stmtNode() {}
