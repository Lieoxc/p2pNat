package visitor

import (
	"net"
	"p2pNat/client/common"

	"github.com/Lieoxc/log"
)

//访问者，通过监听本地代理地址，将收到的数据转发到Nat洞中
func Run(visitor common.VisitorSt) {

	listener, err := net.Listen("tcp4", visitor.LocalIP+":"+visitor.LocalPort)
	if err != nil {
		log.Error("Listen tcp server failed,err:", err)
		return
	}
	for {
		// 建立socket连接
		tcpConnect, err := listener.Accept()
		if err != nil {
			log.Error("Listen.Accept failed,err:", err)
			return
		}
		go visitorRead(tcpConnect)
		go visitorWrite(tcpConnect)
	}

}

func visitorRead(tcpConnect net.Conn) {
	for {
		data := make([]byte, 1024*128)
		n, err := tcpConnect.Read(data)
		if err != nil {
			log.Error("Read from tcp server failed,err:", err)
			break
		}
		common.WriteChannel <- data[:n] //将127.0.0.1:6000上面收到的数据放入到 writeChannel，通知UDP发送
	}
}
func visitorWrite(tcpConnect net.Conn) {
	for {
		msg := <-common.ReadChannel
		_, err := tcpConnect.Write(msg)
		if err != nil {
			log.Error("Write failed,err:", err)
		}
	}
}
