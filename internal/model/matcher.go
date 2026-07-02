package model

import "github.com/gorilla/websocket"

type Client struct {
	ID        string          `json:"id"`         // Diisi oleh googleUser["sub"] (Permanen)
	Name      string          `json:"name"`       // Diisi oleh googleUser["name"]
	Picture   string          `json:"picture"`    // Diisi oleh googleUser["picture"]
	Conn *websocket.Conn 	`json:"-"`	// Koneksi websocket murni ke browser dia
	Peer *Client 			`json:"-"`	// Pasangan/jodoh random dia saat video call
	Send chan []byte 		`json:"-"`	// Channel internal untuk ngirim pesan ke user ini
}