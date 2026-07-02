package handler

import (
	"fmt"
	"log"
	"mafriend-tv/internal/model"
	"mafriend-tv/internal/usecase"
	"math/rand"
	"net/http"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// Gunakan Upgrader Gorilla standard resmi
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Bypass CORS agar bisa ditembak dari React JS localhost
	CheckOrigin: func(r *http.Request) bool { return true },
}

type MatchHandler struct {
	usecase *usecase.MatchUsecase
	db *gorm.DB
}

func NewMatchHandler(u *usecase.MatchUsecase, db *gorm.DB) *MatchHandler {
	return &MatchHandler{
		usecase: u,
		db: db,
	}
}

// 💡 INI ADALAH STANDARD HANDLER NET/HTTP BIASA
func (h *MatchHandler) HandleGorillaWS(w http.ResponseWriter, r *http.Request) {
	// 💡 Ambil GoogleID aman hasil bongkar JWT dari header request
	googleID := r.Header.Get("X-User-GoogleID")

	// Tarik data profil aslinya dari DB untuk dipajang saat match
	var user model.User
	h.db.Where("google_id = ?", googleID).First(&user)

	// Upgrade koneksi HTTP murni menjadi WebSocket murni via Gorilla
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Gagal upgrade via Gorilla:", err)
		return
	}

	// Generate ID acak untuk user sementara
	userID := fmt.Sprintf("User-%d", rand.Intn(9000)+1000)

	client := &model.Client{
		ID:   userID,
		Conn: conn, // Masukkan koneksi gorilla resmi ke struct model kita
		Send: make(chan []byte, 256),
	}

	// Goroutine Writer: Mengirim data dari Go ke Browser
	go func() {
		defer conn.Close()
		for message := range client.Send {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				break
			}
		}
	}()

	// Daftarkan user ke sistem pencari jodoh kita di usecase
	h.usecase.RegisterClient(client)

	// Goroutine Reader: Menunggu data masuk dari browser
	defer func() {
		h.usecase.RemoveClient(client)
		conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break // Jika tab browser diclose atau disconnect, keluar dari loop
		}

		// Jika sudah punya pasangan, oper pesannya langsung ke pasangannya (Signaling WebRTC)
		if client.Peer != nil {
			client.Peer.Send <- message
		}
	}
}