/*
@Time : 2019/9/26 16:47
@Author : Hermes
@File : Conn
@Description:
Package pool是根据net.Conn接口能力来适配实现的，任何实现了conn到close()方法到实例都可以适配
*/
package connPool

import (
	"errors"
)

var (
	// ErrClosed是因为pool.Close()已经执行后，再来操作引发的错误
	ErrClosed = errors.New("pool is closed")
)

// Pool应该有一些参数，比如最大容量，初始连接数量，pool必须实现并发安全，而且要易用
type Pool interface {
	Get() (*PoolConn, error)

	//关闭pool和他里面的所有连接，关闭后不可用了
	Close()

	//返回现在里面还有多少连接是可用的
	Len() int
}
