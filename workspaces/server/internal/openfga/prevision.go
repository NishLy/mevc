package fga

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"encoding/json"

	"github.com/NishLy/go-fiber-boilerplate/config"
	"github.com/openfga/go-sdk/client"
	"github.com/openfga/language/pkg/go/transformer"
)

type FGAIdentifier struct {
	StoreID string
	ModelID string
}

func CreateStore(fgaClient *client.OpenFgaClient, identifier string) (*client.OpenFgaClient, *FGAIdentifier, error) {
	cfg := config.Get()

	ctx := context.Background()
	store, err := fgaClient.CreateStore(ctx).
		Body(client.ClientCreateStoreRequest{
			Name: identifier,
		}).Execute()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create fga store: %w", err)
	}

	storeID := store.GetId()
	if err := fgaClient.SetStoreId(storeID); err != nil {
		return nil, nil, fmt.Errorf("failed to set store ID: %w", err)
	}

	modelID, err := MigrateFromFolder(fgaClient, cfg.OPEN_FGA_MODEL_DIR)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to migrate models: %w", err)
	}

	return fgaClient, &FGAIdentifier{
		StoreID: storeID,
		ModelID: modelID,
	}, nil
}

func MigrateFromFolder(fgaClient *client.OpenFgaClient, folderPath string) (string, error) {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return "", fmt.Errorf("failed to read openfga folder %q: %w", folderPath, err)
	}

	var dslParts []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".fga") {
			continue
		}

		filePath := filepath.Join(folderPath, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file %q: %w", filePath, err)
		}
		dslParts = append(dslParts, string(data))
	}

	if len(dslParts) == 0 {
		return "", fmt.Errorf("no .fga files found in %q", folderPath)
	}

	// 2. Join all DSL parts and transform to an authorization model
	combinedDSL := strings.Join(dslParts, "\n\n")

	// Use the JSON transformer instead of proto
	modelJSON, err := transformer.TransformDSLToJSON(combinedDSL)
	if err != nil {
		return "", fmt.Errorf("failed to parse DSL: %w", err)
	}

	var body client.ClientWriteAuthorizationModelRequest
	if err := json.Unmarshal([]byte(modelJSON), &body); err != nil {
		return "", fmt.Errorf("failed to unmarshal model JSON: %w", err)
	}

	ctx := context.Background()
	resp, err := fgaClient.WriteAuthorizationModel(ctx).Body(body).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to write authorization model: %w", err)
	}

	return resp.GetAuthorizationModelId(), nil
}
