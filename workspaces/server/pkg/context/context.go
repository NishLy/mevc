package pkg

import (
	"context"
	"errors"
)

func GetSubFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return "", errors.New("USER_ID_NOT_FOUND_IN_CONTEXT")
	}
	return userID, nil
}
