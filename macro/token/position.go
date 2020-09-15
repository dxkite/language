package token

import (
	"fmt"
)

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
	return fmt.Sprintf("%d:%d", pos.Line, pos.Column)
}

// Pos转换成Position
type FilePos struct {
	size  int
	lines []int
}

func (f *FilePos) Init(src []byte) {
	var lines []int
	line := 0
	for offset, b := range src {
		if line >= 0 {
			lines = append(lines, line)
		}
		line = -1
		if b == '\n' {
			line = offset + 1
		}
	}
	lines = append(lines, len(src))
	f.lines = lines
	f.size = len(src)
}

func (f FilePos) CreatePosition(p Pos) (pos Position) {
	if int(p) >= 0 && int(p) < f.size {
		line := 0
		column := 0
		for i, l := range f.lines {
			if l > int(p) {
				line = i
				column = int(p) - f.lines[i-1]
				break
			}
		}
		return Position{
			Offset: int(p),
			Line:   line,
			Column: column,
		}
	}
	return
}
