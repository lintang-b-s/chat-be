package redisRepo

import (
	"context"
	"errors"
	"fmt"
	"github.com/lintangbs/chat-be/pkg/redispkg"
	"math/rand"
	"time"
)

var (
	OTPNotMatchError = errors.New("User OTP not match with your otp")
)

type OtpRepo struct {
	rds *redispkg.Redis
}

func NewOtp(rds *redispkg.Redis) *OtpRepo {
	return &OtpRepo{rds}
}

// GetOtp mendapatkan otp dari hash set yang ada di redis lalu menghapus otp tersebut dari hash set redis
func (r *OtpRepo) GetOtp(otp string, ctx context.Context, username string) error {
	res := r.rds.Client.HGet(ctx, "otp", otp)
	err := res.Err()
	if err != nil {
		return fmt.Errorf("OtpRepo - GetOTP - r.rds.Client.HGet(ctx, \"otp\", otp): %w", err)
	}

	if res.Val() != username {
		return fmt.Errorf("OtpRepo - GetOTP - res.Val() != username: %w", OTPNotMatchError)
	}

	err = r.rds.Client.HDel(ctx, "otp", otp).Err()
	if err != nil {
		return fmt.Errorf("OtpRepo - GetOTP - r.rds.Client.HDel(ctx, \"otp\", otp).Err(): %w", err)
	}

	return nil
}

var charset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// CreatOtp menyimpan otp string ke hash set di dalam redis
func (r *OtpRepo) CreateOtp(ctx context.Context, username string) (string, error) {
	rand.Seed(time.Now().UnixNano())
	otp := randStr(6)

	err := r.rds.Client.HSet(ctx, "otp", otp, username).Err()
	if err != nil {
		return "", fmt.Errorf("OtpRepo - CreateOtp - r.rds.Client.HSet(ctx, \"otp\", otp, username).Err(): %w", err)
	}
	return otp, nil
}

// randStr generate random string
// n is the length of random string we want to generate
func randStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		// randomly select 1 character from given charset
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
