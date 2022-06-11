package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	hostA = ""
	portA = 0

	hostB = ""
	portB = 0

	remoteAddrA *net.UDPAddr
	remoteAddrB *net.UDPAddr
)

func main() {
	// 创建监听
	socket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 9090,
	})
	if err != nil {
		fmt.Println("监听失败!", err)
		return
	}
	count := 0
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

		fmt.Println("recv Data:", dataStr)
		if strings.Contains(dataStr, "client") {
			if count%2 == 0 {
				hostA = remoteAddr.IP.String()
				portA = remoteAddr.Port
				remoteAddrA = remoteAddr
			} else {
				hostB = remoteAddr.IP.String()
				portB = remoteAddr.Port
				remoteAddrB = remoteAddr
			}
			count++
		}
		if portB != 0 && portA != 0 {
			fmt.Println("collect PCA and PCB address info OK !")
			// 发送PCB 的NAT信息给PCA
			natInfoB := "Nat:" + hostB + ":" + strconv.Itoa(portB) + ":"
			senddata := []byte(natInfoB)
			_, err = socket.WriteToUDP(senddata, remoteAddrA)
			if err != nil {
				fmt.Println("发送数据失败!", err)
				return
			}
			// 发送PCA 的NAT信息给PCB
			natInfoA := "Nat:" + hostA + ":" + strconv.Itoa(portA) + ":"
			senddata = []byte(natInfoA)
			_, err = socket.WriteToUDP(senddata, remoteAddrB)
			if err != nil {
				fmt.Println("发送数据失败!", err)
				return
			}

			//延时5s  控制 两个客户端互相发送报文
			time.Sleep(time.Duration(3) * time.Second)
			_, err = socket.WriteToUDP([]byte("control"), remoteAddrA)
			if err != nil {
				fmt.Println("发送数据失败!", err)
				return
			}
			//延时2s  控制PCB 发送一个探测响应报文
			time.Sleep(time.Duration(1) * time.Second)
			_, err = socket.WriteToUDP([]byte("control"), remoteAddrB)
			if err != nil {
				fmt.Println("发送数据失败!", err)
				return
			}
		}

	}
}
