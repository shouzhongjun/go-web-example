package utils

import "github.com/rs/xid"

// Xid 生成一个唯一的字符串标识符。
// 该函数使用xid库创建一个新的标识符实例，并将其转换为字符串形式。
// 无需传递参数。
// 返回值是一个字符串类型的唯一标识符。
func Xid() string {
	id := xid.New()
	return id.String()
}

// XidBytes 生成一个唯一的标识符，并以字节切片的形式返回。
// 该函数使用 xid 库创建一个新的唯一标识符，然后调用 Bytes 方法将该标识符转换为字节切片。
// 这个函数在需要唯一标识符的字节表示的场景中特别有用。
//
// 返回值:
//
//	[]byte: 唯一标识符的字节切片表示。
func XidBytes() []byte {
	id := xid.New()
	return id.Bytes()
}
