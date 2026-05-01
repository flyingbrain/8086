package main

type DataType string

const (
	Literal  DataType = "Literal"
	wType    DataType = "W"
	dType    DataType = "D"
	srType   DataType = "SR"
	regType  DataType = "REG"
	modType  DataType = "MOD"
	rmType   DataType = "RM"
	dispType DataType = "Displ"
	dataType DataType = "Data"
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

func parseBinary(s string) uint32 {
	var v uint32
	for _, c := range s {
		v <<= 1
		if c == '1' {
			v |= 1
		}
	}
	return v
}

func B(bits string) Field {
	return Field{Literal, uint8(len(bits)), 0, uint8(parseBinary(bits)), true}
}

func D() Field {
	return Field{dType, 1, 0, 0, false}
}

func W() Field {
	return Field{wType, 1, 0, 0, false}
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

func DATA_IF_W() Field {
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
	return Field{dispType, 0, 0, 0, false}
}

var Commands = []Command{
	{"mov", []Field{B("100010"), D(), W(), MOD(), REG(), RM()}},
	{"mov", []Field{B("1100011"), W(), MOD(), B("000"), RM(), DATA(), DATA_IF_W(), ImpD(0)}},
	{"mov", []Field{B("1011"), W(), REG(), DATA(), DATA_IF_W(), ImpD(1)}},
	{"mov", []Field{B("1010000"), W(), ADDR(), ImpREG(0), ImpMOD(0), ImpRM(0b110), ImpD(1)}},
	{"mov", []Field{B("1010001"), W(), ADDR(), ImpREG(0), ImpMOD(0), ImpRM(0b110), ImpD(0)}},
	{"mov", []Field{B("100011"), D(), B("0"), MOD(), B("0"), SR(), RM()}},
}
