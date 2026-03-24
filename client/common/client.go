package common

import (
	"fmt"
	"os"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

const DATASET_PATH = "./.data/agency-%v.csv"

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	MaxBatchItems int
}

// Client Entity that encapsulates how
type Client struct {
	config      ClientConfig
	socket      *Socket
	stopSignal  chan struct{}
	DatasetFile *os.File
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:     config,
		stopSignal: make(chan struct{}),
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	sock, err := ConnectTCP(c.config.ServerAddress)
	if err != nil {
		log.Criticalf("action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID, err,
		)
	}
	c.socket = sock
	log.Debugf("action: connect | result: success | client_id: %v", c.config.ID)
	return nil
}

func (c *Client) closeClientSocket() {
	if c.socket.conn != nil {
		c.socket.Disconnect()
		log.Debugf("action: connection_close | result: success | client_id: %v", c.config.ID)
	}
}

func (c *Client) DoClientLoop(batchBuilder *BatchBuilder) {
	for !batchBuilder.ReachedEOF {
		select {
		case <-c.stopSignal:
			log.Infof("action: loop_exit_shutdown | result: success | client_id: %v", c.config.ID)
			return
		default:
		}

		c.createClientSocket()

		rawBatch, count, err := batchBuilder.BuildNext()
		if err != nil {
			log.Errorf("action: build_batch | result: fail | client_id: %v | error: %v",
				c.config.ID, err,
			)
			return
		}

		if err := c.socket.SendBytes(rawBatch); err != nil {
			log.Errorf("action: send_batch | result: fail | client_id: %v | error: %v",
				c.config.ID, err,
			)
			return
		}

		response, err := c.socket.ReadLine()
		if err != nil {
			log.Errorf("action: receive_ack | result: fail | client_id: %v | error: %v",
				c.config.ID, err,
			)
			return
		}
		if response != "ACK\n" {
			log.Errorf("action: receive_ack | result: fail | client_id: %v | error: Unexpected response: %s",
				c.config.ID, response,
			)
			return
		}

		log.Infof("action: apuesta_enviada | result: success | cantidad: %v", count)

		c.closeClientSocket()
	}
}

func (c *Client) Start() {
	f, err := os.Open(fmt.Sprintf(DATASET_PATH, c.config.ID))
	if err != nil {
		log.Critical(err)
		return
	}
	c.DatasetFile = f
	batchBuilder := NewBatchBuilder(c.DatasetFile, c.config.MaxBatchItems)
	c.DoClientLoop(batchBuilder)

	c.Shutdown()
}

func (c *Client) Shutdown() {
	c.closeClientSocket()

	if c.DatasetFile != nil {
		c.DatasetFile.Close()
		c.DatasetFile = nil
		log.Debugf("action: dataset_file_close | result: success | client_id: %v", c.config.ID)
	}
}

func (c *Client) OnStopSignal() {
	log.Infof("action: shutdown signal received | result: success | client_id: %v", c.config.ID)

	close(c.stopSignal)
	c.closeClientSocket()
}
