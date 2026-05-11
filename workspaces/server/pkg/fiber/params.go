package pkg

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
)

func GetUint64PtrFromParams(c fiber.Ctx, key string) (*uint64, error) {
	val := c.Params(key)
	if val == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseUint(val, 10, 64)
	return &parsed, err
}
