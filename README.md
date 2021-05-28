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
