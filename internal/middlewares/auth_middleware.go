package middlewares

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// Role defines the user roles.
const (
	RoleAdmin   = "admin"
	RoleManager = "manager"
	RoleCashier = "cashier"
)

// RBACMiddleware checks if the user's role allows access to the route.
func RBACMiddleware(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Ambil Role dari Context (Disimpan oleh JWTMiddleware)
		userRole, ok := c.Locals("role").(string)
		if !ok || userRole == "" {
			// Ini seharusnya tidak terjadi jika JWTMiddleware sudah berjalan.
			// Tapi untuk keamanan, kita tolak.
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Akses ditolak: Informasi peran (role) tidak ditemukan",
			})
		}

		// 2. Cek apakah Role pengguna ada di daftar role yang diizinkan
		isAllowed := false
		for _, role := range allowedRoles {
			if strings.EqualFold(userRole, role) { // Gunakan EqualFold untuk case-insensitivity
				isAllowed = true
				break
			}
		}

		// 3. Beri Keputusan
		if !isAllowed {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Akses ditolak: Peran (" + strings.ToUpper(userRole) + ") tidak memiliki izin",
			})
		}

		// Lanjutkan ke handler jika diizinkan
		return c.Next()
	}
}

// JWTMiddleware memvalidasi token JWT yang disertakan dalam request.
func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Ambil Header Authorization
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token otentikasi diperlukan",
			})
		}

		// 2. Periksa format "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Format token tidak valid (Harus 'Bearer <token>')",
			})
		}
		tokenString := parts[1]
		secret := os.Getenv("JWT_SECRET")

		// 3. Parsing dan Validasi Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Memastikan algoritma yang digunakan adalah HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(secret), nil
		})

		if err != nil {
			// Memeriksa string error untuk handling expired token
			if strings.Contains(err.Error(), "token is expired") {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Token kedaluwarsa",
				})
			}
			// Error umum lainnya (signature invalid, dll.)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token tidak valid",
			})
		}

		// 4. Periksa Validitas Token
		if !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token tidak valid",
			})
		}

		// 5. Simpan Claims ke Context (Penting untuk mendapatkan user_id/role di Handler)
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token claims tidak valid",
			})
		}

		// Tambahkan data user ke context Fiber
		c.Locals("userID", claims["user_id"])
		c.Locals("role", claims["role"])

		// Lanjutkan ke Handler berikutnya
		return c.Next()
	}
}
