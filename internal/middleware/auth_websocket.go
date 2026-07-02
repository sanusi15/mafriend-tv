package middleware

import (
	"mafriend-tv/internal/util" // 💡 Import util JWT lu
	"net/http"
)

func WSMiddlewareAuth(next http.HandlerFunc) http.HandlerFunc { // 💡 Parameter *gorm.DB bisa dihapus karena JWT tidak butuh query DB tiap detik
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Ambil token JWT dari query parameter URL (?token=ey...xxx)
		tokenString := r.URL.Query().Get("token")
		if tokenString == "" {
			http.Error(w, "Unauthorized: Token tidak ditemukan.", http.StatusUnauthorized)
			return
		}

		// 2. Validasi Token JWT menggunakan helper util kita
		claims, err := util.ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized: Token ilegal atau kadaluwarsa: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// 3. Masukkan data GoogleID dari JWT ke dalam header request agar bisa dibaca oleh MatchHandler
		r.Header.Set("X-User-GoogleID", claims.GoogleID)

		// 4. Jika aman, loloskan ke handler WebSocket
		next.ServeHTTP(w, r)
	}
}