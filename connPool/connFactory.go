/*
@Time : 2019/9/26 19:45
@Author : Hermes
@File : connFactory
@Description:
*/
package connPool

type ConnFactory interface {
	Create() (Conn, error)
}
