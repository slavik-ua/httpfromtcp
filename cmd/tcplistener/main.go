package main

import (
	"fmt"
	"io"
	"net"
	"log"
	"time"
	"strings"
)


func getLinesChannel(conn net.Conn) <-chan string {
	
	ch := make(chan string)
	go func() {
		defer close(ch)
		defer conn.Close()
		// result = parts[len(parts)-1]
		
		var result string
		buffer := make([]byte, 8)
		for {
			conn.SetReadDeadline(time.Now().Add(10 * time.Second))
			n, err := conn.Read(buffer)
			
			if n > 0 {
				result += string(buffer[:n])
				parts := strings.Split(result, "\n")

				for i := 0; i < len(parts) - 1; i++ {
					ch <- parts[i]
				}

				result = parts[len(parts)-1]
			}

			if err != nil {
				if err != io.EOF {
					fmt.Println("Error reading:", err)
				}
				if result != "" {
					ch <- result
				}
				return
			}
		}

	}()
	return ch
}

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Println("Connectio has been accepted")

		for line := range getLinesChannel(conn) {
			fmt.Printf("read: %s\n", line)
		}
	}
}