# imgrep-web

[`imgrep`](https://github.com/keeferrourke/imgrep) is a command-line
utility written in Go. This repository contains a web-wrapper/interface
for the cli-tool.

![](.product/web_no_query.png) 

## Installation

 1. Install [`imgrep`](https://github.com/keeferrourke/imgrep).
 2. Install `imgrep-web`
 ```
 go get github.com/keeferrourke/imgrep-web
 go install github.com/keeferrourke/imgrep-web
 ```

### Web UI

`imgrep-web` comes with a familiar search-based web UI that interacts with
the pre-indexed sqlite database. This web-UI was added to make demo-ing
this utitility easier in a hackathon setting, and is a good POC for how
this `imgrep` may be used.

To start a server on localhost:1337:

```
imgrep server
```
Then just visit 'localhost:1337' in your favourite web brower ;)

#### Preview

![](.product/web_node_query.png)
![](.product/web_me_query.png)

## License

`imgrep` is free software licensed under the MIT license.

Copyright (c) 2017 Keefer Rourke, Ivan Zhang, and Thomas Dedinsky.

See LICENSE for details.
