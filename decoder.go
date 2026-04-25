package main

import (
	"fmt"
	"io"
)

type OperationType string

const (
	NoneOp   OperationType = "None"
	RegOp    OperationType = "Reg"
	MemOP    OperationType = "Mem"
	ImmOp    OperationType = "Imm"
	RelImmOp OperationType = "RelImm"
)

type operation struct {
	opType OperationType
	value  register
}

type decodedCommand struct {
	optcode string
	value   [2]operation
}

type register struct {
	reg  registerIndex
	h    int8
	size int8
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

	D, isD := f[dType]
	W, isW := f[wType]
	//MOD, isMOD := f[modType]

	if _, ok := f[regType]; ok {
		if isD && isW {
			command.value[boolToInt(D == 0)] = getReg(f[regType], int(W))
		}
	}

	if _, ok := f[rmType]; ok {
		if isD && isW {
			command.value[D] = getReg(f[rmType], int(W))
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

func regMemDecode(b byte, file io.ReadCloser) (string, error) {
	d := (b >> 1) & 0b1
	w := b & 0b1

	buf := make([]byte, 1)

	_, err := file.Read(buf)
	if err != nil {
		return "", err
	}

	b2 := buf[0]

	mod := (b2 >> 6) & 0b11
	reg := (b2 >> 3) & 0b111
	rm := b2 & 0b111

	rmText := getRM(rm, w, mod, file)
	regText := getRegister(reg, w)
	command := ""

	if d == 0b0 {
		command = fmt.Sprintf("%s, %s\n", rmText, regText)
	} else {
		command = fmt.Sprintf("%s, %s\n", regText, rmText)
	}

	return command, nil
}

func immMovRegMemDecode(b byte, file io.ReadCloser) (string, error) {
	w := b & 0b1

	buf := make([]byte, 1)
	_, err := file.Read(buf)

	if err != nil {
		return "", err
	}

	b2 := buf[0]

	mod := (b2 >> 6) & 0b11
	rm := b2 & 0b111

	rmText := getRM(rm, w, mod, file)

	if w == 0b1 {
		buf = make([]byte, 2)
	}

	_, err = file.Read(buf)
	if err != nil {
		return "", err
	}

	value := getNumber(buf, true)
	command := ""

	if w == 0b1 {
		command = fmt.Sprintf("%s, word %d\n", rmText, value)
	} else {
		command = fmt.Sprintf("%s, byte %d\n", rmText, value)
	}

	return command, nil
}

func immRegMemDecode(b byte, file io.ReadCloser) (string, error) {
	commands := make([]string, 8)
	commands[0] = "add "
	commands[5] = "sub "
	commands[7] = "cmp "

	w := b & 0b1
	s := b >> 1 & 0b1

	buf := make([]byte, 1)
	_, err := file.Read(buf)
	if err != nil {
		return "", err
	}

	b2 := buf[0]

	mod := (b2 >> 6) & 0b11
	opt := b2 >> 3 & 0b111
	rm := b2 & 0b111

	command := commands[uint8(opt)]

	rmText := getRM(rm, w, mod, file)

	if s|w == 0b11 {
		buf = make([]byte, 2)
	}

	_, err = file.Read(buf)
	if err != nil {
		return "", err
	}

	value := getNumber(buf, s == 0b1)

	if w == 0b1 {
		command += fmt.Sprintf("%s, word %d\n", rmText, value)
	} else {
		command += fmt.Sprintf("%s, byte %d\n", rmText, value)
	}

	return command, nil
}

func immRegDecode(b byte, file io.ReadCloser) (string, error) {
	w := b >> 3 & 0b1
	reg := b & 0b111

	regText := getRegister(reg, w)
	buf := make([]byte, 1)

	if w == 0b1 {
		buf = make([]byte, 2)
	}

	_, err := file.Read(buf)
	if err != nil {
		return "", err
	}

	value := getNumber(buf, false)

	return fmt.Sprintf("%s, %d\n", regText, value), nil
}

func accumDecode(b byte, file io.ReadCloser) (string, error) {
	w := b & 0b1
	p := b >> 1 & 0b1
	size := 1
	if w == 0b1 {
		size = 2
	}

	buf := make([]byte, size)
	_, err := file.Read(buf)
	if err != nil {
		return "", err
	}

	reg := "al"
	value := getNumber(buf[:1], false)
	if w == 0b1 {
		value = getNumber(buf, false)
		reg = "ax"
	}

	command := ""

	if p == 0b1 {
		command = fmt.Sprintf("[%d], %s\n", value, reg)
	} else {
		command = fmt.Sprintf("%s,[%d]\n", reg, value)
	}

	return command, err
}

func getNumber(data []byte, signed bool) int {
	var u uint16

	if len(data) == 1 {
		u = uint16(data[0])
		if signed {
			return int(int8(u)) // correct sign for 1 byte
		}
		return int(u)
	}

	if len(data) == 2 {
		u = uint16(data[1])<<8 | uint16(data[0]) // little endian

		if signed {
			return int(int16(u)) // correct sign for 2 bytes
		}
		return int(u)
	}

	panic("unsupported size")
}
