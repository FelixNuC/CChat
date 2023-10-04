package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var clients = make(map[*websocket.Conn]string)

func handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error al hacer upgrade:", err)
		return
	}
	fmt.Println("Nuevo cliente conectado:", conn.RemoteAddr())
	defer conn.Close()

	welcomeMsg := "Por favor, introduce tu nombre de usuario:"
	if err := conn.WriteMessage(websocket.TextMessage, []byte(welcomeMsg)); err != nil {
		fmt.Println("Error al enviar mensaje de bienvenida:", err)
		return
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		fmt.Println("Error al leer mensaje:", err)
		return
	}
	username := string(msg)
	clients[conn] = username
	fmt.Println(username + " se ha unido al chat.")

	welcomeUserMsg := fmt.Sprintf("Bienvenido al chat, %s!", username)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(welcomeUserMsg)); err != nil {
		fmt.Println("Error al enviar mensaje de bienvenida al usuario:", err)
		return
	}
	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error al leer mensaje:", err)
			delete(clients, conn)
			return
		}

		fullMessage := fmt.Sprintf("%s: %s", username, string(msg))

		for client, clientUsername := range clients {
			if client != conn {
				if err := client.WriteMessage(messageType, []byte(fullMessage)); err != nil {
					fmt.Println("Error al escribir mensaje:", err)
					client.Close()
					delete(clients, client)
				}
			} else {
				fmt.Printf("Mensaje de %s: %s\n", clientUsername, string(msg))
			}
		}
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hola, mundo")
	})
	http.HandleFunc("/ws", handleConnection)

	fmt.Println("Servidor corriendo en el puerto 8080")
	http.ListenAndServe("0.0.0.0:8080", nil)
}
