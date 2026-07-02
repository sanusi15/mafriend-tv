package handler

import (
	"context"
	"encoding/json"
	"mafriend-tv/internal/model" // 💡 Sesuaikan dengan nama modul di go.mod lu
	"mafriend-tv/internal/util"
	"os"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

type AuthHandler struct {
	oauthConfig *oauth2.Config
	db          *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	googleClientId := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	googleRedirectUrl := os.Getenv("GOOGLE_REDIRECT_URL")

	return &AuthHandler{
		db: db,
		oauthConfig: &oauth2.Config{
			ClientID:     googleClientId, // 💡 Ganti dengan Client ID lu
			ClientSecret: googleClientSecret,                        // 💡 Ganti dengan Client Secret lu
			RedirectURL:  googleRedirectUrl,            // 💡 URL React penampung callback
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.profile",
				"https://www.googleapis.com/auth/userinfo.email",
			},
			Endpoint: google.Endpoint,
		},
	}
}

// 1. Endpoint untuk mendapatkan URL Login Google
func (h *AuthHandler) GetLoginURL(c fiber.Ctx) error {
	// State digunakan untuk mencegah serangan CSRF
	state := "random-string-state-aman"
	url := h.oauthConfig.AuthCodeURL(state)

	return c.JSON(fiber.Map{"url": url})
}

// 2. Endpoint untuk menukar Code dari React menjadi Data Profile Google asli & Menyimpannya ke MySQL
func (h *AuthHandler) HandleCallback(c fiber.Ctx) error {
	type RequestBody struct {
		Code string `json:"code"`
	}

	var body RequestBody
	if err := c.Bind().Body(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Tukar code dari frontend dengan token resmi dari Google
	token, err := h.oauthConfig.Exchange(context.Background(), body.Code)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menukar code ke Google"})
	}

	// Gunakan token tersebut untuk mengambil data diri user dari API Google
	client := h.oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil info user dari Google"})
	}
	defer resp.Body.Close()

	// Parse data JSON response dari Google
	var googleUser map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membaca profile dari Google"})
	}

	// Ekstrak data dari objek Google secara aman
	googleID, _ := googleUser["sub"].(string)
	name, _ := googleUser["name"].(string)
	email, _ := googleUser["email"].(string)
	picture, _ := googleUser["picture"].(string)

	// Petakan data ke dalam struct Model User GORM
	user := model.User{
		GoogleID: googleID,
		Name:     name,
		Email:    email,
		Picture:  picture,
	}

	// 🚀 LOGIKA SAVE/UPSERT VERSI GORM
	var existingUser model.User
	if err := h.db.Where("google_id = ?", googleID).First(&existingUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Buat baru
			if err := h.db.Create(&user).Error; err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Gagal menyelaraskan data user ke database"})
			}
		} else {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal mengecek database"})
		}
	} else {
		// Update data jika sudah ada
		existingUser.Name = name
		existingUser.Email = email
		existingUser.Picture = picture
		if err := h.db.Save(&existingUser).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal memperbarui data user"})
		}
		user = existingUser // Update referensi untuk respon API
	}

	// 🔥 GENERATE TOKEN JWT BARU DI SINI
	tokenString, err := util.GenerateJWT(user.GoogleID, user.Name, user.Email) // 💡 Import "mafriend-tv/internal/util"
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat token otentikasi"})
	}

	// Kembalikan response sukses beserta TOKEN JWT ke React
	return c.JSON(fiber.Map{
		"message": "Login & Sinkronisasi Database Sukses!",
		"token":   tokenString, // 💡 Kirim token ini ke React
		"user":    user,
	})
}