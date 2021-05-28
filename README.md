# FS Cache

A helper program for managing and monitoring a cache file and updating it as
needed.

It works, but is rough around the edges.

## CtrlP & VIM

`fscache` can be used with VIM and CtrlP by using something like the following
setting.

```vim
let g:ctrlp_user_command = {
  \   'types': {
  \     1: ['.git', 'fscache read -p %s']
  \   },
  \   'fallback': 'fscache read'
  \ }
```

## FZF

`fscache` can be paired with FZF to make a nice and easy command for jumping
around your computer.

```zsh
function fzd() {
  local refresh="fscache read -d"

  if [ "$1" != "" ]; then
    local initial_query="-q $1"
  fi

  local dir=$(eval $refresh | fzf \
    --filepath-word \
    --bind='ctrl-t:next-history'\
    --bind='ctrl-r:previous-history'\
    --bind='ctrl-n:down' \
    --bind='ctrl-p:up' \
    --layout='reverse-list' \
    --tiebreak=length \
    --bind='change:first' \
    --header='Press F5 to reload' \
    --bind="f5:reload(${refresh})" \
    --history=$FZD_HISTORY $initial_query)

  if [ "$dir" != "" ];
  then
    pushd -q $dir
  fi
}
```
