package jwt

import "time"

type JwtTokenMaker interface {
	// CreateToken membuat jwt token baru berdurasi utk user
	CreateToken(username string, duration time.Duration) (string, *Payload, error)

	// VerifyToken cek jika token valid atau tidak
	VerifyToken(token string) (*Payload, error)
}
