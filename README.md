# FS Cache

A helper program for managing and monitoring a cache file and updating it as
needed.

It works, but is rough around the edges.

## Why

CtrlP is amazing for jumping around a repo, but it really struggles with
listing all the files to fuzzy search from. For a long time I had a bit of
code in my ctrlp config that essentially boiled down to:

`[ cwd -e $HOME ] && "echo can't run ctrlp from home"`. 

This worked, and prevented me from locking up my VIM every time I had the
misfortune of invoking it from the home directory. Combine this with `git
ls-files` and I had a working solution.

That was until I started working in a monorepo with around 50k files. Even
with `git ls-files` CtrlP was struggling. Enabling caching made it manageable
but I was constantly struggling with the cache being out of date.

So `fscache` was born. Something that would monitor and maintain a cache. Then
provide some tools for reading from it and quickly filtering it down to
relevant data. This made working in the monorepo quick, and allowed me to hop
around from my home directory.

# Integrations

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

## FZF-VIM

To integrate with [fzf-vim](https://github.com/junegunn/fzf.vim) simply add
the following to your `vimrc`.

```vim
let $FZF_DEFAULT_COMMAND='fscache read -r'
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
