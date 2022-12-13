/*
 * 版本号大小比较
 * 支持三组的版本号，形如 1.2.3，超过三组的超过部分被过滤掉，不参与比较
 */

package version

import (
	"strconv"
	"strings"
)

const (
	ResEQ = "eq" // equal
	ResNE = "ne" // not equal
	ResGT = "gt" // greater than
	ResGE = "ge" // greater than or equal to
	ResLT = "lt" // less than
	ResLE = "le" // less than or equal to
)

type Num struct {
	Major, Minor, Patch int
}

func New(version string) (NewObj Num) {
	verList := strings.Split(version, ".")
	switch len(verList) {
	case 1:
		NewObj.Major, _ = strconv.Atoi(verList[0])
	case 2:
		NewObj.Major, _ = strconv.Atoi(verList[0])
		NewObj.Minor, _ = strconv.Atoi(verList[1])
	case 3, 4, 5, 6, 7, 8, 9:
		NewObj.Major, _ = strconv.Atoi(verList[0])
		NewObj.Minor, _ = strconv.Atoi(verList[1])
		NewObj.Patch, _ = strconv.Atoi(verList[2])
	default:
	}
	return
}

func (vn Num) EQ(obj Num) bool {
	return vn.Major == obj.Major && vn.Minor == obj.Minor && vn.Patch == obj.Patch
}

func (vn Num) NE(obj Num) bool {
	return vn.Major != obj.Major || vn.Minor != obj.Minor || vn.Patch != obj.Patch
}

func (vn Num) GT(obj Num) bool {
	if vn.Major < obj.Major {
		return false
	} else if vn.Major == obj.Major {
		if vn.Minor < obj.Minor {
			return false
		} else if vn.Minor == obj.Minor {
			if vn.Patch <= obj.Patch {
				return false
			}
		}
	}
	return true
}

func (vn Num) GE(obj Num) bool {
	return vn.GT(obj) || vn.EQ(obj)
}

func (vn Num) LT(obj Num) bool {
	if vn.Major > obj.Major {
		return false
	} else if vn.Major == obj.Major {
		if vn.Minor > obj.Minor {
			return false
		} else if vn.Minor == obj.Minor {
			if vn.Patch >= obj.Patch {
				return false
			}
		}
	}
	return true
}

func (vn Num) LE(obj Num) bool {
	return vn.LT(obj) || vn.EQ(obj)
}

func (vn Num) Compare(obj Num) string {
	if vn.EQ(obj) {
		return ResEQ
	}
	if vn.GT(obj) {
		return ResGT
	}
	if vn.LT(obj) {
		return ResLT
	}
	return ""
}
