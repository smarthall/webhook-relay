package cmd

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/smarthall/webhook-relay/internal/messaging"
	"github.com/spf13/cobra"
)

var sendTo string
var extraHeaders bool

func init() {
	transmitterCmd.Flags().StringVarP(&sendTo, "send-to", "", "http://localhost:8000", "URI to send webhooks to")
	transmitterCmd.Flags().BoolVarP(&extraHeaders, "extra-headers", "", true, "Send extra headers to the webhook host")
	rootCmd.AddCommand(transmitterCmd)
}

var transmitterCmd = &cobra.Command{
	Use:   "transmitter",
	Short: "Transmitter listens to RabbitMQ and sends webhooks to a host",
	Long:  `Transmitter listens to RabbitMQ and sends webhooks to a host.`,
	Run: func(cmd *cobra.Command, args []string) {
		sub := messaging.NewSubscriber(amqpURI)
		msgs, err := sub.Subscribe()
		if err != nil {
			log.Panicf("Failed to consume messages: %s", err)
		}

		client := &http.Client{}

		for msg := range msgs {
			var reqmsg messaging.RequestMessage
			err = json.Unmarshal(msg.Body, &reqmsg)
			if err != nil {
				log.Panicf("Failed to unmarshal message: %s", err)
			}

			req, err := http.NewRequest(reqmsg.Method, sendTo, nil)
			if err != nil {
				log.Panicf("Failed to create request: %s", err)
			}

			for k, v := range reqmsg.Headers {
				req.Header[k] = v
			}

			if extraHeaders {
				req.Header.Set("Relay-Original-Path", reqmsg.Path)
			}

			buf := bytes.NewBuffer([]byte(reqmsg.Body))
			req.Write(buf)

			client.Do(req)
		}
	},
}
