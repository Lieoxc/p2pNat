package visited

import (
	"net"
	"p2pNat/client/common"
	"sync"

	"github.com/Lieoxc/log"
)

var (
	wg sync.WaitGroup
)

//被访问者，TCP 客户端，连接 127.0.0.1：80地址
func Run(visited common.VisitedSt) {
	tcpConnect, err := net.Dial("tcp", visited.BindIP+":"+visited.BindPort)
	if err != nil {
		log.Error("Connect to TCP server failed ,err:", err)
		return
	}

	// 建立socket连接
	// 业务处理逻辑
	defer tcpConnect.Close()
	wg.Add(2)
	go visitedRead(tcpConnect)
	go visitedWrite(tcpConnect)
	wg.Wait()
}
func visitedRead(tcpConnect net.Conn) {
	for {
		data := make([]byte, 1024*1024*128)
		n, err := tcpConnect.Read(data)
		if err != nil {
			log.Error("Read from tcp server failed,err:", err)
			break
		}
		common.WriteChannel <- data[:n] //上读到数据放入到 writeChannel，通知UDP发送
	}
}
func visitedWrite(tcpConnect net.Conn) {
	for {
		msg := <-common.ReadChannel
		_, err := tcpConnect.Write(msg) //将p2p Nat 传过来的数据写入到 tcpConnect
		if err != nil {
			log.Error("Write failed,err:", err)
		}
	}

}
