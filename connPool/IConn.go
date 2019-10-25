/*
@Time : 2019/9/26 16:47
@Author : Hermes
@File : Conn
@Description:
*/
package connPool

type Conn interface {
	Close() error
	SendRecv(buff []byte, timeout int) ([]byte, error)
}
