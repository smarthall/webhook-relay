package cmd

import (
	"log"
	"net/http"
	"time"

	"github.com/smarthall/webhook-relay/internal/messaging"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(receiverCmd)
}

var receiverCmd = &cobra.Command{
	Use:   "receiver",
	Short: "Receives webhooks and forwards them to RabbitMQ",
	Long:  `Receiver listens for incoming webhooks and forwards them to a RabbitMQ exchange.`,
	Run: func(cmd *cobra.Command, args []string) {
		var pub = messaging.NewPublisher(amqpURI)

		s := &http.Server{
			Addr: ":8080",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Publish the request to the message broker
				err := pub.Publish(*r)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// If the message was published successfully, return a 204 No Content response
				w.WriteHeader(http.StatusNoContent)
			}),
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		log.Fatal(s.ListenAndServe())
	},
}
