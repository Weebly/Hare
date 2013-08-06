Hare
====
Hare is a HTTP to RabbitMQ publishing gateway, written in Go.
Hare solves some of the issues related to short-lived processes (such as PHP) publishing to RabbitMQ, namely overwhelming RabbitMQ with connections.

Hare connects to your RabbitMQ administration API to retrieve exchange definitions, preventing the need to maintain somewhat irrelevant implementation details in your publishing application logic.


Requirements
------------
 * github.com/bitly/go-simplejson
 * github.com/streadway/amqp

Usage
-----
After adding your details to the configuration constants, start hare with `go run hare`.

API
---
To access Hare, use the following API endpoints:
 * http://harehost:port/alive - Responds with an HTTP 200 and a body of `Alive`, for use in a health check.
 * http://harehost:port/stats - Responds with delivery statistics, and an overview of the current definitions object.
 * http://harehost:port/reload - Reloads the definition object by performing another request to the RabbitMQ administration API. Responds with an overview of the new definitions object.
 * http://harehost:port/exit - Empty response. Stops the Hare gateway, logging the initiator of the request to stdout.

 * http://harehost:port/publish/{exchange}/{routingKey}
 ** Publish requires a POST field of `body``
