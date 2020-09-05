package scanner

import "fmt"

// 位置
type Position struct {
	Offset int // 整体偏移 0
	Line   int // 行
	Column int // 列
}

// 是否可用
func (pos *Position) IsValid() bool { return pos.Offset > 0 }

// 位置打印字符串
func (pos Position) String() string {
	return fmt.Sprintf("%d:%d:%d", pos.Offset, pos.Line, pos.Column)
}
