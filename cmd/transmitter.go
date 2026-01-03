package cmd

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/smarthall/webhook-relay/internal/messaging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	transmitterCmd.Flags().String("key", "#", "The key to subscribe to")
	viper.BindPFlag("key", transmitterCmd.Flags().Lookup("key"))

	transmitterCmd.Flags().String("send-to", "http://localhost:8000", "URI to send webhooks to")
	viper.BindPFlag("send-to", transmitterCmd.Flags().Lookup("send-to"))

	transmitterCmd.Flags().Bool("insecure", false, "Skip SSL verification")
	viper.BindPFlag("insecure", transmitterCmd.Flags().Lookup("insecure"))

	transmitterCmd.Flags().Bool("extra-headers", true, "Send extra headers to the webhook host")
	viper.BindPFlag("extra-headers", transmitterCmd.Flags().Lookup("extra-headers"))

	transmitterCmd.Flags().Bool("preserve-host", false, "Preserve the original host header in the request")
	viper.BindPFlag("preserve-host", transmitterCmd.Flags().Lookup("preserve-host"))

	rootCmd.AddCommand(transmitterCmd)
}

// processDelivery handles a single AMQP delivery: it unmarshals the message and
// sends the contained HTTP request to the destination host.
func processDelivery(msg amqp.Delivery, client *http.Client, sendTo string, extraHeaders bool, preserveHost bool) error {
	var reqmsg messaging.RequestMessage
	if err := json.Unmarshal(msg.Body, &reqmsg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	buf := bytes.NewBuffer([]byte(reqmsg.Body))
	req, err := http.NewRequest(reqmsg.Method, sendTo, buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range reqmsg.Headers {
		req.Header[k] = v
	}

	if extraHeaders {
		req.Header.Set("Relay-Original-Path", reqmsg.Path)
		req.Header["Relay-Original-Host"] = []string{reqmsg.Host}
	}

	if preserveHost {
		req.Host = reqmsg.Host
	}

	log.Printf("Sending request to: %s", req.URL.String())
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	log.Printf("Received response: %s", response.Status)
	return nil
}

var transmitterCmd = &cobra.Command{
	Use:   "transmitter",
	Short: "Transmitter listens to RabbitMQ and sends webhooks to a host",
	Long:  `Transmitter listens to RabbitMQ and sends webhooks to a host.`,
	Run: func(cmd *cobra.Command, args []string) {
		sub := messaging.NewSubscriber(viper.GetString("amqp"), viper.GetString("key"))
		msgs, err := sub.Subscribe()
		if err != nil {
			log.Panicf("Failed to consume messages: %s", err)
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: viper.GetBool("insecure")},
		}
		client := &http.Client{Transport: tr}

		// Create a context that cancels on SIGINT or SIGTERM
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		// Create and start a health checker for this instance. If the
		// health checker signals failure the process will gracefully shut down.
		instanceID := viper.GetString("instance_id")
		if instanceID == "" {
			if hn, err := os.Hostname(); err == nil {
				instanceID = hn
			} else {
				instanceID = "transmitter"
			}
		}

		hc := messaging.NewHealthChecker(viper.GetString("amqp"), instanceID, 1*time.Second, 2*time.Second)
		defer hc.Stop()

		// Process messages until the channel closes or we receive a shutdown signal
		for {
			select {
			case <-ctx.Done():
				log.Printf("shutdown signal received, stopping transmitter")
				return
			case <-hc.Failure:
				log.Printf("healthcheck failure detected, initiating shutdown")
				stop()
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Printf("message channel closed, exiting")
					return
				}
				if err := processDelivery(msg, client, viper.GetString("send-to"), viper.GetBool("extra-headers"), viper.GetBool("preserve-host")); err != nil {
					log.Printf("Failed to process message: %v", err)
					continue
				}
			}
		}
	},
}
