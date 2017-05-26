# Caddy Net #

*Note: This server type plugin is  currently still a work in progress (WIP).*

TCP/UDP server type for [Caddy Server](https://github.com/mholt/caddy)

The server type is called `net`

## Roadmap/TODO 

 * [X] Add UDP
 * [X] Add TLS
 * [X] Add auto - TLS (Let's Encrypt)

## Proposed Caddyfile 

```
echo :22017 {
    host echo.example.com
}

proxy :12017 :22017 {
    host proxy.example.com
}
```

The first server block will listen on port `22017` and echo any traffic back to caller

The second server block will listen on port `12017` and forward traffic to address `:22017`

*Rule: A server block can only echo or proxy, not both.*

***Note***: *When you start caddy you will need to specify they server type.* `caddy -type=net`

This server type leverage the [tls directive](https://caddyserver.com/docs/tls) from the Caddy server and can be added to the server blocks as needed. 

## References ##

[Writing a Plugin: Server Type](https://github.com/mholt/caddy/wiki/Writing-a-Plugin:-Server-Type)

[Caddy Forum discussion](https://forum.caddyserver.com/t/writing-a-tcp-udp-server-type-for-caddy/1589)


