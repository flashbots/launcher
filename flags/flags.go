package flags

import "strings"

const (
	AwsSecretArn = "aws-secret-arn"

	ULimitSoft = "ulimit-soft"
	ULimitHard = "ulimit-hard"
)

func Env(flag string) string {
	return strings.ReplaceAll(strings.ToUpper(flag), "-", "_")
}
