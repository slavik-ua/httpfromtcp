package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	fmt.Println("HERE")
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Connection failed:", err)
		return
	}
	defer conn.Close()

	scanner := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("> ")
		line, err := scanner.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println(line)
			}
			break
		}
		_, err = conn.Write([]byte(line))
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}
}
