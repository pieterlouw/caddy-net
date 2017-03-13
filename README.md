# Caddy Net #



TCP/UDP  server type for [Caddy Server](https://github.com/mholt/caddy)

The server type is called `net`

*Note: This server type plugin is currently in working progress.*

## Roadmap/TODO ## 

 * [ ] Add UDP
 * [ ] Add SSL

## Proposed Caddyfile ## 

```
echo :8080 {
}

proxy :12017 :22017 {
    # proxy config
}

```

The first server block will listen on port `8080` and echo any traffic back to caller

The second server block will listen on port `12017` and forward traffic to address `:22017`

*Rule:A server block can only echo or proxy, not both.*

## References ##

[Writing a Plugin: Server Type](https://github.com/mholt/caddy/wiki/Writing-a-Plugin:-Server-Type)

[Caddy Forum discussion](https://forum.caddyserver.com/t/server-types-other-than-http/65)


