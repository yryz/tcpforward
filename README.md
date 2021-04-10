安全的tcp转发，支持tls双向认证，支持Socks代理。
可以用来作为内网跳板，也可以用来对普通tcp增加tls安全认证。

**tcp -> tcp**
local: `tcpforward -t tcp -l 127.0.0.1:1080 -T tcp -P 192.168.10.1:1080`

**tcp -> |tls| -> tcp**

1. gen cert.pem & key.pem: `tcpforward -gencert`
2. local: `tcpforward -t tcp -l 127.0.0.1:1080 -T tls -P 192.168.10.1:2080 -C cert.pem -K key.pem`
3. remote: `tcpforward -t tls -l 192.168.10.1:2080 -T tcp -P 127.0.0.1:1080 -C cert.pem -K key.pem`

**tcp -> |socks| -> |tls| -> tcp**

1. gen cert.pem & key.pem: `tcpforward -gencert`
2. local: `ALL_PROXY=socks5://username:passwd@host:port tcpforward -t tcp -l 127.0.0.1:1080 -T tls -P 192.168.10.1:2080 -C cert.pem -K key.pem`
3. remote: `tcpforward -t tls -l 192.168.10.1:2080 -T tcp -P 127.0.0.1:1080 -C cert.pem -K key.pem`

