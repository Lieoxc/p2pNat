package main

import (
	"flag"
	"net"
	"p2pNat/client/common"
	"p2pNat/client/config"
	"p2pNat/client/utils"
	"p2pNat/client/visited"
	"p2pNat/client/visitor"
	"strings"
	"sync"
	"time"

	"github.com/Lieoxc/log"
)

var ( //p2p对端的NAT外网地址和端口

	localAddr = &net.UDPAddr{
		IP: net.IPv4(0, 0, 0, 0),
	}

	wg sync.WaitGroup
)
var configFile = flag.String("f", "./client_visitor.yaml", "the config file")

func main() {
	log.Debug("p2p Nate Start")

	flag.Parse()

	localAddr.Port = utils.RandPort(10000, 20000)
	cfg, err := config.ReadCfg(*configFile)
	if err != nil {
		log.Error("ReadCfg error:", err)
		return
	}
	common.WriteChannel = make(chan []byte, 10)
	common.ReadChannel = make(chan []byte, 10)
	if cfg.Visitor.VisitedName != "" { //访问者模式
		go visitor.Run(cfg.Visitor)
	} else { //被访问客户端
		go visited.Run(cfg.Visited)
	}
	// 创建连接
	clientTest(cfg.Server.Host, cfg.Server.Port)
	log.Flush()
}

func clientTest(host, port string) {
	remoteAddr, err := net.ResolveUDPAddr("udp4", host+":"+port)
	if err != nil {
		log.Error("ResolveUDPAddr:", err)
		return
	}
	socket, err := net.DialUDP("udp4", localAddr, remoteAddr)
	if err != nil {
		log.Error("连接失败!", err)
		return
	}
	defer socket.Close()
	senddata := []byte("hello server,I am client!")
	_, err = socket.Write(senddata)
	if err != nil {
		log.Error("发送数据失败!", err)
		return
	}
	peerHost := ""
	peerPort := ""
	for {
		// 读取数据
		data := make([]byte, 128)
		_, _, err := socket.ReadFromUDP(data)
		if err != nil {
			log.Error("读取数据失败!", err)
			continue
		}
		recvData := string(data)
		//服务器发过来 p2p 对端的 NAT外网地址和端口
		nodeKey := "Nat:"
		if strings.Contains(recvData, nodeKey) {
			strs := strings.Split(recvData, ":")
			peerHost, peerPort = strs[1], strs[2]
		}
		// 客户端收到 controlA 指令 发送detect报文给PCB
		if strings.Contains(recvData, "control") {
			socket.Close() //关闭 localAddr和公网服务器的连接,这个地址需要用来与p2p主机进行通信
			p2pHandler(peerHost, peerPort)
		}
	}
}

// 监听localAddr 和定时发送数据到 host:port(NAT 打洞)
func p2pHandler(host, port string) {
	remoteAddr, err := net.ResolveUDPAddr("udp4", host+":"+port)
	if err != nil {
		log.Error("ResolveUDPAddr:", err)
		return
	}
	p2pConnect, err := net.ListenUDP("udp4", localAddr)
	if err != nil {
		log.Error("监听失败!", err)
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
	cgTicker := time.NewTicker(time.Second * 2)
	for {
		select {
		case <-cgTicker.C: //定时发送心跳包，保证 Nat 洞的存活性
			i++
			detectInfo := "p2pHeartbeat"
			senddata := []byte(detectInfo)
			_, err := p2pConnect.WriteToUDP(senddata, remoteAddr)
			if err != nil {
				log.Error("发送数据失败!", err)
				return
			}
		case msg := <-common.WriteChannel: //从本地代理地址收到数据,发送到对端Nat 的洞中
			_, err := p2pConnect.WriteToUDP(msg, remoteAddr)
			if err != nil {
				log.Error("发送数据失败!", err)
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
			log.Error("读取数据失败!", err)
			continue
		}

		if n < 16 { //小于16则认为是心跳包
			log.Error("recv p2pHeartbeat package :", string(data[:n]))
		} else { //大于16 则认为是实际的数据
			common.ReadChannel <- data[:n]
		}
	}
}
