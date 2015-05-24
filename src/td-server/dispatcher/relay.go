/*
 * Provide data forward
 *
 * Version: dispatch.go 0.1 2014/11/17
 *
 * Author: zhangbao <bao.zhang@godinsec.com>
 *
 */

package dispatcher

import (
	"connect-server/model"
	"fmt"
	"net"
	"time"
)

func SetDeadLine(connA, connB *net.Conn, timeout int64) (err error) {
	if err = (*connA).SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second)); err != nil {
		return err
	}
	return nil
}

//从conna中一次性最多读取4096个字节，如果读取的字节数为0直接返回，之后将conna的数据写入connb
func WriteAtoB(connA, connB *net.Conn, ch chan int, timeout int64) {
	buffer := make([]byte, 4096)
	var n int = 0
	var err error

	for {
		if n, err = (*connA).Read(buffer); n == 0 || err != nil {
			fmt.Println("a ->b ", n, err)
			ch <- 1
			return
		}

		if err := SetDeadLine(connA, connB, timeout); err != nil {
			ch <- 1
			return
		}

		if n, err = (*connA).Write(buffer[:n]); err != nil {
			ch <- 1
			return
		}
	}
}

//从connb中一次性最多读取4096个字节，如果读取的字节数为0直接返回，之后将connb的数据写conna
func WriteBtoA(connA, connB *net.Conn, ch chan int, timeout int64) {
	buffer := make([]byte, 4096)
	var n int = 0
	var err error

	for {
		if n, err = (*connB).Read(buffer); n == 0 || err != nil {
			fmt.Println("b ->a ", n, err)
			ch <- 1
			return
		}

		if err := SetDeadLine(connA, connB, timeout); err != nil {
			ch <- 1
			return
		}

		if n, err = (*connA).Write(buffer[:n]); err != nil {
			ch <- 1
			return
		}
	}
}

func DispatchWithTcp(connA, connB *net.Conn, ch chan *model.RelatedV, chanValue *model.RelatedV, timeout int64) {
	wait := make(chan int)
	//设置conna的超时时间
	if err := SetDeadLine(connA, connB, timeout); err != nil {
		return
	}
	fmt.Println("connA remoteaddr: ", (*connA).RemoteAddr(), "connB remotaddr : ", (*connB).RemoteAddr())
	//从conna中读取数据再向connb中写入数据
	go WriteAtoB(connA, connB, wait, timeout)
	//从connb中读取数据再向conna中写入数据
	go WriteBtoA(connA, connB, wait, timeout)

	<-wait
	ch <- chanValue
	return
}
