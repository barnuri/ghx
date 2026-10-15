# ghx

Use the right GitHub account in every folder. `ghx` is a thin wrapper around the official
[`gh`](https://cli.github.com) CLI. It picks a gh account from the directory you are in, then
runs the real `gh` with your arguments unchanged.

![ghx demo](docs/demo.gif)

- Every gh command, flag and extension works, because ghx passes them through as-is.
- gh keeps its own updates. ghx never patches or bundles it.
- macOS, Linux and Windows. One static binary, no runtime.
- Installs as `gh`, so scripts, editors and aliases use it with no changes.
- Updates itself in the background, once a day.

## Install

macOS / Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/barnuri/ghx/master/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/barnuri/ghx/master/install.ps1 | iex
```

The installer downloads the latest release for your OS and CPU, checks its sha256, and
installs `ghx` plus a `gh` shim that points to it:

| OS | Install dir | `gh` shim |
|---|---|---|
| macOS / Linux | `~/.local/bin` | symlink `gh -> ghx` |
| Windows | `%LOCALAPPDATA%\Programs\ghx` | copy `gh.exe` |

Put the install dir before the real gh in `PATH`. ghx skips its own dir when it looks for the
real gh, so the shim never calls itself. Set `GHX_SHIM=0` to install only `ghx`, or
`GHX_INSTALL_DIR` to pick another dir. On Windows, run the installer as Administrator when gh
is in the system PATH, because Windows searches the system PATH before the user PATH.

With Go: `go install github.com/barnuri/ghx@latest`, then `alias gh=ghx`.

## Configuration

1. Log in to each account once. gh stores all of them:

   ```sh
   gh auth login   # work account
   gh auth login   # personal account
   ```

2. Map folders to accounts in `~/.config/ghx/config.yaml` (Windows:
   `%APPDATA%\ghx\config.yaml`, or set `GHX_CONFIG`):

   ```yaml
   profiles:
     work:
       user: my-work-login
     personal:
       user: my-personal-login

   dirs:
     ~/code/work: work
     ~/code/personal: personal
   ```

Now every folder under `~/code/work`, at any depth, uses the work account, and every folder
under `~/code/personal` uses the personal account:

```sh
cd ~/code/work/api/src && gh pr list        # as my-work-login
cd ~/code/personal/blog && gh repo view     # as my-personal-login
```

Full example: [examples/config.yaml](examples/config.yaml).

| Key | Meaning |
|---|---|
| `profiles.<name>.user` | gh login, as shown by `gh auth status` |
| `profiles.<name>.host` | GitHub Enterprise host. Default: `github.com` |
| `dirs` | folder -> profile. The deepest matching folder wins |
| `default` | profile for folders outside every `dirs` entry. Default: plain gh |
| `manage_gh` | `true` makes ghx download gh and keep it updated. Default: use gh from `PATH` |
| `disable_updates` | `true` stops the daily ghx update check |

### Per-repo file: `.ghx`

Put a `.ghx` file in a repo to choose its profile. It holds one line: a profile name, or a gh
login if no profile has that name.

```sh
echo personal > .ghx
echo .ghx >> .git/info/exclude   # or add it to .gitignore
```

ghx uses the nearest `.ghx` in the current folder or any parent folder.

### Resolution order

1. `GHX_PROFILE` environment variable
2. Nearest `.ghx` file
3. Deepest matching `dirs` entry
4. `default`
5. No match: plain gh, with your active gh account

## Commands

ghx adds one command group, `gh ghx`. Everything else goes to gh.

```text
gh ghx which     Show the profile, gh user and gh binary for the current folder
gh ghx update    Update ghx now (and gh when manage_gh is true)
gh ghx version   Show the ghx version
```

## How it works

ghx runs `gh auth token --user <login>` and starts the real gh with `GH_TOKEN` set
(`GH_ENTERPRISE_TOKEN` and `GH_HOST` for an Enterprise host). It never runs `gh auth switch`,
so terminals in different folders do not change each other's account.

- `gh auth ...` commands run without the token, so `gh auth login` and `gh auth status` work as
  usual. `gh auth git-credential` gets the token, so git pushes as the folder's account.
- A `GH_TOKEN` you set yourself wins over ghx.
- Exit codes, stdin, stdout and the TTY belong to gh. On macOS and Linux ghx replaces itself
  with gh (`exec`).

To make `git push` use the folder's account, point git at the shim:

```sh
git config --global credential.https://github.com.helper ''
git config --global --add credential.https://github.com.helper '!gh auth git-credential'
```

## Updates

Once a day ghx starts a background process that downloads the latest ghx release, checks its
sha256 and replaces the binary. With `manage_gh: true` it does the same for gh from
[cli/cli releases](https://github.com/cli/cli/releases). The result goes to `update.log` in the
ghx cache dir (`~/Library/Caches/ghx`, `~/.cache/ghx` or `%LOCALAPPDATA%\ghx`).
Set `GHX_NO_UPDATE=1` or `disable_updates: true` to turn it off.

## Development

```sh
go test ./...
vhs demo/demo.tape   # re-record docs/demo.gif with a stub gh and demo accounts
```

Every push to `master` runs the tests on macOS, Linux and Windows, then publishes release
`v0.1.<run>` with binaries for darwin, linux and windows on amd64 and arm64.

## License

MIT
