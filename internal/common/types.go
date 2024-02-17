package common

type memoryReader interface {
	ReadFromAddress(address uint16, ammount int) (result []byte, err error)
}

type memoryWriter interface {
	WriteToAddress(address uint16, bytes []byte) (err error)
}

type MemoryReadWriter interface {
	memoryReader
	memoryWriter
}
