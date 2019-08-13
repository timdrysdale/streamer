# streamer

Websocket relay that can bridge between two servers or two clients that send binary data.

## Status

Alpha; no automated testing; do not use


## Usage

```
Usage:
  streamer [flags]

Flags:
  -b, --bufsize int          buffer size (max message size) [DEFAULT is 65535 bytes] (default 65535)
  -h, --help                 help for streamer
  -r, --receiver string      <ip>:<port> of the websocket server that will receive messages
  -s, --servers              make endpoints servers [DEFAULT is clients]
  -t, --transmitter string   <ip>:<port> of the websocket server that will transmit messages
  -v, --verbose              print connection and message logs [DEFAULT is quiet]
```

### Example

#### Scenario One

You want to forward data from one websocket server to another, so you want two linked clients to bridge them. Streamer can provide you with those two clients, and facilitate the uni-directional flow of data. 

```$ ./streamer -t ws://localhost:8001/ -r ws://localhost:8002/```

#### Scenario Two

You want to forward data from one websocket client to another, so you want two linked servers to bridge them. Streamer can provide you with those two servers, and facilitate the uni-directional flow of data. 

```$ ./streamer -t ws://localhost:8003/ -r ws://localhost:8004/ --servers```


