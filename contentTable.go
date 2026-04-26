package main

type DataType string

const (
	Literal  DataType = "Literal"
	wType    DataType = "W"
	dType    DataType = "D"
	regType  DataType = "REG"
	modType  DataType = "MOD"
	rmType   DataType = "RM"
	dispType DataType = "Displ"
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

func REG() Field {
	return Field{regType, 3, 0, 0, false}
}

func RM() Field {
	return Field{rmType, 3, 0, 0, false}
}

func MOD() Field {
	return Field{modType, 2, 0, 0, false}
}

var Commands = []Command{
	{"mov", []Field{B("100010"), D(), W(), MOD(), REG(), RM()}},
}
