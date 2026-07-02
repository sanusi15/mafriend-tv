package usecase

import (
	"encoding/json"
	"log"
	"mafriend-tv/internal/model"
	"sync"
)

type MatchUsecase struct {
	waitingClient *model.Client	// Tempat nampung jomblo yg nunggu giliran
	mu			sync.Mutex	// Satpam pengunci biar antran gak rebutan
}

func NewMatchUsecase() *MatchUsecase {
	return &MatchUsecase{}
}

func (u *MatchUsecase) RegisterClient(client *model.Client) {
	u.mu.Lock() // Kunci ruangan!

	// 1. Cek apakah ada orang di ruang tunggu
	if u.waitingClient == nil {
		// Jika kosong, maka orang ini jadi penunggu pertama
		u.waitingClient = client
		u.mu.Unlock() // Buka kunci
		log.Printf("User %s masuk ruang tunggu, jomblo bertambah.", client.ID)

		client.Send <- []byte(`{"status": "waiting", "message": "mencari orang random..."}`)
		return
	}

	// 2. Jika ada orang di ruang tunggu, langsung jodohkan aja!
	peer := u.waitingClient
	u.waitingClient = nil // Kosongkan lagi ruang tunggu untuk menampung si jomblo lagi
	u.mu.Unlock() // Buka kunci segera setelah penjodohan selesai

	// Saling silangkan pasangan mereka
	client.Peer = peer
	peer.Peer = client

	log.Printf("BOOM! User %s berjodoh dengan User %s", client.ID, peer.ID)

	// Beritahu browser masing-masing lewat websocket bahwa mereka sudah dapet pasangan
	resToClient, _ := json.Marshal(map[string]string{"status": "matched", "peer_id": peer.ID, "role": "offerer"})
	resToPeer, _ := json.Marshal(map[string]string{"status": "matched", "peer_id": client.ID, "role": "answerer"})

	client.Send <- resToClient
	peer.Send <- resToPeer
}

func (u *MatchUsecase) RemoveClient(client *model.Client) {
	u.mu.Lock()
	defer u.mu.Unlock()

	// Jika user yang kabur/close browser adalah orang yang lagi nunggu di ruang tunggu, hapus dia
	if u.waitingClient == client {
		u.waitingClient = nil
		log.Printf("User %s keluar dari ruang tunggu sebelum dapet jodoh.", client.ID)
	}

	// Jika dia sudah punya pasangan, putuskan pasangannya dan masukkan mantan pasangannya ke ruang tunggu lagi
	if client.Peer != nil {
		peer := client.Peer
		peer.Peer = nil
		client.Peer = nil

		peer.Send <- []byte(`{"status": "disconnected", "message": "Pasangan Anda kabur, mencari ulang..."}`)
		
		// Otomatis masukkan korban ghosting ke ruang tunggu lagi secara background
		go u.RegisterClient(peer)
	}
}