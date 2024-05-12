package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var amqpURI string

func init() {
	rootCmd.PersistentFlags().StringVar(&amqpURI, "amqp", "amqp://guest:guest@localhost:5672/", "AMQP URI for RabbitMQ")
}

var rootCmd = &cobra.Command{
	Use:   "relay",
	Short: "Webhook relay captures webhoos and publishes them to RabbitMQ",
	Long: `Webhook relay captures webhoos and publishes them to RabbitMQ.
				It is a simple tool that listens for incoming webhooks and
				forwards them to a RabbitMQ exchange. It also listens for
				messages from RabbitMQ and sends them to a webhook. This
				allows for many services to receive a single webhook.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
