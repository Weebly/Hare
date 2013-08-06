package main

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/streadway/amqp"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	HARE_API_PORT_DEFAULT        = "8080"
	RABBITMQ_HOST                = ""
	RABBITMQ_PORT                = "5672"
	RABBITMQ_VHOST               = ""
	RABBITMQ_MANAGEMENT_PORT     = "55672"
	RABBITMQ_USERNAME            = ""
	RABBITMQ_PASSWORD            = ""
	RABBITMQ_MANAGEMENT_USERNAME = ""
	RABBITMQ_MANAGEMENT_PASSWORD = ""
	RABBITMQ_URI                 = "amqp://" + RABBITMQ_USERNAME + ":" + RABBITMQ_PASSWORD + "@" + RABBITMQ_HOST + ":" + RABBITMQ_PORT + "/" + RABBITMQ_VHOST
	RABBITMQ_MANAGEMENT_URI      = "http://" + RABBITMQ_HOST + ":" + RABBITMQ_MANAGEMENT_PORT + "/api/definitions/"

	/**
	 * Statistics related strings
	 */
	RUNTIME_PUBLISHES_RECEIVED = "runtime_publishes_received"
	RUNTIME_PUBLISHES_SUCCESS  = "runtime_publishes_success"
	RUNTIME_PUBLISHES_FAILURE  = "runtime_publishes_failure"
)

/**
 * Maintain the connection to RabbitMQ for publishing
 */
var Connection, Err = amqp.Dial(RABBITMQ_URI)
var Channel, Cerr = Connection.Channel()

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

func main() {
	log.Print("Hare started.\n")

	/**
	 * Are we alive?
	 */
	if Err != nil {
		log.Fatalf("Unable to connect to RabbitMQ via AMQP - %s", Err)
	}
	if Cerr != nil {
		log.Fatalf("Couldn't create channel - %s", Cerr)
	}
	defer Connection.Close()

	log.Print("Connected to RabbitMQ.\n")

	/**
	 * Retrieve data from the RabbitMQ host to maintain definitions
	 */
	if err := getDefinitions(); err != nil {
		log.Fatal(err)
	}

	log.Print("Definitions retrieved.\n")

	/**
	 * We're ready to start our webservice broker.
	 */
	if err := startWebService(); err != nil {
		log.Fatalf("Failed to start webservice - %s\n", err)
	}
}

/**
 * Retrieves definitions from the RabbitMQ Administration API endpoint
 *
 * @return void
 */
func getDefinitions() error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", RABBITMQ_MANAGEMENT_URI, nil)
	req.SetBasicAuth(RABBITMQ_MANAGEMENT_USERNAME, RABBITMQ_MANAGEMENT_PASSWORD)

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf("Failed to perform HTTP request to %s", RABBITMQ_HOST)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to read from management API - %s\n", err)
	}

	json, err := simplejson.NewJson(body)

	c := 0
	for c < len(json.Get("exchanges").MustArray()) {
		c = c + 1

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
 * @return nil
 */
func startWebService() error {
	log.Printf("API service available on port %s.", HARE_API_PORT_DEFAULT)
	http.HandleFunc("/", apiRequestHandler)
	err := http.ListenAndServe(":"+HARE_API_PORT_DEFAULT, nil)
	if err != nil {
		return fmt.Errorf("%s", err)
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
	if err := getDefinitions(); err != nil {
		fmt.Fprint(w, err)
	}

	fmt.Fprintf(w, "Definitions retrieved.\n\nNew Object:\n%s", Definitions)
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
	log.Printf("Exit command received by %s", r.RemoteAddr)
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
		 * We don't have this exchange URI (get out)
		 */
		Statistics[RUNTIME_PUBLISHES_FAILURE] = Statistics[RUNTIME_PUBLISHES_FAILURE] + 1
		http.NotFound(w, r)
		return
	}

	if len(r.FormValue("body")) < 1 {
		Statistics[RUNTIME_PUBLISHES_FAILURE] = Statistics[RUNTIME_PUBLISHES_FAILURE] + 1
		http.Error(w, "POST is required for publish.", 400)
		return
	}

	e := Definitions[path[2]]
	var routingKey string

	if len(path) < 4 {
		/**
		 * No routing key was defined.
		 */
		routingKey = ""
	} else {
		routingKey = path[3]
	}

	if err := publish(*e, routingKey, r.FormValue("body")); err != nil {
		Statistics[RUNTIME_PUBLISHES_FAILURE] = Statistics[RUNTIME_PUBLISHES_FAILURE] + 1
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
	fmt.Fprintf(w, "Delivery Statistics:\n%s\n\nExchange Definition Object:\n%d", Statistics, Definitions)
}

/**
 * Publishes to RabbitMQ with the given connection, exchange, type, etc..
 *
 * @param e ExchangeDefinition
 * @param string routingKey
 * @param string body
 *
 * @return nil|error
 */
func publish(e ExchangeDefinition, routingKey string, body string) error {
	if Cerr = Channel.ExchangeDeclare(
		e.Name,
		e.Type,
		e.Durability,
		e.AutoDelete,
		false,
		false,
		nil,
	); Cerr != nil {
		return fmt.Errorf("Declaration failure. %s\n", Cerr)
	}

	if Cerr = Channel.Publish(
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
	); Cerr != nil {
		return fmt.Errorf("Publishing failure. %s\n", Cerr)
	}

	return nil
}
