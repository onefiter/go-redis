package dict

type Consumer func(key string, val interface{}) bool

// Dict is interface of a key-value data structure
type Dict interface {
	Get(key string) (val interface{}, exists bool)
	Len() int
	Put(key string, val interface{}) (result int)
	PutIfAbsent(key string, val interface{}) (result int)
	PutIfExists(key string, val interface{}) (result int)
	Remove(key string) (result int)
	ForEach(consumer Consumer)
	Keys() []string
	RandomKeys(limit int) []string
	RandomDistinctKeys(limit int) []string // 返回多个不重复的键
	Clear()
}
