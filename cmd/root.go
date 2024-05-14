package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "relay",
	Short: "Webhook relay captures webhoos and publishes them to RabbitMQ",
	Long: `Webhook relay captures webhoos and publishes them to RabbitMQ.
				It is a simple tool that listens for incoming webhooks and
				forwards them to a RabbitMQ exchange. It also listens for
				messages from RabbitMQ and sends them to a webhook. This
				allows for many services to receive a single webhook.`,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("amqp", "amqp://guest:guest@localhost:5672/", "AMQP URI for RabbitMQ")
	viper.BindPFlag("amqp", rootCmd.PersistentFlags().Lookup("amqp"))

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $PWD/config.yaml)")

	rootCmd.PersistentFlags().Lookup("config")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
