package main

import (
	"github.com/streadway/amqp"
	"log"
)

/**
 * Retrieves a connection from the configured AMQP host, with the given ConnectionDefinition
 */
func GetConnection(definition ConnectionDefinition) (*amqp.Connection, *amqp.Channel, error) {
	var conn *amqp.Connection
	var channel *amqp.Channel

	conn, ok := Connection[definition.Vhost]
	if !ok {
		err := Connect(definition)
		if err != nil {
			return nil, nil, err
		}
		conn = Connection[definition.Vhost]
	}

	channel, ok = Channel[definition.Vhost]
	if !ok {
		err := OpenChannel(conn, definition)
		if err != nil {
			return nil, nil, err
		}
		channel = Channel[definition.Vhost]
	}
	return conn, channel, nil
}

func Connect(definition ConnectionDefinition) error {
	log.Printf("Opening connection to %s:%s/%s\n", RABBITMQ_HOST, RABBITMQ_PORT, definition.Vhost)

	connection, err := amqp.Dial("amqp://" + definition.Username + ":" + definition.Password + "@" + RABBITMQ_HOST + ":" + RABBITMQ_PORT + "/" + definition.Vhost)
	if err != nil {
		log.Print(err)
		if err == amqp.ErrClosed {
			return retryAmqpConnection(definition)
		}
		return err
	}

	Connection[definition.Vhost] = connection
	err = OpenChannel(Connection[definition.Vhost], definition)
	return err
}

/**
 * Retrieves a channel from the given connection, ConnectionDefinition
 */
func OpenChannel(connection *amqp.Connection, cd ConnectionDefinition) error {
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
		log.Fatal("Maximum connection retry count has been reached.")
	}

	err := Connect(connection)
	if err != nil {
		return err
	}

	statisticsIncrement(RUNTIME_RABBITMQ_CONNECTION_FAILURE)
	log.Print("Reconnected to remote host.")
	return nil
}
