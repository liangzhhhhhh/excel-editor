// Package api
// @description
// @author      梁志豪
// @datetime    2025/12/30 19:51
package api

import (
	"crypto/md5"
	"encoding/hex"
)

func md5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
