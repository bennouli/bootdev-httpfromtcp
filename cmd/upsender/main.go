package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	raddr, err := net.ResolveUDPAddr("udp", ":42069")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buf := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")

		data, err := buf.ReadString('\n')
		if err != nil {
			// if errors.Is(err, io.EOF) {
			//
			// }
			log.Fatal(err)
		}

		conn.Write([]byte(data))
	}

}
