package main

import (
	"fmt"
	"strings"
)

type modOperand struct {
	base  effectiveAddressBase
	value int16
}

type directOperand struct {
	value int
}

type registerOperand struct {
	reg  uint8
	h    int8
	size int8
}

type operand interface {
	printOp() string
	getValue() uint16
	exec(opcode string, s operand, str *strings.Builder)
}

type commandType string

const (
	jump  commandType = "jump"
	regul commandType = "regul"
)

type decodedCommand struct {
	value   [2]operand
	comType commandType
	optcode string
	amb     bool
	w       bool
}

func getReg(b byte, w int) operand {
	RegTable := [][2]registerOperand{
		{{0, 0, 1}, {0, 0, 2}},
		{{1, 0, 1}, {1, 0, 2}},
		{{2, 0, 1}, {2, 0, 2}},
		{{3, 0, 1}, {3, 0, 2}},
		{{0, 1, 1}, {4, 0, 2}},
		{{1, 1, 1}, {5, 0, 2}},
		{{2, 1, 1}, {6, 0, 2}},
		{{3, 1, 1}, {7, 0, 2}},
	}

	return RegTable[int(b)][w]
}

func getModName(b byte) effectiveAddressBase {

	tab := []effectiveAddressBase{
		EffectiveAddress_bx_si,
		EffectiveAddress_bx_di,
		EffectiveAddress_bp_si,
		EffectiveAddress_bp_di,
		EffectiveAddress_si,
		EffectiveAddress_di,
		EffectiveAddress_bp,
		EffectiveAddress_bx,
	}

	return tab[b]
}

func parceDispl(data []byte, pos *int, needed bool, w bool) int16 {
	if !needed {
		return 0
	}

	p := *pos / 8
	rez := int16(data[p])
	*pos += 8

	if w {
		if data[p+1] != 0 {
			rez = int16(data[p+1])<<8 | int16(rez)
		}
		*pos += 8
	}

	return int16(rez)
}

func decode(buf []byte) ([]decodedCommand, error) {
	cmds := []decodedCommand{}

	for len(buf) > 0 {
		matched := false
		pos := 0

		for _, command := range Commands {
			if readCommand(buf, command, &cmds, &pos) {
				matched = true
				buf = buf[pos/8:]
				break
			}
		}

		if !matched {
			fmt.Printf("can not transalate instruction %08b, trying next byte\n", buf[0])
			buf = buf[1:]
		}

	}
	return cmds, nil
}

func readCommand(buf []byte, com Command, cmds *[]decodedCommand, pos *int) bool {
	f := map[DataType]byte{}

	for _, c := range com.value {

		n := 0
		p := 0
		if *pos != 0 {
			n = *pos / 8
			p = *pos % 8
		}

		b := buf[n]
		if c.Name == Literal && (b<<p)>>(8-int(c.Width)) != c.Value {
			*pos = 0
			return false
		}

		if c.HasValue {
			f[c.Name] = c.Value
		} else {
			f[c.Name] = (b << p) >> (8 - int(c.Width))
		}

		*pos += int(c.Width)
	}

	if *pos%8 != 0 {
		panic(fmt.Sprintf("Commad decode wrong position:%d", *pos))
	}

	command := decodedCommand{
		value:   [2]operand{},
		comType: regul,
		optcode: com.Name,
		amb:     true,
		w:       false,
	}

	D := f[dType]
	W := f[wType]
	S, isS := f[sType]
	MOD := f[modType]
	rm := f[rmType]
	_, isDisp := f[dispType]

	directAddr := MOD == 0b00 && rm == 0b110
	hasDispl := MOD == 0b10 || MOD == 0b01 || directAddr || isDisp
	DisplW := directAddr || MOD == 0b10 || W != 0 && isDisp

	dispData := parceDispl(buf, pos, hasDispl, DisplW)

	_, isData := f[dataType]
	hasData := isData || isS
	dataW := isS && S == 1 || W == 0

	command.w = DisplW || !dataW

	// REG
	if _, ok := f[regType]; ok {
		command.value[boolToInt(D == 0)] = getReg(f[regType], int(W))
		command.amb = false
	}

	// RM
	if _, ok := f[rmType]; ok {
		if MOD == 0b11 {
			command.value[D] = getReg(f[rmType], int(W))
			command.amb = false
		} else {

			modName := getModName(f[rmType])
			if directAddr {
				modName = EffectiveAddress_direct
			}

			operand := modOperand{
				base:  modName,
				value: dispData,
			}

			command.value[D] = operand
		}
	} else if dispData != 0 {
		operand := modOperand{
			base:  EffectiveAddress_direct,
			value: dispData,
		}

		command.comType = jump

		command.value[1] = operand
	}

	// direct data
	if hasData {
		data := parceDispl(buf, pos, hasData, !dataW)
		dataOper := directOperand{
			value: int(data),
		}

		idx := 0
		if command.value[0] != nil {
			idx = 1
		}

		command.value[idx] = dataOper
	}

	*cmds = append(*cmds, command)
	return true
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
