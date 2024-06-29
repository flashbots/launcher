package secret

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

var (
	errAzurePropertyIdIsNil  = errors.New("azure: got nil property id for a vault secret")
	errAzureSecretValueIsNil = errors.New("azure: got nil value for a secret")
)

func Azure(ctx context.Context, vaultName string) (
	map[string]string, error,
) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	vaultURI := fmt.Sprintf("https://%s.vault.azure.net/", vaultName)
	cli, err := azsecrets.NewClient(vaultURI, credential, nil)
	if err != nil {
		return nil, err
	}

	secrets := make(map[string]string)

	pager := cli.NewListSecretPropertiesPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, properties := range page.SecretPropertiesListResult.Value {
			if properties.ID == nil {
				return nil, fmt.Errorf("%w: %s",
					errAzurePropertyIdIsNil, vaultName,
				)
			}
			key := properties.ID.Name()
			res, err := cli.GetSecret(ctx, key, "", nil)
			if err != nil {
				return nil, err
			}
			if res.Value == nil {
				return nil, fmt.Errorf("%w: %s/%s",
					errAzureSecretValueIsNil, vaultName, key,
				)
			}

			secrets[strings.ReplaceAll(key, "-", "_")] = *res.Value
		}
	}

	return secrets, nil
}
