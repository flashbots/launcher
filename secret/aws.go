package secret

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

const (
	timeout      = time.Second
	versionStage = "AWSCURRENT"
)

func AWS(ctx context.Context, arn string) (map[string]string, error) {
	if arn == "test" {
		return map[string]string{
			"_ANSWER":   "42",
			"_QUESTION": "The Ultimate Question of Life, the Universe, and Everything",
		}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	if len(cfg.Region) == 0 {
		// 0   1   2              3         4          5      6
		// arn:aws:secretsmanager:${REGION}:${ACCOUNT}:secret:${SECRET}
		parts := strings.Split(arn, ":")
		if len(parts) != 7 {
			return nil, fmt.Errorf("secret's ARN seems to be corrupt: %s", arn)
		}
		cfg.Region = parts[3]
	}

	cli := secretsmanager.NewFromConfig(cfg)

	res, err := cli.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(arn),
		VersionStage: aws.String(versionStage),
	})
	if err != nil {
		return nil, err
	}
	if res.SecretString == nil || len(*res.SecretString) == 0 {
		return nil, fmt.Errorf("no secret or secret is empty")
	}

	var secrets map[string]string
	err = json.Unmarshal([]byte(*res.SecretString), &secrets)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal the secret '%s': %w", *res.SecretString, err)
	}

	return secrets, nil
}
