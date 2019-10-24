package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	pak "github.com/septianw/jas/common"
)

func main() {

	lib := pak.LoadSo("/home/asep/gocode/src/github.com/septianw/jas/bungkus/modules/core/user/user.so")
	bootsym, err := lib.Lookup("Bootstrap")
	pak.ErrHandler(err)

	routersym, err := lib.Lookup("Router")
	pak.ErrHandler(err)

	bootstrap := bootsym.(func())
	router := routersym.(func(*gin.Engine))

	bootstrap()

	e := gin.Default()

	router(e)

	fmt.Println("vim-go")
}
