package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/flashbots/launcher/flags"
	"github.com/flashbots/launcher/logutils"
	"github.com/flashbots/launcher/process"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	version = "development"
)

func main() {
	var logFormat, logLevel string

	app := &cli.App{
		Name:    "launcher",
		Usage:   "Prepares the environment for the sub-process, launches, and manages it",
		Version: version,

		ArgsUsage: "/path/to/the/subprocess arg1 arg2 ...",
		HideHelp:  true,

		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Category: "Running:",
				EnvVars:  []string{flags.Env(flags.AwsSecretArn)},
				Name:     flags.AwsSecretArn,
				Usage:    "read values from AWS secret manager (`ARN`) and inject them into the environment",
			},

			&cli.StringSliceFlag{
				Category: "Running:",
				EnvVars:  []string{flags.Env(flags.AzureKeyVaultName)},
				Name:     flags.AzureKeyVaultName,
				Usage:    "read secrets from Azure key vault and inject them into the environment",
			},

			&cli.IntFlag{
				Category:    "Running:",
				DefaultText: "don't change",
				EnvVars:     []string{flags.Env(flags.ULimitSoft)},
				Name:        flags.ULimitSoft,
				Usage:       "set soft-limit on `number` of open files",
				Value:       -1,
			},

			&cli.IntFlag{
				Category:    "Running:",
				DefaultText: "don't change",
				EnvVars:     []string{flags.Env(flags.ULimitHard)},
				Name:        flags.ULimitHard,
				Usage:       "set hard-limit `number` of open files",
				Value:       -1,
			},

			&cli.StringFlag{
				Category:    "Logging:",
				Destination: &logLevel,
				EnvVars:     []string{"LOG_LEVEL"},
				Name:        "log-level",
				Usage:       "logging level",
				Value:       "info",
			},

			&cli.StringFlag{
				Category:    "Logging:",
				Destination: &logFormat,
				EnvVars:     []string{"LOG_MODE"},
				Name:        "log-mode",
				Usage:       "logging mode",
				Value:       "prod",
			},
		},

		Before: func(c *cli.Context) error {
			l, err := setupLogger(logLevel, logFormat)
			if err == nil {
				c.Context = logutils.ContextWithLogger(c.Context, l)
			} else {
				fmt.Fprintf(os.Stderr, "failed to configure the logging: %s\n", err)
			}
			return err
		},

		Action: func(c *cli.Context) error {
			if c.Args().Len() == 0 {
				return cli.ShowAppHelp(c)
			}
			p, err := process.New(c)
			if err != nil {
				return err
			}
			return p.Run(c.Context)
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Failed with error: %s\n", err)
		os.Exit(1)
	}
}

func setupLogger(level, mode string) (*zap.Logger, error) {
	var config zap.Config
	switch strings.ToLower(mode) {
	case "dev":
		config = zap.NewDevelopmentConfig()
	case "prod":
		config = zap.NewProductionConfig()
	default:
		return nil, fmt.Errorf("invalid log-mode '%s'", mode)
	}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("invalid log-level '%s': %w", level, err)
	}
	config.Level = logLevel

	l, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build the logger: %w", err)
	}
	zap.ReplaceGlobals(l)

	return l, nil
}
