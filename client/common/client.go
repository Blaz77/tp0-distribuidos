package common

import (
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	ClientBet     *Bet
}

// Client Entity that encapsulates how
type Client struct {
	config     ClientConfig
	socket     *Socket
	stopSignal chan struct{}
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
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.socket = sock
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {

		select {
		case <-c.stopSignal:
			log.Infof("action: loop_exit_shutdown | result: success | client_id: %v", c.config.ID)
			return
		default:
		}

		// Create the connection the server in every loop iteration. Send an
		c.createClientSocket()

		rawBet, err := c.config.ClientBet.Serialize()
		if err != nil {
			log.Errorf("action: serialize_bet | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		if err := c.socket.SendBytes(rawBet); err != nil {
			log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}
		// TODO Implement ack
		//response, err := c.socket.ReadLine()
		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
			c.config.ClientBet.Document,
			c.config.ClientBet.Number,
		)

		c.socket.Disconnect()

		// Wait a time between sending one message and the next one
		// Can be cancelled through stop_signal, without busy-waiting
		select {
		case <-c.stopSignal:
			log.Infof("action: loop_exit_shutdown | result: success | client_id: %v", c.config.ID)
			return
		case <-time.After(c.config.LoopPeriod):
		}
	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func (c *Client) OnStopSignal() {
	log.Infof("action: shutdown signal received | result: success | client_id: %v", c.config.ID)

	close(c.stopSignal)

	if c.socket.conn != nil {
		c.socket.Disconnect()
		log.Infof("action: connection close | result: success | client_id: %v", c.config.ID)
	}
}
