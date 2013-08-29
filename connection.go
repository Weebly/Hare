package hare

import (
	"errors"
	"github.com/streadway/amqp"
	"log"
)

/**
 * Retrieves a connection from the configured AMQP host, with the given ConnectionDefinition
 */
func GetConnection(definition ConnectionDefinition) error {
	connection, err := amqp.Dial("amqp://" + definition.Username + ":" + definition.Password + "@" + RABBITMQ_HOST + ":" + RABBITMQ_PORT + "/" + definition.Vhost)
	if err != nil {
		if err == amqp.ErrClosed {
			return retryAmqpConnection(definition)
		}
		return err
	}

	Connection[definition.Vhost] = connection
	getChannel(Connection[definition.Vhost], definition)
	return nil
}

/**
 * Retrieves a channel from the given connection, ConnectionDefinition
 */
func getChannel(connection *amqp.Connection, cd ConnectionDefinition) error {
	channel, err := connection.Channel()
	if err != nil {
		return err
	}

	Channel[cd.Vhost] = channel
	return nil
}

/**
 * Handles closed connection errors with a retry
 */
func retryAmqpConnection(connection ConnectionDefinition) error {
	log.Printf("AMQP connection dropped.")

	/**
	 * Don't want to loop through this for too long.
	 */
	if Statistics[RUNTIME_RABBITMQ_CONNECTION_FAILURE] > HARE_RABBITMQ_CONNECTION_RETRY_MAXIMUM {
		return errors.New("Maximum connection retry count has been reached.")
	}

	err := GetConnection(connection)
	if err != nil {
		return err
	}

	statisticsIncrement(RUNTIME_RABBITMQ_CONNECTION_FAILURE)
	log.Print("Reconnected to remote host.")
	return nil
}
