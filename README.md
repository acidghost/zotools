# zotools

This is a simple collection of tools to operate on a Zotero library.

Commands implemented:
- `sync`: creates or updates a local cache with useful info from the remote library
- `search`: searches via regular expression for items in the cached library

## Usage

You need to provide a configuration file with the Zotero API key and various
paths the tool uses to operate (look at the template in the root of this
repository).

The configuration file can be passed via the command line (`-config` flag) or
via an environment variable (`ZOTOOLS`). The former overwrites the latter.

For detailed usage information try `zotools help`.
