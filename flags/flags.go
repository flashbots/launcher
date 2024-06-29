package flags

import "strings"

const (
	AwsSecretArn      = "aws-secret-arn"
	AzureKeyVaultName = "azure-key-vault-name"

	ULimitSoft = "ulimit-soft"
	ULimitHard = "ulimit-hard"
)

func Env(flag string) string {
	return strings.ReplaceAll(strings.ToUpper(flag), "-", "_")
}
