// Package api
// @description
// @author      梁志豪
// @datetime    2025/12/30 20:04
package api

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"testing"
)

type LoginForm struct {
	UserName string `form:"username"`
	Password string `form:"password"`
}

func TestLogin(t *testing.T) {
	form := LoginForm{
		UserName: "梁志豪",
		Password: "lzh5201314.",
	}
	name, err := sanitizeUserName(form.UserName)
	if err != nil {
		fmt.Println(err)
		return
	}
	form.UserName = name
	data, err := Login(form.UserName, form.Password)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%+v\n", data)
}

func sanitizeUserName(name string) (string, error) {
	if len(name) > 30 {
		return "", errors.New("username too long")
	}

	illegal := []string{`"`, `'`, "!", "@", "+", "="}
	for _, c := range illegal {
		if strings.Contains(name, c) {
			return "", errors.New("illegal character")
		}
	}

	name = strings.TrimSpace(name)
	name = strings.NewReplacer(`"`, "", "!", "", "0x", "").Replace(name)

	if len(name) > 100 {
		name = name[:100]
	}

	return name, nil
}

func TestMd5(t *testing.T) {
	text := "lzh5201314."
	hash := md5.Sum([]byte(text))
	content := hex.EncodeToString(hash[:])
	fmt.Printf("%s\n", content)
}
