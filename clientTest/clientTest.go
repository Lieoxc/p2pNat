package main

import (
	"flag"
	"fmt"
	"net"
	"time"
)

/*
	探测内网的客户端是哪一种NAT类型
	mode : 	normal 作为普通UDP客户端 host和port 为公网服务器的地址和端口
			NAT    作为NAT探测,验证是否为  端口限制型NAT
*/
func main() {
	host := flag.String("h", "192.168.1.1", "an int")
	port := flag.String("p", "8080", "an int")
	mode := flag.String("m", "normal", "mode")
	flag.Parse()
	if *mode == "Nat" {
		NatTEST(*host, *port)
	} else {
		Server(*host, *port)
	}

}
func Server(host, port string) {
	publicServerAddr, err := net.ResolveUDPAddr("udp4", host+":"+port)
	if err != nil {
		fmt.Println("ResolveUDPAddr:", err)
		return
	}
	socket, err := net.ListenUDP("udp4", publicServerAddr)
	if err != nil {
		fmt.Println("监听失败!", err)
		return
	}
	defer socket.Close()
	for {
		// 读取数据
		data := make([]byte, 4096)
		_, remoteAddr, err := socket.ReadFromUDP(data)
		if err != nil {
			fmt.Println("读取数据失败!", err)
			continue
		}
		dataStr := string(data)
		host := remoteAddr.IP.String()
		port := remoteAddr.Port

		fmt.Println("recv Data:", dataStr, host, port)

		dataInfo := "serverTest ACK"
		for i := 0; i < 100; i++ {
			_, err = socket.WriteToUDP([]byte(dataInfo), remoteAddr)
			if err != nil {
				fmt.Println("发送数据失败!", err)
				return
			}
			time.Sleep(time.Duration(1) * time.Second)
		}
	}
}

/*
	验证是否为端口限制型NAT
	TODO ：还需要增加其他类型NAT的检测
*/
func NatTEST(host, port string) {
	socket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 9090,
	})
	if err != nil {
		fmt.Println("监听失败!", err)
		return
	}
	defer socket.Close()
	dataInfo := "NAT TEST"
	remoteAddr, err := net.ResolveUDPAddr("udp4", host+":"+port)
	if err != nil {
		fmt.Println("ResolveUDPAddr:", err)
		return
	}

	for i := 0; i < 100; i++ {
		_, err = socket.WriteToUDP([]byte(dataInfo), remoteAddr)
		if err != nil {
			fmt.Println("发送数据失败!", err)
			return
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
}
