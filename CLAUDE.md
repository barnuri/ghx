# ghx

## Local install

After every code change, build from source and replace the local binary:

```sh
go build -trimpath -o ~/.local/bin/ghx .
```

The local build reports version `dev`. The self-updater skips `dev`, so a release never replaces it.
