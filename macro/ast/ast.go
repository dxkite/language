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

// 宏字面量
type MacroLiter interface {
	Expr
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
		From, To token.Pos     // 标识符位置
		Name     *Ident        // 定义的标识符
		Body     MacroLitArray // 定义的语句
	}

	// 取消定义指令
	ValUnDefineStmt struct {
		From, To token.Pos // 标识符位置
		Name     *Ident    // 定义的标识符
	}

	// 函数定义
	FuncDefineStmt struct {
		From, To  token.Pos     // 标识符位置
		Name      *Ident        // 定义的标识符
		LParam    token.Pos     // (
		IdentList []*Ident      // 定义的参数
		RParam    token.Pos     // )
		Body      MacroLitArray // 定义的语句
	}

	// 文件包含语句
	IncludeStmt struct {
		From, To token.Pos   // 标识符位置
		Path     string      // 文件路径
		Type     IncludeType // 文件包含类型
	}

	// 宏调用
	MacroCallExpr struct {
		From, To  token.Pos     // 标识符位置
		Name      *Ident        // 定义的标识符
		LParam    token.Pos     // (
		ParamList MacroLitArray // 调用的参数列表
		RParam    token.Pos     // )
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
	ErrorStmt struct {
		Offset token.Pos // 标识符位置
		Msg    string    // 报错文本内容
	}

	// 行语句
	LineStmt struct {
		From, To token.Pos // 标识符位置
		Line     string    // 文件行
		Path     string    // 文件名
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
		Offset token.Pos   // 操作符位置
		Op     token.Token // 操作类型
		Y      Expr        // 右值
	}
)

//------ Node
func (t *BadExpr) Pos() token.Pos { return t.Offset }
func (t *BadExpr) End() token.Pos { return token.Pos(int(t.Offset) + len(t.Lit)) }
func (*BadExpr) exprNode()        {}
func (*BadExpr) stmtNode()        {}
func (*BadExpr) litNode()         {}

func (t *Ident) Pos() token.Pos { return t.Offset }
func (t *Ident) End() token.Pos { return token.Pos(int(t.Offset) + len(t.Name)) }
func (*Ident) exprNode()        {}
func (*Ident) litNode()         {}

func (t *Text) Pos() token.Pos { return t.Offset }
func (t *Text) End() token.Pos { return token.Pos(int(t.Offset) + len(t.Text)) }
func (*Text) stmtNode()        {}
func (*Text) exprNode()        {}
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

func (t *ValUnDefineStmt) Pos() token.Pos { return t.From }
func (t *ValUnDefineStmt) End() token.Pos { return t.To }
func (*ValUnDefineStmt) stmtNode()        {}

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
func (*MacroCallExpr) exprNode()        {}
func (*MacroCallExpr) litNode()         {}

func (t *LitExpr) Pos() token.Pos { return t.Offset }
func (t *LitExpr) End() token.Pos { return token.Pos(int(t.Offset) + len(t.Value)) }
func (*LitExpr) exprNode()        {}
func (*LitExpr) litNode()         {}

func (t *ErrorStmt) Pos() token.Pos { return t.Offset }
func (t *ErrorStmt) End() token.Pos { return token.Pos(int(t.Offset) + len(t.Msg)) }
func (*ErrorStmt) stmtNode()        {}

func (t *LineStmt) Pos() token.Pos { return t.From }
func (t *LineStmt) End() token.Pos { return t.To }
func (*LineStmt) stmtNode()        {}

func (t *NopStmt) Pos() token.Pos { return t.Offset }
func (t *NopStmt) End() token.Pos { return token.Pos(int(t.Offset) + len(t.Text)) }
func (*NopStmt) stmtNode()        {}

func (t *IfStmt) Pos() token.Pos { return t.From }
func (t *IfStmt) End() token.Pos { return t.To }
func (*IfStmt) stmtNode()        {}

func (t *UnaryExpr) Pos() token.Pos { return t.Offset }
func (t *UnaryExpr) End() token.Pos { return t.X.End() }
func (*UnaryExpr) exprNode()        {}
func (*UnaryExpr) litNode()         {}

func (t *BinaryExpr) Pos() token.Pos { return t.X.Pos() }
func (t *BinaryExpr) End() token.Pos { return t.Y.End() }
func (*BinaryExpr) exprNode()        {}
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

func (t *MacroLitArray) AddLit(liter MacroLiter) {
	l := len(*t)
	if l > 0 {
		ta, okl := liter.(*Text)
		ea, oke := (*t)[l-1].(*Text)
		if okl && oke {
			ea.Append(ta)
			(*t)[l-1] = ea
			return
		}
	}
	*t = append(*t, liter)
}

func (MacroLitArray) exprNode() {}
func (MacroLitArray) litNode()  {}
func (MacroLitArray) stmtNode() {}
