package main

import (
	"database/sql"
	"io"
	"fmt"
	"mime/multipart"
	"strings"
	"strconv"
	"path/filepath"
	"time"
	"os"
	"net/url"
	"net/http"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

var db *sql.DB

func statusParam(status string) string {
	switch status {
	case "sudah dibayar":
		return "success"
	case "menunggu":
		return "warning"
	case "gagal":
		return "error"
	default:
		return "success"
	}
}


func formatRupiah(num int) string {
	str := fmt.Sprintf("%d", num)
	n := len(str)
	if n <= 3 {
		return "Rp " + str
	}

	var b strings.Builder
	mod := n % 3
	if mod > 0 {
		b.WriteString(str[:mod])
		if n > mod {
			b.WriteString(".")
		}
	}
	for i := mod; i < n; i += 3 {
		b.WriteString(str[i : i+3])
		if i+3 < n {
			b.WriteString(".")
		}
	}
	return "Rp " + b.String()
}

func main() {
	
	var err error
	db, err = sql.Open("mysql", "root:@tcp(localhost)/vibe_apparel")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	e := echo.New()

	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}

	e.Renderer = renderer
	e.Static("/static", "static")

	// Login
	e.GET("/login", func(c echo.Context) error {
		return c.Render(http.StatusOK, "login.html", map[string]interface{}{
			"Title": "Login - Vibe Apparel",
		})
	})

	e.POST("/auth", func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")

		var hashedPassword string
		var userID int
		err := db.QueryRow("SELECT id, password FROM users WHERE username = ?", username).Scan(&userID, &hashedPassword)

		if err != nil {
			return c.String(http.StatusUnauthorized, "Username atau password salah")
		}

		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
		if err != nil {
			return c.String(http.StatusUnauthorized, "Username atau password salah")
		}

		
		userIDCookie := new(http.Cookie)
		userIDCookie.Name = "user_id"
		userIDCookie.Value = strconv.Itoa(userID)
		userIDCookie.Path = "/"
		userIDCookie.HttpOnly = true
		c.SetCookie(userIDCookie)

		usernameCookie := new(http.Cookie)
    	usernameCookie.Name = "username"
    	usernameCookie.Value = username
    	usernameCookie.Path = "/"
    	c.SetCookie(usernameCookie)


		return c.Redirect(http.StatusSeeOther, "/home")
	})

	e.GET("/logout", func(c echo.Context) error {
		// Hapus cookie username
		usernameCookie := new(http.Cookie)
		usernameCookie.Name = "username"
		usernameCookie.Value = ""
		usernameCookie.Path = "/"
		usernameCookie.MaxAge = -1 
		c.SetCookie(usernameCookie)
		
		// Hapus juga cookie user_id
		userIDCookie := new(http.Cookie)
		userIDCookie.Name = "user_id"
		userIDCookie.Value = ""
		userIDCookie.Path = "/"
		userIDCookie.MaxAge = -1
		c.SetCookie(userIDCookie)
	
		return c.Redirect(http.StatusSeeOther, "/login")
	})
	

	// Register
	e.GET("/register", func(c echo.Context) error {
		return c.Render(http.StatusOK, "register.html", map[string]interface{}{
			"Title": "Register - Vibe Apparel",
		})
	})

	e.POST("/register", func(c echo.Context) error {
		email := c.FormValue("email")
		username := c.FormValue("username")
		password := c.FormValue("password")

		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		_, err := db.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, ?)", email, username, string(hash))
		if err != nil {
			return c.String(http.StatusInternalServerError, "Gagal registrasi: "+err.Error())
		}

		return c.Redirect(http.StatusSeeOther, "/login")

	})

	// Lupa Password
	e.GET("/lupa_password", func(c echo.Context) error {
		return c.Render(http.StatusOK, "lupa_password.html", map[string]interface{}{
			"Title": "Lupa Password",
		})
	})

	e.POST("/forgot-password", func(c echo.Context) error {
		email := c.FormValue("email")
		newPassword := c.FormValue("password_baru")

		hashedNewPassword, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)

		res, err := db.Exec("UPDATE users SET password = ? WHERE email = ?", hashedNewPassword, email)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Terjadi kesalahan saat reset password")
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			return c.Render(http.StatusOK, "lupa_password.html", map[string]interface{}{
				"Title": "Lupa Password",
				"ErrorMessage": "Email tidak ditemukan",
			})			
		}

		return c.Render(http.StatusOK, "lupa_password.html", map[string]interface{}{
			"Title": "Lupa Password",
			"SuccessMessage": "Password berhasil direset!",
		})
		
	})

	e.GET("/home", func(c echo.Context) error {

		userIDCookie, err := c.Cookie("user_id")
    	if err != nil || userIDCookie.Value == "" {
        	return c.Redirect(http.StatusSeeOther, "/login")
    	}
    

    	usernameCookie, err := c.Cookie("username")
    	username := "Guest"
    	if err == nil && usernameCookie.Value != "" {
        	username = usernameCookie.Value
    	}
	
		// Query database untuk mendapatkan data jersey
		rows, err := db.Query("SELECT id, deskripsi, image_url, harga FROM jerseys")
		if err != nil {
			return c.String(http.StatusInternalServerError, "Gagal mengambil data jersey")
		}
		defer rows.Close()
	
		// Struct untuk menyimpan data jersey
		type Jersey struct {
			ID        int
			Deskripsi string
			Image     string
			Harga     float64
		}
		var jerseys []Jersey
		for rows.Next() {
			var j Jersey
			err := rows.Scan(&j.ID, &j.Deskripsi, &j.Image, &j.Harga)
			if err != nil {
				return c.String(http.StatusInternalServerError, "Gagal membaca data jersey")
			}
			jerseys = append(jerseys, j)
		}
	
		// Data untuk dikirim ke template
		data := map[string]interface{}{
			"Username": username,
			"Jerseys":  jerseys,
		}
		return c.Render(http.StatusOK, "home.html", data)
	})

	e.GET("/status", func(c echo.Context) error {
		// Ambil cookie user_id
		userIDCookie, err := c.Cookie("user_id")
		if err != nil || userIDCookie.Value == "" {
			return c.Redirect(http.StatusSeeOther, "/login")
		}
	
		// Parsing user_id ke int
		userID, err := strconv.Atoi(userIDCookie.Value)
		if err != nil {
			return c.String(http.StatusBadRequest, "User ID tidak valid")
		}
	
		// Struct Jersey dan StatusPesanan
		type Jersey struct {
			Deskripsi string
			ImageURL  string
		}
		type StatusPesanan struct {
			Jersey             Jersey
			NamaKlub           string
			JumlahPemain       int
			StatusPembayaran   string
			CreatedAtFormatted string
			PaidAtFormatted    string
		}
	
		// Format tanggal ke format Indonesia
		formatTanggalIndo := func(t time.Time) string {
			bulan := [...]string{
				"Januari", "Februari", "Maret", "April", "Mei", "Juni",
				"Juli", "Agustus", "September", "Oktober", "November", "Desember",
			}
			return fmt.Sprintf("%02d %s %d", t.Day(), bulan[int(t.Month())-1], t.Year())
		}
	
		// Query database, ambil created_at dari tabel payments (p.created_at)
		rows, err := db.Query(`
			SELECT j.deskripsi, j.image_url, cj.nama_klub, cj.jumlah_pemain, 
		   		COALESCE(p.status, 'menunggu'), p.created_at, p.updated_at
			FROM custom_jersey cj
			JOIN jerseys j ON cj.jersey_id = j.id
			LEFT JOIN payments p ON cj.id = p.custom_id
			WHERE cj.user_id = ?`, userID)

		if err != nil {
			return c.String(http.StatusInternalServerError, "Gagal mengambil data status pesanan")
		}
		defer rows.Close()
	
		var daftar []StatusPesanan
		for rows.Next() {
			var item StatusPesanan
			var createdAtStr sql.NullString
			var updatedAtStr sql.NullString
			err := rows.Scan(
				&item.Jersey.Deskripsi,
				&item.Jersey.ImageURL,
				&item.NamaKlub,
				&item.JumlahPemain,
				&item.StatusPembayaran,
				&createdAtStr,
				&updatedAtStr,
			)
			if err != nil {
				return c.String(http.StatusInternalServerError, fmt.Sprintf("Gagal membaca data status pesanan: %v", err))
			}
			
			if item.StatusPembayaran == "sudah dibayar" && updatedAtStr.Valid {
				updatedAtTime, err := time.Parse("2006-01-02 15:04:05", updatedAtStr.String)
				if err == nil {
					item.PaidAtFormatted = formatTanggalIndo(updatedAtTime)
				} else {
					item.PaidAtFormatted = "Tanggal tidak valid"
				}
			} else {
				item.PaidAtFormatted = "-"
			}
			
	
			// Parse tanggal hanya jika ada nilainya
			if createdAtStr.Valid {
				createdAtTime, err := time.Parse("2006-01-02 15:04:05", createdAtStr.String)
				if err == nil {
					item.CreatedAtFormatted = formatTanggalIndo(createdAtTime)
				} else {
					item.CreatedAtFormatted = "Tanggal tidak valid"
				}
			} else {
				item.CreatedAtFormatted = "-"
			}
	
			daftar = append(daftar, item)
		}
	
		// Kirim ke template
		data := map[string]interface{}{
			"Pesanan": daftar,
		}
		return c.Render(http.StatusOK, "status_pemesanan.html", data)
	})
	
	
	e.GET("/detail/:id", func(c echo.Context) error {
		id := c.Param("id")
		var jersey struct {
			ID        int
			Deskripsi string
			ImageURL  string
			Harga     float64
		}
	
		err := db.QueryRow("SELECT id, deskripsi, image_url, harga FROM jerseys WHERE id = ?", id).
			Scan(&jersey.ID, &jersey.Deskripsi, &jersey.ImageURL, &jersey.Harga)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Jersey tidak ditemukan")
		}
	
		data := map[string]interface{}{
			"Title":   "Detail Pemesanan - Vibe Apparel",
			"Jersey":  jersey,
		}
		return c.Render(http.StatusOK, "detail.html", data)
	})
	
	e.GET("/custom/:id", func(c echo.Context) error {
		id := c.Param("id")
		data := map[string]interface{}{
			"JerseyID": id,
		}
		return c.Render(http.StatusOK, "custom.html", data)
	})
	
	e.POST("/submit-custom", func(c echo.Context) error {
		// Ambil user_id dari cookie
		userCookie, err := c.Cookie("user_id")
		if err != nil {
			return c.String(http.StatusBadRequest, "User belum login")
		}
		userID, _ := strconv.Atoi(userCookie.Value)
	
		jerseyID, _ := strconv.Atoi(c.FormValue("jersey_id"))
		adaNamaKlub := c.FormValue("ada_nama_klub") == "yes"
		namaKlub := c.FormValue("nama_klub")
		jumlahPemain, _ := strconv.Atoi(c.FormValue("jumlah_pemain"))
	
		// Upload file logo
		logoTengah, _ := c.FormFile("logo_tengah")
		logoDadaKanan, _ := c.FormFile("logo_dada_kanan")
		logoDadaKiri, _ := c.FormFile("logo_dada_kiri")
	
		saveFile := func(file *multipart.FileHeader, fieldName string) (string, error) {
			if file == nil {
				return "", nil
			}
			filename := fmt.Sprintf("%d_%s_%s", time.Now().UnixNano(), fieldName, file.Filename)
			dst := filepath.Join("static/uploads", filename)
			src, _ := file.Open()
			defer src.Close()
			out, _ := os.Create(dst)
			defer out.Close()
			io.Copy(out, src)
			return "/static/uploads/" + filename, nil
		}
	
		logoTengahPath, _ := saveFile(logoTengah, "tengah")
		logoKananPath, _ := saveFile(logoDadaKanan, "kanan")
		logoKiriPath, _ := saveFile(logoDadaKiri, "kiri")
	
		// Simpan ke custom_jersey
		result, err := db.Exec(`
			INSERT INTO custom_jersey 
			(jersey_id, user_id, nama_klub, ada_nama_klub, jumlah_pemain, logo_tengah, logo_dada_kanan, logo_dada_kiri)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			jerseyID, userID, namaKlub, adaNamaKlub, jumlahPemain, logoTengahPath, logoKananPath, logoKiriPath,
		)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Gagal menyimpan data custom jersey")
		}
		customID, _ := result.LastInsertId()
	
		// Simpan data pemain
		for i := 1; i <= jumlahPemain; i++ {
			name := c.FormValue(fmt.Sprintf("nama_pemain_%d", i))
			number := c.FormValue(fmt.Sprintf("nomor_punggung_%d", i))
			size := c.FormValue(fmt.Sprintf("ukuran_%d", i))
		
			_, err := db.Exec(`
				INSERT INTO custom_pemain (custom_id, nama_pemain, nomor_punggung, ukuran)
				VALUES (?, ?, ?, ?)`,
				customID, name, number, size,
			)
			if err != nil {
				return c.String(http.StatusInternalServerError, fmt.Sprintf("Gagal menyimpan pemain %d", i))
			}
		}		
	
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/checkout/%d", customID))
	})
	
	e.GET("/checkout/:id", func(c echo.Context) error {
		customID, _ := strconv.Atoi(c.Param("id"))
	
		// Ambil data dari tabel custom_jersey
		var jersey struct {
			NamaKlub    string
			AdaNamaKlub bool
		}
		err := db.QueryRow(`
			SELECT nama_klub, ada_nama_klub 
			FROM custom_jersey 
			WHERE id = ?
		`, customID).Scan(&jersey.NamaKlub, &jersey.AdaNamaKlub)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Gagal mengambil data custom")
		}
	
		// Ambil daftar pemain dari tabel custom_pemain 
		rows, err := db.Query(`
			SELECT nama_pemain, nomor_punggung, ukuran 
			FROM custom_pemain 
			WHERE custom_id = ?
		`, customID)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Gagal mengambil data pemain")
		}
		defer rows.Close()
	
		var players []map[string]string
		for rows.Next() {
			var namaPemain, nomorPunggung, ukuran string
			if err := rows.Scan(&namaPemain, &nomorPunggung, &ukuran); err != nil {
				return c.String(http.StatusInternalServerError, "Gagal membaca data pemain")
			}
			players = append(players, map[string]string{
				"nama_pemain":    namaPemain,
				"nomor_punggung": nomorPunggung,
				"ukuran":         ukuran,
			})
		}
	
		// Kirim data ke checkout.html
		data := map[string]interface{}{
			"CustomID": customID,
			"Jersey":   jersey,
			"Players":  players,
		}
		return c.Render(http.StatusOK, "checkout.html", data)
	})
	
	
	e.POST("/submit-checkout", func(c echo.Context) error {
		customID, _ := strconv.Atoi(c.FormValue("custom_id"))
		namaPenerima := c.FormValue("nama_penerima")
		alamat := c.FormValue("alamat")
		noTelepon := c.FormValue("no_telepon")
	
		_, err := db.Exec(`
			INSERT INTO checkout (custom_id, nama_penerima, alamat, no_telepon)
			VALUES (?, ?, ?, ?)`,
			customID, namaPenerima, alamat, noTelepon,
		)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Gagal menyimpan data checkout")
		}
	
		return c.Redirect(http.StatusSeeOther, "/payment/"+strconv.Itoa(customID))
	})

	e.GET("/payment/:id", func(c echo.Context) error {
		customID, _ := strconv.Atoi(c.Param("id"))
	
		// 1. Ambil jersey_id dari custom_jersey
		var jerseyID int
		err := db.QueryRow(`
			SELECT jersey_id
			FROM custom_jersey
			WHERE id = ?
		`, customID).Scan(&jerseyID)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Gagal mengambil jersey_id")
		}
	
		// 2. Ambil harga dari tabel jerseys
		var hargaPerPemain int
		err = db.QueryRow(`
			SELECT harga
			FROM jerseys
			WHERE id = ?
		`, jerseyID).Scan(&hargaPerPemain)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Gagal mengambil harga jersey")
		}
	
		// 3. Hitung jumlah pemain
		var jumlahPemain int
		err = db.QueryRow(`
			SELECT COUNT(*)
			FROM custom_pemain
			WHERE custom_id = ?
		`, customID).Scan(&jumlahPemain)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Gagal menghitung jumlah pemain")
		}
	
		// 4. Hitung total harga
		const ongkir = 20000
		totalHarga := (hargaPerPemain * jumlahPemain) + ongkir

	
		data := map[string]interface{}{
			"CustomID":    customID,
			"TotalHarga":  totalHarga,
			"Ongkir":      ongkir,
			"NoRekening":  "8730802731",
			"Bank":        "BCA",
			"AtasNama":    "Viodika Arya R",
		}		
		return c.Render(http.StatusOK, "payment.html", data)
	})
	
	e.POST("/submit-payment", func(c echo.Context) error {
		// 1. Ambil custom_id
		customID := c.FormValue("custom_id")
	
		// 2. Ambil file bukti transfer
		file, err := c.FormFile("bukti_transfer")
		if err != nil {
			return c.String(http.StatusBadRequest, "Gagal upload file")
		}
	
		// 3. Buka file
		src, err := file.Open()
		if err != nil {
			return c.String(http.StatusInternalServerError, "Gagal membuka file")
		}
		defer src.Close()
	
		// 4. Simpan file ke folder static/bukti/
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename) // contoh: 1714200000_buktitrf.jpg
		dstPath := "static/bukti/" + filename
	
		dst, err := os.Create(dstPath)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Gagal menyimpan file")
		}
		defer dst.Close()
	
		if _, err = io.Copy(dst, src); err != nil {
			return c.String(http.StatusInternalServerError, "Gagal menyalin file")
		}
	
		// 5. Simpan informasi ke database
		_, err = db.Exec(`
			INSERT INTO payments (custom_id, bukti_transfer)
			VALUES (?, ?)
		`, customID, filename)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Gagal menyimpan ke database")
		}
	
		return c.Redirect(http.StatusSeeOther, "/home")
	})
	
	e.GET("/login-admin", func(c echo.Context) error {
		return c.Render(http.StatusOK, "login_admin.html", map[string]interface{}{
			"Title": "Login Admin - Vibe Apparel",
		})
	})
	
	e.POST("/auth-admin", func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")
	
		var dbPassword string
		var adminID int
		err := db.QueryRow("SELECT id, password FROM admin WHERE username = ?", username).Scan(&adminID, &dbPassword)
		if err != nil {
			return c.String(http.StatusUnauthorized, "Username atau password salah")
		}
	
	
		// Bandingkan password langsung (TANPA bcrypt)
		if password != dbPassword {
			return c.String(http.StatusUnauthorized, "Username atau password salah")
		}
	
		// Password cocok → lanjut login
		adminIDCookie := new(http.Cookie)
		adminIDCookie.Name = "admin_id"
		adminIDCookie.Value = strconv.Itoa(adminID)
		adminIDCookie.Path = "/"
		adminIDCookie.HttpOnly = true
		c.SetCookie(adminIDCookie)
	
		return c.Redirect(http.StatusSeeOther, "/dashboard")
	})

	e.GET("/logout-admin", func(c echo.Context) error {
		// Hapus cookie admin_id
		expiredCookie := new(http.Cookie)
		expiredCookie.Name = "admin_id"
		expiredCookie.Value = ""
		expiredCookie.Path = "/"
		expiredCookie.Expires = time.Unix(0, 0) // langsung expired
		expiredCookie.MaxAge = -1
		c.SetCookie(expiredCookie)
	
		// Redirect ke halaman login admin
		return c.Redirect(http.StatusSeeOther, "/login-admin")
	})
	
	e.GET("/dashboard", func(c echo.Context) error {
		// Cek cookie admin_id untuk autentikasi
		_, err := c.Cookie("admin_id")
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/login-admin")
		}
	
		// Hitung total pendapatan
		var totalPendapatan int
		err = db.QueryRow(`
			SELECT COALESCE(SUM((SELECT harga FROM jerseys WHERE id = cj.jersey_id) * 
			(SELECT COUNT(*) FROM custom_pemain WHERE custom_id = cj.id)) + 
			(COUNT(*) * 20000), 0)
			FROM custom_jersey cj
			JOIN payments p ON p.custom_id = cj.id
			WHERE p.status = 'sudah dibayar' OR p.status IS NULL
		`).Scan(&totalPendapatan)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Gagal mengambil data pendapatan: "+err.Error())
		}
	
		// Hitung jumlah produk
		var jumlahProduk int
		err = db.QueryRow("SELECT COUNT(*) FROM jerseys").Scan(&jumlahProduk)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Gagal mengambil jumlah produk")
		}
	
		// Hitung jumlah pesanan
		var jumlahPesanan int
		err = db.QueryRow("SELECT COUNT(*) FROM custom_jersey").Scan(&jumlahPesanan)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Gagal mengambil jumlah pesanan")
		}
	
		return c.Render(http.StatusOK, "dashboard.html", map[string]interface{}{
			"Title": "Dashboard Admin - Vibe Apparel",
			"TotalPendapatan": formatRupiah(totalPendapatan),
			"JumlahProduk": jumlahProduk,
			"JumlahPesanan": jumlahPesanan,
		})
	})

	// Halaman daftar produk admin
e.GET("/produk", func(c echo.Context) error {
    _, err := c.Cookie("admin_id")
    if err != nil {
        return c.Redirect(http.StatusSeeOther, "/login-admin")
    }

    rows, err := db.Query("SELECT id, deskripsi, image_url, harga FROM jerseys")
    if err != nil {
        return c.String(http.StatusInternalServerError, "Gagal mengambil data jersey")
    }
    defer rows.Close()

    type Jersey struct {
        ID        int
        Deskripsi string
        ImageURL  string
        Harga     float64
    }
    var jerseys []Jersey
    for rows.Next() {
        var j Jersey
        err := rows.Scan(&j.ID, &j.Deskripsi, &j.ImageURL, &j.Harga)
        if err != nil {
            return c.String(http.StatusInternalServerError, "Gagal membaca data jersey")
        }
        jerseys = append(jerseys, j)
    }

    return c.Render(http.StatusOK, "produk.html", map[string]interface{}{
        "Title": "Kelola Produk - Vibe Apparel",
        "Produk": jerseys,
		"Success": c.QueryParam("success"),
    })
})

// GET form tambah produk
e.GET("/TambahProduk", func(c echo.Context) error {
	return c.Render(http.StatusOK, "TambahProduk.html", nil)
})

// POST simpan produk baru
e.POST("/tambah-produk", func(c echo.Context) error {
	deskripsi := c.FormValue("deskripsi")
	hargaStr := c.FormValue("harga")
	
	// Mendapatkan file gambar
    gambarFile, err := c.FormFile("gambar")
    if err != nil {
        return c.String(http.StatusBadRequest, "Gagal mengupload gambar: "+err.Error())
    }
    
    // Validasi & parsing harga
    harga, err := strconv.ParseFloat(hargaStr, 64)
    if err != nil {
        return c.String(http.StatusBadRequest, "Harga tidak valid")
    }
    
    // Dapatkan ID terakhir dari database untuk menentukan nomor jersey berikutnya
    var lastID int
    err = db.QueryRow("SELECT MAX(id) FROM jerseys").Scan(&lastID)
    if err != nil && err != sql.ErrNoRows {
        return err
    }
    
    // Nomor jersey baru = lastID + 1
    newJerseyNumber := lastID + 1
    
    // Tentukan ekstensi file dari file yang diupload
    fileExt := filepath.Ext(gambarFile.Filename) // Contoh: .jpg, .jpeg, .png
    if fileExt == "" {
        fileExt = ".jpeg" // Default jika tidak ada ekstensi
    }
    
    // Buat nama file baru dengan format jersey + nomor
    newFileName := fmt.Sprintf("jersey%d%s", newJerseyNumber, fileExt)
    
    // Simpan file gambar
    src, err := gambarFile.Open()
    if err != nil {
        return err
    }
    defer src.Close()
    
    // Buat lokasi penyimpanan gambar dengan nama file baru
    savePath := filepath.Join("static/images", newFileName)
    
    // Buat file tujuan
    dstFile, err := os.Create(savePath)
    if err != nil {
        return err
    }
    defer dstFile.Close()
    
    // Salin konten dari file upload ke file tujuan
    if _, err = io.Copy(dstFile, src); err != nil {
        return err
    }
    
    // Path yang akan disimpan di database
    imageURLPath := "/static/images/" + newFileName
    
    // Simpan ke database menggunakan path lengkap
    _, err = db.Exec("INSERT INTO jerseys (deskripsi, image_url, harga) VALUES (?, ?, ?)", 
        deskripsi, imageURLPath, harga)
    if err != nil {
        return err
    }

	return c.Redirect(http.StatusSeeOther, "/produk?success=tambah")
})

// HAPUS produk
e.GET("/HapusProduk/:id", func(c echo.Context) error {
    id := c.Param("id")
    
    // Hapus file gambar dulu (opsional, bisa ditambah kalau perlu)
    // Kemudian hapus dari database
    _, err := db.Exec("DELETE FROM jerseys WHERE id = ?", id)
    if err != nil {
        return err
    }

    return c.Redirect(http.StatusSeeOther, "/produk?success=hapus")
})

// GET halaman edit produk
e.GET("/EditProduk/:id", func(c echo.Context) error {
    id := c.Param("id")
	type Produk struct {
		ID        int
		Deskripsi string
		ImageURL  string
		Harga     float64
	}

    var p Produk
    err := db.QueryRow("SELECT id, deskripsi, image_url, harga FROM jerseys WHERE id = ?", id).
        Scan(&p.ID, &p.Deskripsi, &p.ImageURL, &p.Harga)

    if err != nil {
        return err
    }

    return c.Render(http.StatusOK, "EditProduk.html", map[string]interface{}{
        "Produk": p,
    })
})

// POST simpan edit produk
e.POST("/edit-produk/:id", func(c echo.Context) error {
    id := c.Param("id")
    deskripsi := c.FormValue("deskripsi")
    hargaStr := c.FormValue("harga")
    harga, err := strconv.ParseFloat(hargaStr, 64)
    if err != nil {
        return c.String(http.StatusBadRequest, "Harga tidak valid")
    }

    // Cek apakah ada gambar baru
    file, err := c.FormFile("gambar")
    if err == nil {
        // Simpan gambar baru
        src, _ := file.Open()
        defer src.Close()

        dstPath := "static/images/" + file.Filename
        dst, _ := os.Create(dstPath)
        defer dst.Close()
        io.Copy(dst, src)

        // Update termasuk gambar
        _, err = db.Exec("UPDATE jerseys SET deskripsi=?, image_url=?, harga=? WHERE id=?",
            deskripsi, file.Filename, harga, id)
    } else {
        // Update tanpa gambar
        _, err = db.Exec("UPDATE jerseys SET deskripsi=?, harga=? WHERE id=?",
            deskripsi, harga, id)
    }

    if err != nil {
        return err
    }

    return c.Redirect(http.StatusSeeOther, "/produk?success=edit")
})


e.GET("/pesanan", func(c echo.Context) error {
    _, err := c.Cookie("admin_id")
    if err != nil {
        return c.Redirect(http.StatusSeeOther, "/login-admin")
    }
	success := c.QueryParam("success")
	warning := c.QueryParam("warning")
	errorMsg := c.QueryParam("error")
	
    rows, err := db.Query(`
        SELECT 
            ch.id, 
            cj.nama_klub, 
            cj.jumlah_pemain,
            cj.logo_tengah,
            cj.logo_dada_kanan,
            cj.logo_dada_kiri
        FROM checkout ch
        JOIN custom_jersey cj ON ch.custom_id = cj.id
        ORDER BY ch.created_at DESC
    `)
    if err != nil {
        return c.String(http.StatusInternalServerError, "Gagal mengambil data pesanan")
    }
    defer rows.Close()

    type Pesanan struct {
        ID            int
        NamaKlub      string
        JumlahPemain  int
        LogoTengah    string
        LogoKanan     string
        LogoKiri      string
    }

    var pesanan []Pesanan
    for rows.Next() {
        var p Pesanan
        err := rows.Scan(&p.ID, &p.NamaKlub, &p.JumlahPemain, &p.LogoTengah, &p.LogoKanan, &p.LogoKiri)
        if err != nil {
            return c.String(http.StatusInternalServerError, "Gagal membaca data pesanan")
        }
        pesanan = append(pesanan, p)
    }

    return c.Render(http.StatusOK, "pesanan.html", map[string]interface{}{
        "Title": "Kelola Pesanan - Vibe Apparel",
        "Pesanan": pesanan,
		"Success": success,
		"Warning": warning,
		"Error":   errorMsg,
    })
})

e.GET("/pesanan/:id", func(c echo.Context) error {
	id := c.Param("id")

	// Ambil data pesanan dari tabel checkout
	type Pesanan struct {
		ID           int
		NamaPenerima string
		Alamat       string
		NoTelepon    string
		CustomID     int
	}

	var pesanan Pesanan
	err := db.QueryRow(`
		SELECT id, nama_penerima, alamat, no_telepon, custom_id 
		FROM checkout WHERE id = ?`, id).
		Scan(&pesanan.ID, &pesanan.NamaPenerima, &pesanan.Alamat, &pesanan.NoTelepon, &pesanan.CustomID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Pesanan tidak ditemukan")
	}

	// Format tanggal Indonesia
	formatTanggalIndo := func(t time.Time) string {
		bulan := [...]string{
			"Januari", "Februari", "Maret", "April", "Mei", "Juni",
			"Juli", "Agustus", "September", "Oktober", "November", "Desember",
		}
		return fmt.Sprintf("%02d %s %d", t.Day(), bulan[int(t.Month())-1], t.Year())
	}

	// Ambil daftar pemain dari tabel custom_pemain
	type Pemain struct {
		Nama       string
		NoPunggung string
		Ukuran     string
	}

	var pemain []Pemain
	rows, _ := db.Query(`SELECT nama_pemain, nomor_punggung, ukuran FROM custom_pemain WHERE custom_id = ?`, pesanan.CustomID)
	defer rows.Close()
	for rows.Next() {
		var p Pemain
		rows.Scan(&p.Nama, &p.NoPunggung, &p.Ukuran)
		pemain = append(pemain, p)
	}

	// Ambil detail jersey
	type Jersey struct {
		Deskripsi string
		ImageURL  string
		Harga     float64
	}

	var jerseyList []Jersey
	var hargaJersey float64
	jrows, _ := db.Query(`
		SELECT j.deskripsi, j.image_url, j.harga 
		FROM custom_jersey cj
		JOIN jerseys j ON cj.jersey_id = j.id
		WHERE cj.id = ?`, pesanan.CustomID)
	defer jrows.Close()
	for jrows.Next() {
		var j Jersey
		jrows.Scan(&j.Deskripsi, &j.ImageURL, &j.Harga)
		hargaJersey = j.Harga
		jerseyList = append(jerseyList, j)
	}

	// Ambil jumlah pemain dari custom_jersey
	var jumlahPemain int
	db.QueryRow(`SELECT jumlah_pemain FROM custom_jersey WHERE id = ?`, pesanan.CustomID).Scan(&jumlahPemain)

	// Hitung total harga
	totalHarga := (float64(jumlahPemain) * hargaJersey) + 20000

	// Ambil status pembayaran dan created_at
	var statusPembayaran string
	var createdAtStr sql.NullString
	db.QueryRow(`SELECT status, created_at FROM payments WHERE custom_id = ?`, pesanan.CustomID).
		Scan(&statusPembayaran, &createdAtStr)

	// Format tanggal jika ada
	var createdAtFormatted string
	if createdAtStr.Valid {
		createdAtTime, err := time.Parse("2006-01-02 15:04:05", createdAtStr.String)
		if err == nil {
			createdAtFormatted = formatTanggalIndo(createdAtTime)
		} else {
			createdAtFormatted = "Tanggal tidak valid"
		}
	} else {
		createdAtFormatted = "-"
	}

	// Ambil bukti transfer
	var buktiTransfer string
	db.QueryRow(`SELECT bukti_transfer FROM payments WHERE custom_id = ?`, pesanan.CustomID).Scan(&buktiTransfer)

	return c.Render(http.StatusOK, "detail_pesanan.html", map[string]interface{}{
		"Pesanan":           pesanan,
		"Pemain":            pemain,
		"Jersey":            jerseyList,
		"TotalHarga":        totalHarga,
		"StatusPembayaran":  statusPembayaran,
		"BuktiTransfer":     buktiTransfer,
		"CreatedAtFormatted": createdAtFormatted,
	})
})

e.POST("/pesanan/update-status", func(c echo.Context) error {
	id := c.FormValue("id")         // ID dari tabel checkout
	status := c.FormValue("status") // 'menunggu', 'sudah dibayar', atau 'gagal'

	// Ambil custom_id dan nama penerima
	var customID int
	var namaPenerima string
	err := db.QueryRow(`SELECT custom_id, nama_penerima FROM checkout WHERE id = ?`, id).Scan(&customID, &namaPenerima)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Gagal menemukan data pesanan")
	}

	// Update status di tabel payments
	_, err = db.Exec(`UPDATE payments SET status = ? WHERE custom_id = ?`, status, customID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Gagal mengupdate status pembayaran")
	}

	// Redirect ke halaman pesanan dengan notifikasi
	message := fmt.Sprintf("Status pesanan untuk %s berhasil diubah menjadi '%s'", namaPenerima, status)
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/pesanan?%s=%s", statusParam(status), url.QueryEscape(message)))
})
	e.Logger.Fatal(e.Start(":8080"))
}
