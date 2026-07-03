package main

import (
	"log"
	"mafriend-tv/internal/config"
	"mafriend-tv/internal/handler"
	"mafriend-tv/internal/usecase"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor" // 💡 Adaptor resmi Fiber v3
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// load env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Gagal load env")	
	}

	db := config.InitDB()

	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	app := fiber.New()


	// Setting CORS ORIGIN
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3001", "http://localhost:5173", "https://mafriendtv.skuycode.my.id"}, // URL React kamu
		AllowHeaders: []string{"Origin, Content-Type", "Accept", "Authorization"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))

	app.Use(logger.New())

	authHandler := handler.NewAuthHandler(db)

	// Inisialisasi Modul Pencari Jodoh Random
	matchUsecase := usecase.NewMatchUsecase()
	matchHandler := handler.NewMatchHandler(matchUsecase, db)


	// Jalur API HTTP untuk login Google
	app.Get("/api/auth/google", authHandler.GetLoginURL)
	app.Post("/api/auth/callback", authHandler.HandleCallback)

	// Ubah handler net/http Gorilla agar bisa dijalankan di rute Fiber v3
	app.Get("/ws-mafriend", adaptor.HTTPHandlerFunc(matchHandler.HandleGorillaWS))

	log.Println("Server MaFriendTv Backend berjalan di port 3000...")
	log.Fatal(app.Listen(":3000"))
}