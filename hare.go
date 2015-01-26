package main

import (
	"log"
)

func main() {
	log.Printf("Hare started, connected to remote RabbitMQ host at %s.", RABBITMQ_HOST)

	/**
	 * Connect to our default vhost to ensure the remote host is up.
	 */
	defaultConnDef := &ConnectionDefinition{
		RABBITMQ_DEFAULT_VHOST,
		RABBITMQ_DEFAULT_USERNAME,
		RABBITMQ_DEFAULT_PASSWORD,
	}

	// Open a connection to the default vhost
	_, _, err := GetConnection(*defaultConnDef)
	if err != nil {
		log.Print(err)
	}

	/**
	 * We're ready to start our web API.
	 */
	if err := StartHttpApi(); err != nil {
		log.Fatalf("Failed to start webservice, %s.", err)
	}
}
