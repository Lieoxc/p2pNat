# p2pNat
p2p NAT穿透工具（UDP方式实现）
## client
NAT内网主机端运行代码，当前实现效果为UDP穿透，两个位于不同NAT下的内网主机，可以互相通信
## server
服务端代码，部署在公网服务器，最为探测发现，消息传递
## clientTest
客户端测试，测试当前主机网络属于哪一种NAT类型，完全对称型NAT不能够穿透，大部分家用网络都是端口限制型NAT
