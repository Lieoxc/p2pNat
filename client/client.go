package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"math"
	"math/big"
	"net"
	"strings"
	"sync"
	"time"
)

var ( //p2p对端的NAT外网地址和端口
	nodeHost  = ""
	nodePort  = ""
	localAddr = &net.UDPAddr{
		IP: net.IPv4(0, 0, 0, 0),
	}
	wg           sync.WaitGroup
	readChannel  chan []byte
	writeChannel chan []byte //需要写数据到 p2p 对端的通道
)

func main() {
	localAddr.Port = RandPort(10000, 20000)
	mode := flag.String("m", "visitor", "p2p client mode")
	host := flag.String("h", "110.42.255.112", "an int")
	port := flag.String("p", "9090", "an int")

	readChannel = make(chan []byte, 16)
	writeChannel = make(chan []byte, 16)

	flag.Parse()
	if *mode == "visitor" { //访问者模式
		go visitor()
	} else { //被访问客户端
		go visited()
	}
	// 创建连接
	clientTest(*host, *port)
}

//访问者，通过监听本地 127.0.0.1:6000 地址，将收到的数据转发到Nat洞中
func visitor() {
	listener, err := net.Listen("tcp4", "127.0.0.1:6000")
	if err != nil {
		fmt.Println("Listen tcp server failed,err:", err)
		return
	}

	// 建立socket连接
	tcpConnect, err := listener.Accept()
	if err != nil {
		fmt.Println("Listen.Accept failed,err:", err)
		return
	}

	// 业务处理逻辑
	defer tcpConnect.Close()
	wg.Add(2)
	go visitorRead(tcpConnect)
	go visitorWrite(tcpConnect)
	wg.Wait()
}
func visitorRead(tcpConnect net.Conn) {
	for {
		data := make([]byte, 1024*128)
		n, err := tcpConnect.Read(data)
		if err != nil {
			fmt.Println("Read from tcp server failed,err:", err)
			break
		}
		writeChannel <- data[:n] //将127.0.0.1:6000上面收到的数据放入到 writeChannel，通知UDP发送
	}
}
func visitorWrite(tcpConnect net.Conn) {
	for {
		msg := <-readChannel
		_, err := tcpConnect.Write(msg)
		if err != nil {
			fmt.Println("Write failed,err:", err)
		}
	}
}

//被访问者，TCP 客户端，连接 127.0.0.1：80地址
func visited() {
	tcpConnect, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println("Connect to TCP server failed ,err:", err)
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
		data := make([]byte, 1024*128)
		n, err := tcpConnect.Read(data) //从127.0.0.1:80读到数据
		if err != nil {
			fmt.Println("Read from tcp server failed,err:", err)
			break
		}
		writeChannel <- data[:n] //上读到数据放入到 writeChannel，通知UDP发送
	}
}
func visitedWrite(tcpConnect net.Conn) {
	for {
		msg := <-readChannel
		_, err := tcpConnect.Write(msg) //将p2p Nat 传过来的数据写入到 tcpConnect（127.0.0.1：80）
		if err != nil {
			fmt.Println("Write failed,err:", err)
		}
	}

}

func clientTest(host, port string) {
	remoteAddr, err := net.ResolveUDPAddr("udp4", host+":"+port)
	if err != nil {
		fmt.Println("ResolveUDPAddr:", err)
		return
	}
	socket, err := net.DialUDP("udp4", localAddr, remoteAddr)
	if err != nil {
		fmt.Println("连接失败!", err)
		return
	}
	defer socket.Close()
	senddata := []byte("hello server,I am client!")
	_, err = socket.Write(senddata)
	if err != nil {
		fmt.Println("发送数据失败!", err)
		return
	}
	for {
		// 读取数据
		data := make([]byte, 128)
		_, _, err := socket.ReadFromUDP(data)
		if err != nil {
			fmt.Println("读取数据失败!", err)
			continue
		}
		recvData := string(data)

		fmt.Println("recv Data:", recvData)
		//服务器发过来 p2p 对端的 NAT外网地址和端口  eg: Nat:121.32.254.146:51279
		nodeKey := "Nat:"
		if strings.Contains(recvData, nodeKey) {
			strs := strings.Split(recvData, ":")
			nodeHost = strs[1]
			nodePort = strs[2]
			fmt.Println("PCB node info finsh:", nodeHost, nodePort)
		}
		// 客户端收到 controlA 指令 发送detect报文给PCB
		if strings.Contains(recvData, "control") {
			//正常来讲此数据不会被PCB收到，但是能够在PCA的NAT网关起到一个记录nodeHost，nodePort的
			socket.Close() //关闭 localAddr和公网服务器的连接,这个地址需要用来与p2p主机进行通信
			p2pHandler(nodeHost, nodePort)
		}
	}
}

// 监听localAddr 和定时发送数据到 host:port(NAT 打洞)
func p2pHandler(host, port string) {
	remoteAddr, err := net.ResolveUDPAddr("udp4", host+":"+port)
	if err != nil {
		fmt.Println("ResolveUDPAddr:", err)
		return
	}
	p2pConnect, err := net.ListenUDP("udp4", localAddr)
	if err != nil {
		fmt.Println("监听失败!", err)
		return
	}

	defer p2pConnect.Close()
	wg.Add(2)
	go p2pRead(p2pConnect)
	go p2pSend(p2pConnect, remoteAddr)
	wg.Wait()
}

func p2pSend(p2pConnect *net.UDPConn, remoteAddr *net.UDPAddr) {
	i := 0
	cgTicker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-cgTicker.C: //定时发送心跳包，保证 Nat 洞的存活性
			i++
			detectInfo := "p2pHeartbeat"
			senddata := []byte(detectInfo)
			_, err := p2pConnect.WriteToUDP(senddata, remoteAddr)
			if err != nil {
				fmt.Println("发送数据失败!", err)
				return
			}
		case msg := <-writeChannel: //从本地代理地址收到数据,发送到对端Nat 的洞中
			_, err := p2pConnect.WriteToUDP(msg, remoteAddr)
			if err != nil {
				fmt.Println("发送数据失败!", err)
				return
			}
		}
	}
}
func p2pRead(p2pConnect *net.UDPConn) {
	for {
		// 读取数据
		data := make([]byte, 1024*128)
		n, _, err := p2pConnect.ReadFromUDP(data)
		if err != nil {
			fmt.Println("读取数据失败!", err)
			continue
		}

		if n < 16 { //小于16则认为是心跳包
			fmt.Println("recv p2pHeartbeat package :", string(data[:n]))
		} else { //大于16 则认为是实际的数据
			readChannel <- data[:n]
		}
	}
}

// RandPort 生成区间范围内的随机端口
func RandPort(min, max int64) int {
	if min > max {
		panic("the min is greater than max!")
	}
	if min < 0 {
		f64Min := math.Abs(float64(min))
		i64Min := int64(f64Min)
		result, _ := rand.Int(rand.Reader, big.NewInt(max+1+i64Min))
		return int(result.Int64() - i64Min)
	}
	result, _ := rand.Int(rand.Reader, big.NewInt(max-min+1))
	return int(min + result.Int64())
}
