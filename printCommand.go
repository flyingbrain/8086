package main

import (
	"fmt"
	"strings"
)

func printCommand(data []decodedCommand) {
	fmt.Print("bits 16\n\n")
	var str strings.Builder

	for _, com := range data {

		fmt.Fprintf(&str, "%s ", com.optcode)
		if com.amb && com.comType != jump {
			size := "byte"
			if com.w {
				size = "word"
			}

			fmt.Fprintf(&str, "%s ", size)
		}

		sep := ", "

		for _, o := range com.value {
			if o != nil {
				fmt.Fprintf(&str, "%s%s", o.printOp(), sep)
			}
			sep = ""
		}

		str.WriteString("\n")
	}

	fmt.Print(str.String())
}

func (o modOperand) printOp() string {
	if o.value != 0 {
		if o.base == "" {
			return fmt.Sprintf("[%d]", o.value)
		}

		if o.value > 0 {
			return fmt.Sprintf("[%s + %d]", o.base, o.value)
		}

		return fmt.Sprintf("[%s - %d]", o.base, o.value*-1)
	}

	return fmt.Sprintf("%s", o.base)
}

func (o directOperand) printOp() string {
	return fmt.Sprintf("%d", o.value)
}

func (o registerOperand) printOp() string {
	c := string(regs[o.reg])
	if o.size == 2 && o.reg < 4 {
	} else if o.size == 1 {
		c = string(regs[o.reg])[0:1]
		if o.h == 1 {
			c = "h"
		} else {
			c = "l"
		}
	}

	return c
}
