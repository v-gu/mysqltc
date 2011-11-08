package mtclib

import (
	"fmt"
	"strings"
	"strconv"
)

type MySQLServer struct {
	Host string
	Port int
	User string
	Pass string
}

func ParseNid(nidStr string) *MySQLServer {
	defer func() {
		if err := recover(); err != nil {
			panic(fmt.Sprintf("ParseNid('%v')->%v", nidStr, err))
		}
	}()

	server := &MySQLServer{
		Host: "localhost",
		Port: 3306,
		User: "rpl",
		Pass: "pass"}
	for _, token := range strings.Split(nidStr, ",") {
		kv := strings.Split(token, "=")
		if len(kv) != 2 {
			panic(fmt.Sprintf("mulformed NID: '%v'", token))
		}
		switch kv[0] {
		case "h":
			server.Host = kv[1]
		case "P":
			port, err := strconv.Atoi(kv[1])
			if err != nil {
				panic(fmt.Sprintf("port should be a number: %v", server.Port))
			}
			if server.Port < 0 || server.Port > 65535 {
				panic(fmt.Sprintf("incorrect port range: %v", server.Port))
			}
			server.Port = port
		case "u":
			server.User = kv[1]
		case "p":
			server.Pass = kv[1]
		default:
			panic(fmt.Sprintf("mulformed NID: %v=%v", kv[0], kv[1]))
		}
	}
	return server
}
