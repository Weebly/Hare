package main

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/streadway/amqp"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
)

const (
	HARE_API_PORT_DEFAULT                  = "8080"
	RABBITMQ_HOST                          = "5672"
	RABBITMQ_PORT                          = ""
	RABBITMQ_VHOST                         = ""
	RABBITMQ_MANAGEMENT_PORT               = "55672"
	RABBITMQ_USERNAME                      = ""
	RABBITMQ_PASSWORD                      = ""
	RABBITMQ_MANAGEMENT_USERNAME           = ""
	RABBITMQ_MANAGEMENT_PASSWORD           = ""
	RABBITMQ_URI                           = "amqp://" + RABBITMQ_USERNAME + ":" + RABBITMQ_PASSWORD + "@" + RABBITMQ_HOST + ":" + RABBITMQ_PORT + "/" + RABBITMQ_VHOST
	RABBITMQ_MANAGEMENT_URI                = "http://" + RABBITMQ_HOST + ":" + RABBITMQ_MANAGEMENT_PORT + "/api/definitions/"
	HARE_RABBITMQ_CONNECTION_RETRY_MAXIMUM = 5

	/**
	 * Statistics related strings
	 */
	RUNTIME_PUBLISHES_RECEIVED          = "runtime_publishes_received"
	RUNTIME_PUBLISHES_SUCCESS           = "runtime_publishes_success"
	RUNTIME_PUBLISHES_FAILURE           = "runtime_publishes_failure"
	RUNTIME_PUBLISHES_FAILURE_400       = "runtime_publishes_failure_bad_request"
	RUNTIME_PUBLISHES_FAILURE_404       = "runtime_publishes_failure_not_found"
	RUNTIME_PUBLISHES_FAILURE_500       = "runtime_publishes_failure_publish_error"
	RUNTIME_RABBITMQ_CONNECTION_FAILURE = "runtime_rabbitmq_connection_failure"
)

/**
 * Definitions contains the details of the exchanges on the given host, vhost
 */
type ExchangeDefinition struct {
	Name       string
	Type       string
	Durability bool
	AutoDelete bool
}

var Definitions = make(map[string]*ExchangeDefinition)

/**
 * Statistics Collector
 */
var Statistics = make(map[string]int)

/**
 * AMQP connection and channel
 */
var Connection *amqp.Connection
var Channel *amqp.Channel

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	Connection = getConnection()
	Channel = getChannel()

	log.Printf("Hare started, connected to remote RabbitMQ host at %s.", RABBITMQ_HOST)

	if err := populateDefinitions(); err != nil {
		log.Fatalf("Failed to retrieve definitions, %s.", err)
	}

	log.Print("Definitions retrieved.")

	/**
	 * We're ready to start our web API.
	 */
	if err := startWebService(); err != nil {
		log.Fatalf("Failed to start webservice, %s.", err)
	}
}

/**
 * Retrieves a connection from the configured AMQP endpoint
 *
 * @return amqp.Connection
 */
func getConnection() *amqp.Connection {
	connection, err := amqp.Dial(RABBITMQ_URI)
	if err != nil {
		handleAmqpError(err)
	}

	return connection
}

/**
 * Retrieves a channel from the current connected AMQP host
 *
 * @return amqp.Channel
 */
func getChannel() *amqp.Channel {
	channel, err := Connection.Channel()
	if err != nil {
		handleAmqpError(err)
	}

	return channel
}

func handleAmqpError(err error) {
	switch err {
	case amqp.ErrClosed:
		log.Printf("AMQP connection error, %v.", err)
		/**
		 * Don't want to loop through this for too long.
		 */
		if Statistics[RUNTIME_RABBITMQ_CONNECTION_FAILURE] > HARE_RABBITMQ_CONNECTION_RETRY_MAXIMUM {
			log.Fatal("Maximum connection retry count has been reached. Exiting.")
		}
		Connection = getConnection()
		Channel = getChannel()

		Statistics[RUNTIME_RABBITMQ_CONNECTION_FAILURE] = Statistics[RUNTIME_RABBITMQ_CONNECTION_FAILURE] + 1
		log.Print("Reconnected to remote host.")
	case amqp.ErrSASL, amqp.ErrCredentials, amqp.ErrVhost:
		/**
		 * These errors should be considered pretty much fatal.
		 */
		log.Fatalf("Fatal AMQP connection error, %s.", err.Error())
	default:
		log.Fatalf("Unexpected AMQP error, %s.", err)
	}
}

/**
 * Retrieves definitions from the RabbitMQ Administration API endpoint
 *
 * @return void
 */
func populateDefinitions() error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", RABBITMQ_MANAGEMENT_URI, nil)
	req.SetBasicAuth(RABBITMQ_MANAGEMENT_USERNAME, RABBITMQ_MANAGEMENT_PASSWORD)

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	json, err := simplejson.NewJson(body)

	for c := 0; c < len(json.Get("exchanges").MustArray()); c++ {
		if len(json.Get("exchanges").GetIndex(c).Get("name").MustString()) > 0 {
			durable, _ := json.Get("exchanges").GetIndex(c).Get("durable").Bool()
			delete, _ := json.Get("exchanges").GetIndex(c).Get("auto_delete").Bool()

			d := &ExchangeDefinition{
				json.Get("exchanges").GetIndex(c).Get("name").MustString(),
				json.Get("exchanges").GetIndex(c).Get("type").MustString(),
				durable,
				delete,
			}

			Definitions[json.Get("exchanges").GetIndex(c).Get("name").MustString()] = d
		}
	}
	return nil
}

