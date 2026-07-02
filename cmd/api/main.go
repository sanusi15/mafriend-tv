package main

import (
	"log"
	"mafriend-tv/internal/handler"
	"mafriend-tv/internal/usecase"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor" // 💡 Adaptor resmi Fiber v3
)

func main() {
	app := fiber.New()

	// Inisialisasi Modul Pencari Jodoh Random
	matchUsecase := usecase.NewMatchUsecase()
	matchHandler := handler.NewMatchHandler(matchUsecase)

	// Ubah handler net/http Gorilla agar bisa dijalankan di rute Fiber v3
	app.Get("/ws-ome", adaptor.HTTPHandlerFunc(matchHandler.HandleGorillaWS))

	log.Println("Server OmeTV Backend berjalan di port 3000...")
	log.Fatal(app.Listen(":3000"))
}