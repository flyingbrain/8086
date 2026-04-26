package main

import (
	"fmt"
)

type OperationType string

const (
	NoneOp   OperationType = "None"
	RegOp    OperationType = "Reg"
	MemOP    OperationType = "Mem"
	ImmOp    OperationType = "Imm"
	RelImmOp OperationType = "RelImm"
)

type modOperand struct {
	base  effectiveAddressBase
	value uint16
	reg   registerIndex
}

type register struct {
	reg  registerIndex
	h    int8
	size int8
}

type operand interface {
	printO() string
}

type operation struct {
	opType OperationType
	value  operand
}

type decodedCommand struct {
	optcode string
	value   [2]operation
}

func getReg(b byte, w int) operation {
	RegTable := [][2]register{
		{{Register_a, 0, 1}, {Register_a, 0, 2}},
		{{Register_c, 0, 1}, {Register_c, 0, 2}},
		{{Register_d, 0, 1}, {Register_d, 0, 2}},
		{{Register_b, 0, 1}, {Register_b, 0, 2}},
		{{Register_a, 1, 1}, {Register_sp, 0, 2}},
		{{Register_c, 1, 1}, {Register_bp, 0, 2}},
		{{Register_d, 1, 1}, {Register_si, 0, 2}},
		{{Register_b, 1, 1}, {Register_di, 0, 2}},
	}

	rez := operation{}
	rez.opType = RegOp
	rez.value = RegTable[int(b)][w]

	return rez
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

func parceDispl(data []byte, pos *int, needed bool, w bool) uint16 {
	if !needed {
		return 0
	}

	p := *pos / 8
	rez := uint16(data[p])
	*pos += 8

	if w {
		rez = uint16(rez)<<8 | uint16(data[p+1])
		*pos += 8
	}

	return rez
}

func decode(buf []byte) ([]decodedCommand, error) {
	cmds := []decodedCommand{}
	for {
		pos := 0
		if len(buf) == 0 {
			break
		}

		for _, command := range Commands {
			readCommand(buf, command, &cmds, &pos)
			buf = buf[pos/8:]
		}

	}
	return cmds, nil
}

// TODO add data if nesessary
func readCommand(buf []byte, com Command, cmds *[]decodedCommand, pos *int) {
	command := decodedCommand{
		optcode: com.Name,
		value:   [2]operation{},
	}
	f := map[DataType]byte{}

	for _, c := range com.value {

		n := 0
		p := 0
		if *pos != 0 {
			n = *pos / 8
			p = *pos % 8
		}

		b := buf[n]
		if *pos == 0 && b>>(8-c.Width) != c.Value {
			break
		}

		f[c.Name] = (b << p) >> (8 - int(c.Width))

		*pos += int(c.Width)
	}

	if *pos%8 != 0 {
		panic(fmt.Sprintf("Commad decode wrong position:%d", *pos))
	}

	D := f[dType]
	W := f[wType]
	MOD := f[modType]
	rm := f[rmType]

	directAddr := MOD == 0b00 && rm == 0b110
	hasDispl := MOD == 0b10 || MOD == 0b01 || directAddr
	DisplW := directAddr || MOD == 0b10

	dispData := parceDispl(buf, pos, hasDispl, DisplW)

	if _, ok := f[regType]; ok {
		command.value[boolToInt(D == 0)] = getReg(f[regType], int(W))
	}

	if _, ok := f[rmType]; ok {
		if MOD == 0b11 {
			command.value[D] = getReg(f[rmType], int(W))
		} else {
			m := modOperand{
				base:  getModName(f[rmType]),
				value: dispData,
				reg:   Register_none,
			}

			op := operation{
				opType: MemOP,
				value:  m,
			}
			command.value[D] = op
		}
	}

	*cmds = append(*cmds, command)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
