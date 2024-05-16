package cmd

import (
	"log"
	"net/http"
	"time"

	"github.com/smarthall/webhook-relay/internal/messaging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	receiverCmd.Flags().String("listen", ":8080", "Address to listen on")
	viper.BindPFlag("listen", receiverCmd.Flags().Lookup("listen"))

	rootCmd.AddCommand(receiverCmd)
}

var receiverCmd = &cobra.Command{
	Use:   "receiver",
	Short: "Receives webhooks and forwards them to RabbitMQ",
	Long:  `Receiver listens for incoming webhooks and forwards them to a RabbitMQ exchange.`,
	Run: func(cmd *cobra.Command, args []string) {
		var pub = messaging.NewPublisher(viper.GetString("amqp"))

		s := &http.Server{
			Addr: viper.GetString("listen"),
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Printf("Received request at: %s", r.URL.Path)

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

		// TODO: Implement a signal handler to gracefully shutdown the server
	},
}