/**
 * Starts the API endpoint webservice
 *
 * @return error
 */
func startWebService() error {
	log.Printf("API service available on port %s.", HARE_API_PORT_DEFAULT)
	http.HandleFunc("/", apiRequestHandler)
	err := http.ListenAndServe(":"+HARE_API_PORT_DEFAULT, nil)
	if err != nil {
		return err
	}
	return nil
}

/**
 * Routes and serves API requests based on the path
 *
 * @param w ResponseWriter
 * @param r Request
 *
 * @return void
 */
func apiRequestHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")

	switch path[1] {
	case "publish":
		apiRequestPublish(w, r, path)
	case "stats":
		apiRequestStats(w, r)
	case "alive":
		fmt.Fprint(w, "Alive")
	case "reload":
		apiRequestReload(w, r)
	case "exit":
		apiRequestExit(w, r)
	default:
		http.NotFound(w, r)
	}
}

/**
 * Prepares and dispatches a response via ResponseWriter for a reload API request
 *
 * @param w ResponseWriter
 * @param r Request
 *
 * @return void
 */
func apiRequestReload(w http.ResponseWriter, r *http.Request) {
	Definitions = make(map[string]*ExchangeDefinition)

	if err := populateDefinitions(); err != nil {
		fmt.Fprint(w, err)
	}

	fmt.Fprintf(w, "Definitions retrieved.\nNew Object:\n%v", Definitions)
}

/**
 * Prepares and dispatches a response via ResponseWriter for an exit request
 *
 * @param w ResponseWriter
 * @param r Request
 *
 * @return void
 */
func apiRequestExit(w http.ResponseWriter, r *http.Request) {
	log.Printf("Exit command received by %s.", r.RemoteAddr)
	os.Exit(2)
}

/**
 * Prepares and dispatches a response via ResponseWriter for a publish API request
 * Also parses the URI and POST for a publish event to RabbitMQ
 *
 * @param w ResponseWriter
 * @param r Request
 *
 * @return void
 */
func apiRequestPublish(w http.ResponseWriter, r *http.Request, path []string) {
	Statistics[RUNTIME_PUBLISHES_RECEIVED] = Statistics[RUNTIME_PUBLISHES_RECEIVED] + 1

	/**
	 * Retrieve the definition of the exchange from the URI
	 */
	if Definitions[path[2]] == nil {
		/**
		 * We don't have this exchange URI
		 */
		Statistics[RUNTIME_PUBLISHES_FAILURE] = Statistics[RUNTIME_PUBLISHES_FAILURE] + 1
		Statistics[RUNTIME_PUBLISHES_FAILURE_404] = Statistics[RUNTIME_PUBLISHES_FAILURE_404] + 1
		http.NotFound(w, r)
		return
	}

	if len(r.FormValue("body")) < 1 {
		/**
		 * No body was associated with the request
		 */
		Statistics[RUNTIME_PUBLISHES_FAILURE] = Statistics[RUNTIME_PUBLISHES_FAILURE] + 1
		Statistics[RUNTIME_PUBLISHES_FAILURE_400] = Statistics[RUNTIME_PUBLISHES_FAILURE_400] + 1
		http.Error(w, "POST is required for publish.", 400)
		return
	}

	e := Definitions[path[2]]
	var routingKey string

	if len(path) < 4 {
		/**
		 * No routing key was defined, which is fine -- use emptystring
		 */
		routingKey = ""
	} else {
		routingKey = path[3]
	}

	if err := publish(*e, routingKey, r.FormValue("body")); err != nil {
		Statistics[RUNTIME_PUBLISHES_FAILURE] = Statistics[RUNTIME_PUBLISHES_FAILURE] + 1
		Statistics[RUNTIME_PUBLISHES_FAILURE_500] = Statistics[RUNTIME_PUBLISHES_FAILURE_500] + 1
		http.Error(w, "Error Publishing", 500)
		return
	}

	Statistics[RUNTIME_PUBLISHES_SUCCESS] = Statistics[RUNTIME_PUBLISHES_SUCCESS] + 1
	return
}

/**
 * Prepares and dispatches a response via ResponseWriter for a stats API request
 *
 * @param w ResponseWriter
 * @param r Request
 *
 * @return void
 */
func apiRequestStats(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Delivery Statistics:\n%v\n\nExchange Definition Object:\n%s", Statistics, Definitions)
}

/**
 * Publishes the given message to RabbitMQ with the current connection, given ExchangeDefinition
 *
 * @param e ExchangeDefinition
 * @param string routingKey
 * @param string body
 *
 * @return error
 */
func publish(e ExchangeDefinition, routingKey string, body string) error {
	if err := Channel.ExchangeDeclare(
		e.Name,
		e.Type,
		e.Durability,
		e.AutoDelete,
		false,
		false,
		nil,
	); err != nil {
		handleAmqpError(err)
		return err
	}

	if err := Channel.Publish(
		e.Name,
		routingKey,
		false,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            []byte(body),
			DeliveryMode:    amqp.Transient,
		},
	); err != nil {
		handleAmqpError(err)
		return err
	}

	return nil
}
