package main

import (
	"github.com/streadway/amqp"
)

const (
	/**
	 * Statistics related strings
	 */
	RUNTIME_PUBLISHES_RECEIVED              = "runtime_publishes_received"
	RUNTIME_PUBLISHES_SUCCESS               = "runtime_publishes_success"
	RUNTIME_PUBLISHES_FAILURE               = "runtime_publishes_failure"
	RUNTIME_PUBLISHES_FAILURE_400           = "runtime_publishes_failure_bad_request"
	RUNTIME_PUBLISHES_FAILURE_401           = "runtime_publishes_failure_no_credentials"
	RUNTIME_PUBLISHES_FAILURE_403           = "runtime_publishes_failure_bad_credentials"
	RUNTIME_PUBLISHES_FAILURE_404           = "runtime_publishes_failure_not_found"
	RUNTIME_PUBLISHES_FAILURE_500           = "runtime_publishes_failure_publish_error"
	RUNTIME_RABBITMQ_CONNECTION_FAILURE     = "runtime_rabbitmq_connection_failure"
)

const (
	ERROR_PUBLISH_NO_POST_BODY      = "No Post Body"
	ERROR_PUBLISH_BAD_AUTHORIZATION = "Bad Authorization"
	ERROR_PUBLISH_NO_AUTHORIZATION  = "No Authorization"
)

/**
 * Statistics Collector
 */
var Statistics = make(map[string]uint64)

/**
 * AMQP connection and channel
 */
type ConnectionDefinition struct {
	Vhost    string
	Username string
	Password string
}

var Connection = make(map[string]*amqp.Connection)
var Channel = make(map[string]*amqp.Channel)
