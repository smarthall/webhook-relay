package cmd

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/smarthall/webhook-relay/internal/messaging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	transmitterCmd.Flags().String("send-to", "http://localhost:8000", "URI to send webhooks to")
	viper.BindPFlag("send-to", transmitterCmd.Flags().Lookup("send-to"))

	transmitterCmd.Flags().Bool("extra-headers", true, "Send extra headers to the webhook host")
	viper.BindPFlag("extra-headers", transmitterCmd.Flags().Lookup("extra-headers"))

	transmitterCmd.Flags().Bool("preserve-host", false, "Preserve the original host header in the request")
	viper.BindPFlag("preserve-host", transmitterCmd.Flags().Lookup("preserve-host"))

	rootCmd.AddCommand(transmitterCmd)
}

var transmitterCmd = &cobra.Command{
	Use:   "transmitter",
	Short: "Transmitter listens to RabbitMQ and sends webhooks to a host",
	Long:  `Transmitter listens to RabbitMQ and sends webhooks to a host.`,
	Run: func(cmd *cobra.Command, args []string) {
		sub := messaging.NewSubscriber(viper.GetString("amqp"))
		msgs, err := sub.Subscribe()
		if err != nil {
			log.Panicf("Failed to consume messages: %s", err)
		}

		client := &http.Client{}

		for msg := range msgs {
			// TODO Subscriber should unmarshal the message
			var reqmsg messaging.RequestMessage
			err = json.Unmarshal(msg.Body, &reqmsg)
			if err != nil {
				log.Panicf("Failed to unmarshal message: %s", err)
			}

			// TODO This should be a method on RequestMessage
			// TODO: Replace send-to with something compatible with environment variables
			req, err := http.NewRequest(reqmsg.Method, viper.GetString("send-to"), nil)
			if err != nil {
				log.Panicf("Failed to create request: %s", err)
			}

			for k, v := range reqmsg.Headers {
				req.Header[k] = v
			}

			if viper.GetBool("extra-headers") {
				req.Header.Set("Relay-Original-Path", reqmsg.Path)
			}

			if viper.GetBool("extra-headers") {
				req.Header["Relay-Original-Host"] = []string{reqmsg.Host}
			}

			if viper.GetBool("preserve-host") {
				req.Host = reqmsg.Host
			}

			buf := bytes.NewBuffer([]byte(reqmsg.Body))
			req.Write(buf)

			client.Do(req)
		}

		// TODO: Implement a signal handler to gracefully shutdown the server
	},
}
