# Record from the repo root: vhs demo/demo.tape
Output docs/demo.gif
Set Shell bash
Set FontSize 18
Set Width 1100
Set Height 700
Set Padding 24
Set Theme "Catppuccin Mocha"
Set TypingSpeed 45ms

Hide
Type `export DEMO=$(cd "$(mktemp -d)" && pwd -P) && go build -o $DEMO/bin/ghx . && ln -s ghx $DEMO/bin/gh` Enter
Type `mkdir -p $DEMO/home/.local/share/gh/bin && cp demo/stub-gh/gh $DEMO/home/.local/share/gh/bin/` Enter
Type `export HOME=$DEMO/home PATH=$DEMO/bin:$DEMO/home/.local/share/gh/bin:$PATH GHX_NO_UPDATE=1` Enter
Type `mkdir -p ~/code/work/api/src ~/code/personal/blog ~/.config/ghx && cp examples/config.yaml ~/.config/ghx/` Enter
Type `cd ~/code && export PS1='\[\e[36m\]\w\[\e[0m\] $ ' && clear` Enter
Sleep 500ms
Show

Type "cat ~/.config/ghx/config.yaml | grep -v '^#'" Enter
Sleep 2.5s
Type "cd ~/code/work/api/src && gh ghx which" Enter
Sleep 2s
Type "gh api user --jq .login" Enter
Sleep 1.5s
Type "cd ~/code/personal/blog && gh api user --jq .login" Enter
Sleep 1.5s
Type "echo work > .ghx && gh api user --jq .login" Enter
Sleep 1.5s
Type "gh ghx which" Enter
Sleep 3s
