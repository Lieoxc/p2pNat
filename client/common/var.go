package common

type Config struct {
	Server  ServerSt  `yaml:"server"`
	Visited VisitedSt `yaml:"visited"`
	Visitor VisitorSt `yaml:"visitor"`
}
type ServerSt struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}
type VisitedSt struct {
	Name     string `yaml:"name"`
	BindIP   string `yaml:"bindIP"`
	BindPort string `yaml:"bindPort"`
}

type VisitorSt struct {
	VisitedName string `yaml:"visitedName"`
	LocalIP     string `yaml:"localIP"`
	LocalPort   string `yaml:"localPort"`
}

var (
	ReadChannel  chan []byte
	WriteChannel chan []byte //需要写数据到 p2p 对端的通道
)
