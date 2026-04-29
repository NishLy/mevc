package fga

import (
	"fmt"
	"time"

	"github.com/NishLy/go-fiber-boilerplate/config"
	"github.com/NishLy/go-fiber-boilerplate/internal/platform/cache"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	"github.com/oklog/ulid"
	"github.com/openfga/go-sdk/client"
)

var fgaClients = make(map[string]*client.OpenFgaClient)

func InitOpenFGA(indentifier *string) (*client.OpenFgaClient, error) {
	cfg := config.Get()
	var fgaClient *client.OpenFgaClient
	var err error

	if indentifier != nil {
		fgaClient, err = client.NewSdkClient(&client.ClientConfiguration{
			ApiUrl:  cfg.OPEN_FGA_API_URL, // OpenFGA server address
			StoreId: *indentifier,         // Created via CLI or API
		})
	} else {
		fgaClient, err = client.NewSdkClient(&client.ClientConfiguration{
			ApiUrl: cfg.OPEN_FGA_API_URL, // OpenFGA server address
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenFGA client: %w", err)
	}

	return fgaClient, nil
}

func isValidULID(s string) bool {
	_, err := ulid.Parse(s)
	return err == nil
}

func GetFGAClient(identifier string) (*client.OpenFgaClient, error) {
	key := GetFGAClientStoreKey(identifier)

	cachedKey, _ := cache.Get[string](key, 0, nil)

	if cachedKey != "" {
		identifier = cachedKey
	}

	if client, exists := fgaClients[identifier]; exists {
		return client, nil
	}

	logger.Sugar.Infof("Initialized new OpenFGA client for identifier %s", identifier)

	// If the identifier is a valid ULID, use it directly as the store ID. Otherwise, treat it as a tenant identifier and generate a store ID based on the app name and tenant ID.
	if isValidULID(identifier) {
		fgaClient, err := InitOpenFGA(&identifier)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize OpenFGA client: %w", err)
		}
		fgaClients[identifier] = fgaClient
		cache.Set(key, identifier, time.Hour)
		return fgaClient, nil
	} else {
		logger.Sugar.Infof("Identifier %s is not a valid ULID, treating it as tenant identifier", identifier)
		fgaClient, err := InitOpenFGA(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize OpenFGA client: %w", err)
		}

		fgaClient, ids, err := CreateStore(fgaClient, identifier)
		if err != nil {
			return nil, fmt.Errorf("failed to provision new store: %w", err)
		}

		fgaClients[ids.StoreID] = fgaClient

		cache.Set(key, ids.StoreID, time.Hour)
		cache.Set(GetFGAClientModelKey(identifier), ids.ModelID, time.Hour)
		return fgaClient, nil
	}
}
