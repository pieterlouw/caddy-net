# Caddy Net #

TCP/UDP server type for [Caddy Server](https://github.com/mholt/caddy)

The server type is called `net`

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

**Rule:** A server block can only echo or proxy, not both.


### host directive ###

The `host` directive is the hostname/address of the site to serve, and is needed for TLS , especially in cases where the auto TLS feature [Let's encrypt](https://letsencrypt.org/) is used.

## TLS ##

This server type leverage the [tls directive](https://caddyserver.com/docs/tls) from the Caddy server and can be added to the server blocks as needed.


## Start/Run 

***Note***: When you start caddy you will need to specify the server type using the `-type` flag: `caddy -type=net`

## Status ##

*This server type plugin works as intended but is still considered BETA* 

***Note***: *Because the server type is still in early development the syntax for the Caddyfile might change, but will try to havee syntax above backward compatible.*

## Use cases ##

[Using Caddy To Create A Secure Socket Server](https://www.chaoswebs.net/blog/using-caddy-to-create-a-secure-socket-server.html)


## References ##

[Writing a Plugin: Server Type](https://github.com/mholt/caddy/wiki/Writing-a-Plugin:-Server-Type)

[Caddy Forum discussion](https://caddy.community/t/writing-a-tcp-udp-server-type-for-caddy/1589)


