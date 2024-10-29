package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type client chan<- string // canal de mensagem

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
	private  = make(chan string)
	chans    = make(map[string]client)
)

func reverse(str string) (result string) {
	for _, v := range str {
		result = string(v) + result
	}
	return
}

func broadcaster() {
	clients := make(map[client]bool) // todos os clientes conectados

	for {
		select {
    case cli := <-entering:
			clients[cli] = true
		case msg := <-messages:
			// Envio para todos
			for cli := range clients {
				cli <- msg
			}
		case msg := <-private:
			messagePrivate := strings.SplitN(msg, " ", 4)
			transmissor := messagePrivate[0]
			receptor := messagePrivate[2]
			message := messagePrivate[3]
			enviada := false
			for key, _ := range clients {
				if key == chans[receptor] && receptor != "bot"{
					enviada = true
					chans[receptor] <- transmissor + " disse em privado: " + message
					break
				}
				if key == chans[receptor] && receptor == "bot"{
					enviada = true
					messageInvert = reverse(message)
					chans[transmissor] <- receptor + "\nrecebi: " + message "\nresposta: " + messageInvert
					break
				}
			}
			if (!enviada){
				chans[transmissor] <- "nenhum cliente encontrado"
			}
    case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		if err != io.EOF {
			fmt.Printf("Read error - %s\n", err)
		}
	}

	apelido := string(buf[:n])
	ch <- "eu sou " + apelido
	messages <- apelido + " chegou!"
	entering <- ch
	chans[apelido] = ch

	input := bufio.NewScanner(conn)
	for input.Scan() {
		cmd := strings.Split(input.Text(), " ")
		comando := cmd[0]

		switch comando {
		case "\\changenick":
			messages <- "usuario " + apelido + " agora se chama " + cmd[1]
			delete(chans, apelido)
			apelido = cmd[1]
			chans[apelido] = ch
		case "\\msg":
			private <- apelido + " " + input.Text()
		case "\\exit":
			leaving <- ch
			messages <- apelido + " se foi "
			delete(chans, apelido)
			return
		case "\\all":
			lista := "online: "
			for k, _ := range chans {
				lista += k + " "
			}
			ch <- lista
		case "\\help":
			ch <- "\\changenick [nick] ( trocar nickname )\n" +
				"\\msg [nick] ( enviar mensagem privada )\n" +
				"\\all ( consultar clientes online )\n" +
				"\\exit ( sair do server )\n"

		default:
			fmt.Println("Enviado uma mensagem")
			messages <- apelido + ": " + input.Text()
		}
	}
	conn.Close()
}

func main() {
	fmt.Println("Iniciando servidor...")
	listener, err := net.Listen("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}
