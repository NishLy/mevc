package fga

import (
	"context"
	"time"

	"github.com/NishLy/go-fiber-boilerplate/internal/platform/cache"
	"github.com/NishLy/go-fiber-boilerplate/pkg/logger"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
)

func GetFGAWriteOptions(fgaClient *client.OpenFgaClient, indentifier string) *client.ClientWriteOptions {
	modelID, _ := cache.Get[string](GetFGAClientModelKey(indentifier), 0, nil)

	if modelID == "" {
		latestModel, err := fgaClient.ReadLatestAuthorizationModel(context.Background()).Execute()
		if err == nil {
			modelID = latestModel.AuthorizationModel.GetId()
			cache.Set(GetFGAClientModelKey(indentifier), modelID, time.Hour)
		}
	}

	var options = client.ClientWriteOptions{
		AuthorizationModelId: openfga.PtrString(modelID),
	}

	return &options
}

func GetFGAClientCheckOptions(fgaClient *client.OpenFgaClient, indentifier string) *client.ClientCheckOptions {
	modelID, _ := cache.Get[string](GetFGAClientModelKey(indentifier), 0, nil)

	if modelID == "" {
		latestModel, err := fgaClient.ReadLatestAuthorizationModel(context.Background()).Execute()
		if err == nil {
			modelID = latestModel.AuthorizationModel.GetId()
			cache.Set(GetFGAClientModelKey(indentifier), modelID, time.Hour)
		}
	}

	var options = client.ClientCheckOptions{
		AuthorizationModelId: openfga.PtrString(modelID),
	}

	return &options
}

func WriteRelationships(clientIns *client.OpenFgaClient,
	options *client.ClientWriteOptions,
	relationships []client.ClientTupleKey) error {

	body := client.ClientWriteRequest{
		Writes: relationships,
	}

	_, err := clientIns.Write(context.Background()).Body(body).Options(*options).Execute()

	if err != nil {
		logger.Sugar.Errorf("Failed to write relationships: %v", err)
		return err
	}

	return err
}

func CheckRelationship(clientIns *client.OpenFgaClient, options *client.ClientCheckOptions, body client.ClientCheckRequest) (bool, error) {
	resp, err := clientIns.Check(context.Background()).Body(body).Options(*options).Execute()
	if err != nil {
		logger.Sugar.Errorf("Failed to check relationship: %v", err)
		return false, err
	}

	return *resp.Allowed, nil
}
