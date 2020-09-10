package interpreter

import "dxkite.cn/language/macro/token"

// 运行二元操作
func evalBinary(x, y interface{}, tok token.Token) interface{} {
	ix, ixo := x.(int64)
	fx, _ := x.(float64)
	iy, iyo := y.(int64)
	fy, _ := y.(float64)
	switch tok {
	case token.ADD:
		if ixo {
			if iyo {
				return ix + iy
			} else {
				return float64(ix) + fy
			}
		} else {
			if iyo {
				return fx + float64(iy)
			} else {
				return fx + fy
			}
		}
	case token.MUL:
		if ixo {
			if iyo {
				return ix * iy
			} else {
				return float64(ix) * fy
			}
		} else {
			if iyo {
				return fx * float64(iy)
			} else {
				return fx * fy
			}
		}
	case token.SUB:
		if ixo {
			if iyo {
				return ix - iy
			} else {
				return float64(ix) - fy
			}
		} else {
			if iyo {
				return fx - float64(iy)
			} else {
				return fx - fy
			}
		}
	case token.QUO:
		if ixo {
			if iyo {
				return ix / iy
			} else {
				return float64(ix) / fy
			}
		} else {
			if iyo {
				return fx / float64(iy)
			} else {
				return fx / fy
			}
		}
	case token.LSS:
		if ixo {
			if iyo {
				return ix < iy
			} else {
				return float64(ix) < fy
			}
		} else {
			if iyo {
				return fx < float64(iy)
			} else {
				return fx < fy
			}
		}
	case token.GTR:
		if ixo {
			if iyo {
				return ix > iy
			} else {
				return float64(ix) > fy
			}
		} else {
			if iyo {
				return fx > float64(iy)
			} else {
				return fx > fy
			}
		}
	case token.LEQ:
		if ixo {
			if iyo {
				return ix <= iy
			} else {
				return float64(ix) <= fy
			}
		} else {
			if iyo {
				return fx <= float64(iy)
			} else {
				return fx <= fy
			}
		}
	case token.GEQ:
		if ixo {
			if iyo {
				return ix >= iy
			} else {
				return float64(ix) >= fy
			}
		} else {
			if iyo {
				return fx >= float64(iy)
			} else {
				return fx >= fy
			}
		}
	case token.REM:
		if ixo {
			if iyo {
				return ix % iy
			} else {
				return ix % int64(fy)
			}
		} else {
			if iyo {
				return int64(fx) % iy
			} else {
				return int64(fx) % int64(fy)
			}
		}
	case token.AND:
		if ixo {
			if iyo {
				return ix & iy
			} else {
				return ix & int64(fy)
			}
		} else {
			if iyo {
				return int64(fx) & iy
			} else {
				return int64(fx) & int64(fy)
			}
		}
	case token.OR:
		if ixo {
			if iyo {
				return ix | iy
			} else {
				return ix | int64(fy)
			}
		} else {
			if iyo {
				return int64(fx) | iy
			} else {
				return int64(fx) | int64(fy)
			}
		}
	case token.XOR:
		if ixo {
			if iyo {
				return ix ^ iy
			} else {
				return ix ^ int64(fy)
			}
		} else {
			if iyo {
				return int64(fx) ^ iy
			} else {
				return int64(fx) ^ int64(fy)
			}
		}
	case token.SHL:
		if ixo {
			if iyo {
				return ix << iy
			} else {
				return ix << int64(fy)
			}
		} else {
			if iyo {
				return int64(fx) << iy
			} else {
				return int64(fx) << int64(fy)
			}
		}
	case token.SHR:
		if ixo {
			if iyo {
				return ix >> iy
			} else {
				return ix >> int64(fy)
			}
		} else {
			if iyo {
				return int64(fx) >> iy
			} else {
				return int64(fx) >> int64(fy)
			}
		}
	case token.LAND:
		if ixo {
			if iyo {
				return (ix > 0) && (iy > 0)
			} else {
				return (ix > 0) && (fy > 0)
			}
		} else {
			if iyo {
				return (fx > 0) && (iy > 0)
			} else {
				return (fx > 0) && (fy > 0)
			}
		}
	case token.LOR:
		if ixo {
			if iyo {
				return (ix > 0) || (iy > 0)
			} else {
				return (ix > 0) || (fy > 0)
			}
		} else {
			if iyo {
				return (fx > 0) || (iy > 0)
			} else {
				return (fx > 0) || (fy > 0)
			}
		}
	case token.EQL:
		if ixo {
			if iyo {
				return (ix > 0) == (iy > 0)
			} else {
				return (ix > 0) == (fy > 0)
			}
		} else {
			if iyo {
				return (fx > 0) == (iy > 0)
			} else {
				return (fx > 0) == (fy > 0)
			}
		}
	case token.NEQ:
		if ixo {
			if iyo {
				return (ix > 0) != (iy > 0)
			} else {
				return (ix > 0) != (fy > 0)
			}
		} else {
			if iyo {
				return (fx > 0) != (iy > 0)
			} else {
				return (fx > 0) != (fy > 0)
			}
		}
	}
	return ""
}
