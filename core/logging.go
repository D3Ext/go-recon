package core

import (
	"fmt"
	"github.com/fatih/color"
)

var red func(a ...interface{}) string = color.New(color.FgRed).SprintFunc()
var green func(a ...interface{}) string = color.New(color.FgGreen).SprintFunc()
var magenta func(a ...interface{}) string = color.New(color.FgMagenta).SprintFunc()

func Red(str string, c bool) {
	if c {
		fmt.Println("["+red("-")+"]", str)
	} else {
		fmt.Println("[-]", str)
	}
}

func Green(str string, c bool) {
	if c {
		fmt.Println("["+green("+")+"]", str)
	} else {
		fmt.Println("[+]", str)
	}
}

func Magenta(str string, c bool) {
	if c {
		fmt.Println("["+magenta("*")+"]", str)
	} else {
		fmt.Println("[*]", str)
	}
}
