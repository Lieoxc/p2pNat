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
	wg sync.WaitGroup
)

func main() {
	localAddr.Port = int(RandPort(10000, 20000))
	host := flag.String("h", "110.42.255.112", "an int")
	port := flag.String("p", "9090", "an int")
	flag.Parse()
	// 创建连接
	clientTest(*host, *port)
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
			p2pdetect(nodeHost, nodePort)
		}
	}
}

// 监听localAddr 和定时发送数据到 host:port(NAT 打洞)
func p2pdetect(host, port string) {
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
	for i := 0; i < 300; i++ {
		detectInfo := "p2p detect"
		detectInfo = fmt.Sprintf("%s:%d", detectInfo, i)
		senddata := []byte(detectInfo)
		_, err := p2pConnect.WriteToUDP(senddata, remoteAddr)
		if err != nil {
			fmt.Println("发送数据失败!", err)
			return
		}
		time.Sleep(time.Duration(3) * time.Second)
	}
}
func p2pRead(p2pConnect *net.UDPConn) {
	for {
		// 读取数据
		data := make([]byte, 4096)
		_, remoteAddr, err := p2pConnect.ReadFromUDP(data)
		if err != nil {
			fmt.Println("读取数据失败!", err)
			continue
		}
		dataStr := string(data)

		fmt.Println("**** p2p READ:", dataStr, remoteAddr.IP.String(),
			remoteAddr.Port)
	}
}

// RandPort 生成区间范围内的随机端口
func RandPort(min, max int64) int64 {
	if min > max {
		panic("the min is greater than max!")
	}
	if min < 0 {
		f64Min := math.Abs(float64(min))
		i64Min := int64(f64Min)
		result, _ := rand.Int(rand.Reader, big.NewInt(max+1+i64Min))
		return result.Int64() - i64Min
	}
	result, _ := rand.Int(rand.Reader, big.NewInt(max-min+1))
	return min + result.Int64()
}
