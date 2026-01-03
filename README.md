# Webhook Relay

This tool is an executable that can receive webhooks and publish them the an AMQP queue, and also receive messages from the AMQP queue and call internal services using the original webhook.

This allows you to setup an external service that relays webhooks to an internal service.
