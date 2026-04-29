package fga

import "fmt"

func GetFGAClientStoreKey(identifier string) string {
	return fmt.Sprintf("%s_fga_store_id", identifier)
}

func GetFGAClientModelKey(identifier string) string {
	return fmt.Sprintf("%s_fga_model_id", identifier)
}
