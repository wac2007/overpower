package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

var ipCmd = &cobra.Command{
	Use:   "ip",
	Short: "Check ip status",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel((context.Background()))
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		defer func() {
			signal.Stop((c))
			cancel()
		}()

		go func() {
			select {
			case <-c:
				cancel()
			case <-ctx.Done():
			}
		}()

		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err != nil {
			log.Fatal(err)
		}

		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		fmt.Print(localAddr)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(ipCmd)
}
