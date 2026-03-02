package jwt_util

import (
	"time"
)

type Config struct {
	Issuer         string
	SecretKey      string
	ExpirationTime time.Duration
}
