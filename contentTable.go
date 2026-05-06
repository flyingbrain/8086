package main

type DataType string

const (
	Literal  DataType = "Literal"
	wType    DataType = "W"
	sType    DataType = "S"
	dType    DataType = "D"
	srType   DataType = "SR"
	regType  DataType = "REG"
	modType  DataType = "MOD"
	rmType   DataType = "RM"
	dispType DataType = "Displ"
	dataType DataType = "Data"
	addrType DataType = "Addr"
)

type Field struct {
	Name     DataType
	Width    uint8
	Shift    uint8
	Value    byte
	HasValue bool
}

type Command struct {
	Name  string
	value []Field
}

func parseBinary(s string) byte {
	var v byte
	for _, c := range s {
		v <<= 1
		if c == '1' {
			v |= 1
		}
	}
	return v
}

func B(bits string) Field {
	return Field{Literal, uint8(len(bits)), 0, parseBinary(bits), true}
}

func D() Field {
	return Field{dType, 1, 0, 0, false}
}

func W() Field {
	return Field{wType, 1, 0, 0, false}
}

func S() Field {
	return Field{sType, 1, 0, 0, false}
}

func SR() Field {
	return Field{srType, 2, 0, 0, false}
}

func REG() Field {
	return Field{regType, 3, 0, 0, false}
}

func RM() Field {
	return Field{rmType, 3, 0, 0, false}
}

func MOD() Field {
	return Field{modType, 2, 0, 0, false}
}

func DATA() Field {
	return Field{dataType, 0, 0, 0, false}
}

func ImpD(val byte) Field {
	return Field{dType, 0, 0, val, true}
}

func ImpREG(val byte) Field {
	return Field{regType, 0, 0, val, true}
}

func ImpRM(val byte) Field {
	return Field{rmType, 0, 0, val, true}
}

func ImpMOD(val byte) Field {
	return Field{modType, 0, 0, val, true}
}

func ADDR() Field {
	return Field{addrType, 0, 0, 0, false}
}

func DISP() Field {
	return Field{dispType, 0, 0, 0, false}
}

var Commands = []Command{
	//mov
	{"mov", []Field{B("100010"), D(), W(), MOD(), REG(), RM()}},
	{"mov", []Field{B("1100011"), W(), MOD(), B("000"), RM(), DATA(), ImpD(0)}},
	{"mov", []Field{B("1011"), W(), REG(), DATA(), ImpD(1)}},
	{"mov", []Field{B("1010000"), W(), ADDR(), ImpREG(0), ImpMOD(0), ImpRM(0b110), ImpD(1)}},
	{"mov", []Field{B("1010001"), W(), ADDR(), ImpREG(0), ImpMOD(0), ImpRM(0b110), ImpD(0)}},
	{"mov", []Field{B("100011"), D(), B("0"), MOD(), B("0"), SR(), RM()}},

	// add
	{"add", []Field{B("000000"), D(), W(), MOD(), REG(), RM()}},
	{"add", []Field{B("100000"), S(), W(), MOD(), B("000"), RM(), DATA()}},
	{"add", []Field{B("0000010"), W(), DATA(), ImpREG(0), ImpD(1)}},

	// sub
	{"sub", []Field{B("001010"), D(), W(), MOD(), REG(), RM()}},
	{"sub", []Field{B("100000"), S(), W(), MOD(), B("101"), RM(), DATA()}},
	{"sub", []Field{B("0010110"), W(), DATA(), ImpREG(0), ImpD(1)}},

	// cmp
	{"cmp", []Field{B("001110"), D(), W(), MOD(), REG(), RM()}},
	{"cmp", []Field{B("100000"), S(), W(), MOD(), B("111"), RM(), DATA()}},
	{"cmp", []Field{B("0011110"), W(), DATA(), ImpREG(0), ImpD(1)}},

	// jumps
	{"je", []Field{B("01110100"), DISP()}},
	{"jl", []Field{B("01111100"), DISP()}},
	{"jle", []Field{B("01111110"), DISP()}},
	{"jb", []Field{B("01110010"), DISP()}},
	{"jbe", []Field{B("01110110"), DISP()}},
	{"jp", []Field{B("01111010"), DISP()}},
	{"jo", []Field{B("01110000"), DISP()}},
	{"js", []Field{B("01111000"), DISP()}},
	{"jne", []Field{B("01110101"), DISP()}},
	{"jnl", []Field{B("01111101"), DISP()}},
	{"jg", []Field{B("01111111"), DISP()}},
	{"jnb", []Field{B("01110011"), DISP()}},
	{"ja", []Field{B("01110111"), DISP()}},
	{"jnp", []Field{B("01111011"), DISP()}},
	{"jno", []Field{B("01110001"), DISP()}},
	{"jns", []Field{B("01111001"), DISP()}},

	// loops
	{"loop", []Field{B("11100010"), DISP()}},
	{"loopz", []Field{B("11100001"), DISP()}},
	{"loopnz", []Field{B("11100000"), DISP()}},
	{"jcxz", []Field{B("11100011"), DISP()}},
}
