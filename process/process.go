package process

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/flashbots/launcher/flags"
	"github.com/flashbots/launcher/logutils"
	"github.com/flashbots/launcher/secret"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

type Process struct {
	command []string
	secrets map[string]string

	doSetUlimit bool
	ulimit      syscall.Rlimit
}

func New(c *cli.Context) (*Process, error) {
	ctx := c.Context

	p := &Process{
		command: c.Args().Slice(),
	}

	l := logutils.LoggerFromContext(ctx)
	defer l.Sync() //nolint:errcheck

	// fetch secrets

	secrets := make(map[string]string)
	for _, arn := range c.StringSlice(flags.AwsSecretArn) {
		start := time.Now()
		moreSecrets, err := secret.AWS(ctx, arn)
		if err != nil {
			l.Error("Failed to fetch AWS secret",
				zap.Error(err),
				zap.String("arn", arn),
			)
			return nil, err
		}
		l.Debug("Fetched AWS secret",
			zap.String("arn", arn),
			zap.Duration("duration-ms", time.Duration(time.Now().Sub(start).Milliseconds())),
		)
		for k, v := range moreSecrets {
			if _, collision := secrets[k]; collision {
				l.Warn("Secrets key collision detected",
					zap.String("key", k),
				)
			}
			secrets[k] = v
		}
	}
	p.secrets = secrets

	// figure out ulimit

	if c.Int(flags.ULimitSoft) > 0 || c.Int(flags.ULimitHard) > 0 {
		p.doSetUlimit = true

		err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &p.ulimit)
		if err != nil {
			l.Error("Failed to read the ulimit",
				zap.Error(err),
			)
			return nil, err
		}
		if c.Int(flags.ULimitSoft) > 0 {
			p.ulimit.Cur = uint64(c.Int(flags.ULimitSoft))
		}
		if c.Int(flags.ULimitHard) > 0 {
			p.ulimit.Max = uint64(c.Int(flags.ULimitHard))
		}
	}

	return p, nil
}

func (p *Process) Run(ctx context.Context) error {
	l := logutils.LoggerFromContext(ctx)
	defer l.Sync() //nolint:errcheck

	env := os.Environ()
	for k, v := range p.secrets {
		if len(os.Getenv(k)) > 0 {
			l.Warn("Overwriting already existing environment variable",
				zap.String("key", k),
			)
		}
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	target := p.command[0]
	if _, err := os.Stat(target); err != nil && errors.Is(err, os.ErrNotExist) {
		if strings.ContainsRune(target, os.PathSeparator) {
			return err
		}
		// try to find it under PATH
		found := false
		for _, path := range strings.Split(os.Getenv("PATH"), string(os.PathListSeparator)) {
			candidate := filepath.Join(path, target)
			if _, errStat := os.Stat(candidate); errStat == nil {
				found = true
				target = candidate
				break
			}
		}
		if !found {
			return err
		}
	}

	if p.doSetUlimit {
		if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &p.ulimit); err != nil {
			l.Error("Failed to modify the ulimit",
				zap.Error(err),
				zap.Uint64("hard", p.ulimit.Max),
				zap.Uint64("soft", p.ulimit.Cur),
			)
			return err
		}
		l.Debug("Set the new ulimit",
			zap.Uint64("hard", p.ulimit.Max),
			zap.Uint64("soft", p.ulimit.Cur),
		)
	}

	l.Debug("Launching the process",
		zap.String("target", target),
		zap.Strings("cmd", p.command),
	)
	return syscall.Exec(target, p.command, env)
}
