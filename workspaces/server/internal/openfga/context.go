package fga

import (
	"context"
	"fmt"

	"github.com/openfga/go-sdk/client"
)

func GetFGAFromContext(ctx context.Context) (*client.OpenFgaClient, error) {
	identifier, ok := ctx.Value("fga").(*client.OpenFgaClient)
	if !ok {
		return nil, fmt.Errorf("fga not found in context")
	}

	return identifier, nil
}
