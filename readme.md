# launcher

Prepare the environment for a sub-process and launch it:

- Read secrets from AWS secret manager and inject them into the environment.
- Set max open file limit.

## TL;DR

```shell
go run github.com/flashbots/launcher/cmd \
    --aws-secret-arn test \
  bash -c 'echo "${_ANSWER} is the answer to ${_QUESTION}"'
```

```text
42 is the answer to The Ultimate Question of Life, the Universe, and Everything
```

Also:

```shell
go run github.com/flashbots/launcher/cmd \
    --ulimit-soft 1024 \
    --ulimit-hard 4096 \
  bash -c "ulimit -Sn; ulimit -Hn"
```

```text
1024
4096
```
