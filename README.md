# Caddy Net #

*Note: This server type plugin is  currently still a work in progress (WIP).*

TCP/UDP server type for [Caddy Server](https://github.com/mholt/caddy)

The server type is called `net`

## Roadmap/TODO 

 * [x] Add UDP
 * [ ] Add TLS
 * [ ] Add auto - TLS (Let's Encrypt)

## Proposed Caddyfile 

```
echo :22017 {
}

proxy :12017 :22017 {
    # proxy config
}

```

The first server block will listen on port `22017` and echo any traffic back to caller

The second server block will listen on port `12017` and forward traffic to address `:22017`

*Rule: A server block can only echo or proxy, not both.*

## References ##

[Writing a Plugin: Server Type](https://github.com/mholt/caddy/wiki/Writing-a-Plugin:-Server-Type)

[Caddy Forum discussion](https://forum.caddyserver.com/t/writing-a-tcp-udp-server-type-for-caddy/1589)


