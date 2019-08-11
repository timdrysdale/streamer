package cmd

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/spf13/cobra"
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
		spinner(100*time.Millisecond, 20)
		close(msg)
		wg.Wait()
	},
}

func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func HandleTransmitter(msg chan []byte, wg *sync.WaitGroup, t *url.URL) {
	defer wg.Done()

	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for t := range ticker.C {
			msg <- []byte(fmt.Sprintf("Tick at %v", t))
		}
	}()

}

func spinner(delay time.Duration, countdown int) {

	for {
		for _, r := range `-\|/` {

			fmt.Printf("\r%c %d  ", r, countdown)
			time.Sleep(delay)
			if countdown <= 0 {
				fmt.Printf("\r         \r")
				return
			}
			countdown--
		}
	}
}

func HandleReceiver(msg <-chan []byte, wg *sync.WaitGroup, t *url.URL) {
	defer wg.Done()
	for buf := range msg {
		fmt.Println(string(buf))
	}

}
