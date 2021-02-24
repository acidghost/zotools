![zotools logo](logo.png)

[![build](https://github.com/acidghost/zotools/actions/workflows/ci.yml/badge.svg)](https://github.com/acidghost/zotools/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/acidghost/zotools/branch/main/graph/badge.svg?token=EXOHWAJHWW)](https://codecov.io/gh/acidghost/zotools)

This is a simple collection of tools to operate on a Zotero library.

Commands implemented:
- `sync`: creates or updates a local cache with useful info from the remote
  library
- `search`: searches via regular expression for items in the cached library
- `act`: performs an action on a selected result from a previous search

Please, feel free to copy, improve, distribute and share. Feedback and patches
are always welcome!

# Installation

* With `go get github.com/acidghost/zotools/cmd/...`
* By downloading the source and running `make build` followed by `make install`

# Usage

You need to provide a configuration file with the Zotero API key and various
paths the tool uses to operate (look at the template in the root of this
repository):
* `key` is the Zotero API key; you can get one from
  https://www.zotero.org/settings/keys
* `zotero` is the path to the folder where Zotero downloads all the attachments
* `storage` is the file `zotools` will use to store all its information (e.g.
  Zotero items, search results, etc.)

The configuration file can be passed via the command line (`-config` flag) or
via an environment variable (`ZOTOOLS`). The former overwrites the latter.

For detailed usage information try `zotools help` and `zotools cmd -h`.

## Use Cases

Search for an item and then open it. First issue `zotools search <term>` and
then `zotools act -i=<idx> zathura` to open the result numbered `idx` with
`zathura`.

## fzf

If you desire a more interactive experience than running `zotools` twice to
first search and then to act, you may have a look at
[fzf](https://github.com/junegunn/fzf). The function below uses fzf to
interactively search with `zotools` and act on your selection with one enter at
the correct line. To use it add the following function to your `.bashrc` (or
equivalent) and install fzf alongside `zotools`.

```shell
zotools_pdf() {
    local INITIAL_QUERY=$1
    local SEARCH_CMD="zotools search "
    FZF_DEFAULT_COMMAND="$SEARCH_CMD '$INITIAL_QUERY'" \
        fzf --bind "change:reload:$SEARCH_CMD {q} || true" \
            --ansi --disabled --query "$INITIAL_QUERY" \
            --height=50% --layout=reverse \
        | awk '{print $1}' | tr -d ')' \
        | xargs -I{} -o zotools act -i={} zathura
}
```
