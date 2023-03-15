package resp

// Reply is the interface of redis serialization protocol message
// TCP just deal with read and write bytes
type Reply interface {
	ToBytes() []byte
}
