package main

import (
	"encoding/base64"
	"errors"
	"github.com/streadway/amqp"
	"net/http"
	"strconv"
	"strings"
	"time"
)

/**
 * Prepares and dispatches a response via ResponseWriter for a publish API request
 */
func apiRequestPublish(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")

	statisticsIncrement(RUNTIME_PUBLISHES_RECEIVED)

	message, err := getMessageFromPost(r)
	if err != nil {
		statisticsIncrement(RUNTIME_PUBLISHES_FAILURE)
		statisticsIncrement(RUNTIME_PUBLISHES_FAILURE_400)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	auth, err := getAuthentication(r)
	if err != nil {
		header := http.Header{}
		header.Add("WWW-Authenticate", "Basic realm=hare")
		w.WriteHeader(http.StatusUnauthorized)
	}

	vhost := getUriValue(path, 1)
	if len(vhost) < 1 {
		statisticsIncrement(RUNTIME_PUBLISHES_FAILURE)
		statisticsIncrement(RUNTIME_PUBLISHES_FAILURE_404)
		http.NotFound(w, r)
		return
	}

	definition := &ConnectionDefinition{
		vhost,
		auth[0],
		auth[1],
	}

	if Connection[path[1]] == nil {
		err = GetConnection(*definition)
		if err != nil {
			statisticsIncrement(RUNTIME_PUBLISHES_FAILURE)
			if err == amqp.ErrVhost || err == amqp.ErrCredentials {
				statisticsIncrement(RUNTIME_PUBLISHES_FAILURE_AUTHORIZATION)
				w.WriteHeader(http.StatusForbidden)
				return
			} else {
				statisticsIncrement(RUNTIME_PUBLISHES_FAILURE_500)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			return
		}
	}

	exchange := getUriValue(path, 2)
	routingKey := getUriValue(path, 3)

	if err := publish(*definition, exchange, routingKey, *message); err != nil {
		statisticsIncrement(RUNTIME_PUBLISHES_FAILURE)
		statisticsIncrement(RUNTIME_PUBLISHES_FAILURE_500)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	statisticsIncrement(RUNTIME_PUBLISHES_SUCCESS)
	w.WriteHeader(http.StatusNoContent)
	return
}

/**
 * Publishes the given message to RabbitMQ with the given connnection definition
 */
func publish(connection ConnectionDefinition, exchange string, routingKey string, message amqp.Publishing) error {
	if err := Channel[connection.Vhost].Publish(
		exchange,
		routingKey,
		false,
		false,
		message,
	); err != nil {
		if err == amqp.ErrClosed {
			retryAmqpConnection(connection)
		} else {
			return err
		}
	}

	return nil
}

/**
 * Parses the given request's form values to be used for the Publishing struct
 */
func getMessageFromPost(r *http.Request) (*amqp.Publishing, error) {
	if len(r.FormValue("body")) < 1 {
		return &amqp.Publishing{}, errors.New(ERROR_PUBLISH_NO_POST_BODY)
	}

	/**
	 * Deal with non-string type conversions
	 */
	dm, err := strconv.Atoi(r.FormValue("delivery_mode"))
	if err != nil {
		dm = 0
	}
	pri, err := strconv.Atoi(r.FormValue("priority"))
	if err != nil {
		pri = 0
	}

	var delivery_mode uint8 = uint8(dm)
	var priority uint8 = uint8(pri)

	timestamp, err := time.Parse(time.UnixDate, r.FormValue("timestamp"))
	if err != nil {
		timestamp = time.Now()
	}

	message := &amqp.Publishing{
		Body:            []byte(r.FormValue("body")),
		Headers:         amqp.Table{},
		ContentType:     r.FormValue("content_type"),
		ContentEncoding: r.FormValue("content_encoding"),
		DeliveryMode:    delivery_mode,
		Priority:        priority,
		CorrelationId:   r.FormValue("correlation_id"),
		ReplyTo:         r.FormValue("reply_to"),
		Expiration:      r.FormValue("expiration"),
		MessageId:       r.FormValue("message_id"),
		Timestamp:       timestamp,
		Type:            r.FormValue("type"),
		UserId:          r.FormValue("user_id"),
		AppId:           r.FormValue("app_id"),
	}

	return message, nil
}

/**
 * Retrieves the passed authentication user/password from the given Authentication header
 */
func getAuthentication(r *http.Request) ([]string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New(ERROR_PUBLISH_NO_AUTHORIZATION)
	}

	a := strings.IndexRune(authHeader, ' ')
	str, err := base64.StdEncoding.DecodeString(authHeader[a+1:])
	if err != nil {
		return nil, errors.New(ERROR_PUBLISH_NO_AUTHORIZATION)
	} else {
		return strings.Split(string(str), ":"), nil
	}
}
