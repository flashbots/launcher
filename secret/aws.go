package secret

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

const (
	timeout      = 5 * time.Second
	versionStage = "AWSCURRENT"
)

var (
	errAwsSecretEmpty             = errors.New("aws: no secret or secret is empty")
	errAwsSecretFailedToUnmarshal = errors.New("aws: failed to unmarshal the secret")
	errAwsSecretInvalidArn        = errors.New("aws: secret's ARN seems to be corrupt")
)

func AWS(ctx context.Context, arn string) (
	map[string]string, error,
) {
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
			return nil, fmt.Errorf("%w: %s",
				errAwsSecretInvalidArn, arn,
			)
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
		return nil, errAwsSecretEmpty
	}

	var secrets map[string]string
	err = json.Unmarshal([]byte(*res.SecretString), &secrets)
	if err != nil {
		return nil, fmt.Errorf("%w: %w: %s",
			errAwsSecretFailedToUnmarshal, err, *res.SecretString,
		)
	}

	return secrets, nil
}
