package cmd

import (
	"context"
	//"errors"
	"fmt"
	"io"
	//"log"
	//"math"
	//"net"
	//"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	//"time"

	"github.com/spf13/cobra"
	//"golang.org/x/time/rate"
	//"golang.org/x/xerrors"
	"nhooyr.io/websocket"
	//"nhooyr.io/websocket/wsjson"
)

var transmitter string
var receiver string
var verbose bool

func init() {
	rootCmd.PersistentFlags().StringVarP(&transmitter, "transmitter", "t", "", "<ip>:<port> of the websocket server that will transmit messages")
	rootCmd.MarkFlagRequired("transmitter")
	rootCmd.PersistentFlags().StringVarP(&receiver, "receiver", "r", "", "<ip>:<port> of the websocket server that will receive messages")
	rootCmd.MarkFlagRequired("receiver")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "print connection and message logs")

}

var rootCmd = &cobra.Command{
	Use:   "streamer",
	Short: "Streamer connects websocket servers",
	Long: `Streamer is a dual websocket client that allows 
         two servers to communicate without needing any client functionality `,
	Run: func(cmd *cobra.Command, args []string) {

		// parse urls
		// see https://www.alexedwards.net/blog/validation-snippets-for-go#url-validation)
		t, err := url.Parse(transmitter)
		if err != nil {
			panic(err)
		} else if t.Scheme == "" || t.Host == "" {
			fmt.Println("error: transmitter must be an absolute URL")
			return
		} else if t.Scheme != "ws" && t.Scheme != "wss" {
			fmt.Println("error: transmitter must begin with ws or wss")
			return
		}

		r, err := url.Parse(receiver)
		if err != nil {
			panic(err)
		} else if r.Scheme == "" || r.Host == "" {
			fmt.Println("error: receiver must be an absolute URL")
			return
		} else if r.Scheme != "ws" && r.Scheme != "wss" {
			fmt.Println("error: receiver must begin with ws or wss")
			return
		}

		if verbose {
			fmt.Println(t.Host, " is sending to ", r.Host)
		}

		var wg sync.WaitGroup
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		msg := make(chan []byte)

		go func() {
			for _ = range c {

				close(msg)
				wg.Wait()
				os.Exit(1)

			}
		}()

		wg.Add(2)
		go HandleTransmitter(msg, &wg, t)
		go HandleReceiver(msg, &wg, r)
		wg.Wait()
	},
}

func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func HandleTransmitter(msg chan<- []byte, wg *sync.WaitGroup, t *url.URL) {
	//TODO signal with a different channel so it does not panic on writing to closed channel after SIGINT
	defer wg.Done()

	var buf = make([]byte, 65535+1) //1000kbps rate at 30fps is just over 4200byte/s
	var n int

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fmt.Println("connecting to", t.String())
	c, _, err := websocket.Dial(ctx, t.String(), websocket.DialOptions{})
	//c.SetReadLimit(65535)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer c.Close(websocket.StatusInternalError, fmt.Sprintf("Internal error with websocket client %s", t.String()))
	fmt.Println("in handle transmitter")
	for {
		fmt.Println("trying to read from transmitter")
		typ, r, err := c.Reader(ctx)

		if err != nil {

			fmt.Println("tx: io.Reader", err)
			return
		}

		if typ != websocket.MessageBinary {
			fmt.Println("Not binary")
			//return
		}

		n, err = r.Read(buf)
		fmt.Println("Got: ", n)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Read:", err)
			}

		}

		//fmt.Println(buf[:n])
		msg <- buf[:n]

	}

	c.Close(websocket.StatusNormalClosure, "")
	return

}

func HandleReceiver(msg <-chan []byte, wg *sync.WaitGroup, t *url.URL) {
	for pkt := range msg {
		fmt.Println(len(pkt))
	}
}

func HandleReceiverOld(msg <-chan []byte, wg *sync.WaitGroup, t *url.URL) {
	// NOTE that nhooyr websockets must read messages sent to them else
	// control frames are not processed ... so bidirectional is needed
	defer wg.Done()

	//var buf = make([]byte, 65535+1) //1000kbps rate at 30fps is just over 4200byte/s
	var n int

	//ctx, cancel := context.WithTimeout(context.Background(), 18*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, _, err := websocket.Dial(ctx, t.String(), websocket.DialOptions{})
	//c.SetReadLimit(65535)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer c.Close(websocket.StatusInternalError, fmt.Sprintf("Internal error with websocket client %s", t.String()))
	fmt.Println("In Handle Receiver")

	//assume we don't get any messages so we must keep extending the readDeadline
	//ticker := time.NewTicker(30 * time.Second)
	//go func() {
	//	for t := range ticker.C {
	//		//ctx.SetReadDeadline(time.Minute)
	//		c.setReadTimeout <- context.Background()
	//	}
	//}()
	// read but ignore any messages we do receive
	//go func() {
	//	typ, r, err := c.Reader(ctx)
	//
	//	if err != nil {
	//
	//		fmt.Println("io.Reader", err)
	//		return
	//	}
	//
	//	if typ != websocket.MessageBinary {
	//		fmt.Println("Not binary")
	//		//return
	//	}
	//
	//	n, err = r.Read(buf)
	//	if err != nil {
	//		if err != io.EOF {
	//			fmt.Println("Read:", err)
	//		}
	//
	//	} else {
	//		fmt.Println("Got: ", n)
	//	}
	//
	//}()

	for pkt := range msg {
		fmt.Printf("%T %d\n", pkt, len(pkt))
		fmt.Println("attempting to read from channel")
		//func (c *Conn) Writer(ctx context.Context, typ MessageType) (io.WriteCloser, error)
		w, err := c.Writer(ctx, websocket.MessageBinary)
		fmt.Println("got writer")
		if err != nil {

			fmt.Println("io.Writer", err)
			return
		}
		fmt.Println("pkt size is", len(pkt))
		//nn := int64(math.Min(float64(len(pkt)), float64(65535)))
		n, err = w.Write(pkt[:511])

		fmt.Println("wrote pkt length", n)

		if err != nil {
			if err != io.EOF {
				fmt.Println("Write:", err)
			}
		}

		err = w.Close()
		fmt.Println("closed writer")
		if err != nil {
			fmt.Println("Closing Write failed:", err)
		}

	}

	c.Close(websocket.StatusNormalClosure, "")
	return

}
