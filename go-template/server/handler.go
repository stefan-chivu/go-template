package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	}}

func (s *Server) home(w http.ResponseWriter, req *http.Request) {
	ws, err := Upgrade(w, req)

	if err != nil {
		fmt.Fprintf(w, "%+V\n", err)
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, "Parse form failed", http.StatusBadRequest)
		return
	}

	username := req.Form.Get("username")

	// TODO better valid username check
	if username == "" {
		// error case
		http.Error(w, "Invalid username", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.Clients[ws] = username
	s.mu.Unlock()

	log.Default().Println("Connected new client from: " + req.RemoteAddr + "; Username: " + username)
	s.reader(ws)
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return ws, err
	}
	return ws, nil
}

func (s *Server) reader(conn *websocket.Conn) {
	for {
		messageType, buff, err := conn.ReadMessage()
		if err != nil {
			if err == io.EOF {
				continue
			}

			if _, ok := s.Clients[conn]; ok {
				if websocket.IsCloseError(err, websocket.CloseGoingAway) {
					s.handleClose(conn, fmt.Sprintf("%s is going away", s.Clients[conn]))
					break
				}
				if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
					s.handleClose(conn, fmt.Sprintf("%s closed unexpectedly", s.Clients[conn]))
					break
				}
				if websocket.IsCloseError(err, websocket.CloseAbnormalClosure) {
					s.handleClose(conn, fmt.Sprintf("%s closed abnormally", s.Clients[conn]))
					break
				}
			}

			log.Default().Println("Websocket read error", err)
			break
		}
		if err := conn.WriteMessage(messageType, buff); err != nil {
			log.Println(err)
			continue
		}
	}
}

func (s *Server) ServeWs(w http.ResponseWriter, req *http.Request) {
	ws, err := Upgrade(w, req)

	if err != nil {
		fmt.Fprintf(w, "%+V\n", err)
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, "Parse form failed", http.StatusBadRequest)
		return
	}

	username := req.Form.Get("username")

	// TODO better valid username check
	if username == "" {
		// error case
		http.Error(w, "Invalid username", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.Clients[ws] = username
	s.mu.Unlock()

	log.Default().Println("Connected new client from: " + req.RemoteAddr + "; Username: " + username)
}

func (s *Server) getUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userList := []string{}

	for _, v := range s.Clients {
		userList = append(userList, v)
	}

	responseData, err := json.Marshal(userList)

	if err != nil {
		http.Error(w, "User list JSON marshalling failed", http.StatusInternalServerError)
		return
	}

	w.Write(responseData)
}

func httpReqLogMsg(r *http.Request, message string) string {
	return "Src: [ " + r.RemoteAddr + " ] ; Dst: [ " + r.Host + " ] ; Method: " + r.Method + "; " + message
}

func (s *Server) handleClose(ws *websocket.Conn, message string) {
	log.Default().Print(message)
	delete(s.Clients, ws)
}
