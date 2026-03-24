package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"

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
func (c *Client) CreateClientSocket() error {
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

func (c *Client) CloseClientSocket() {
	if c.socket.conn != nil {
		c.socket.Disconnect()
		log.Debugf("action: connection_close | result: success | client_id: %v", c.config.ID)
	}
}

func (c *Client) IdentifyAgency() {
	const identifyHeaderId = "AI\x00\x01"
	agencyId, _ := strconv.ParseUint(c.config.ID, 10, 32)
	buf := make([]byte, 8)
	copy(buf[:4], identifyHeaderId)
	binary.BigEndian.PutUint32(buf[4:], uint32(agencyId))

	if err := c.socket.SendBytes(buf); err != nil {
		log.Criticalf("action: identify_agency | result: fail | client_id: %v | error: %v",
			c.config.ID, err,
		)
	}
	log.Info("action: identify_agency | result: success")
}

func (c *Client) SendBetsLoop(batchBuilder *BatchBuilder) {
	for !batchBuilder.ReachedEOF {
		select {
		case <-c.stopSignal:
			log.Infof("action: loop_exit_shutdown | result: success | client_id: %v", c.config.ID)
			return
		default:
		}

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
			log.Errorf("action: receive_ack | result: fail | client_id: %v | error: Invalid response: %s",
				c.config.ID, response,
			)
			return
		}

		log.Infof("action: apuesta_enviada | result: success | cantidad: %v", count)
	}

	// Notify server that no more bets will be sent
	endMsg := batchBuilder.BuildEnd()
	if err := c.socket.SendBytes(endMsg); err != nil {
		log.Errorf("action: send_end | result: fail | client_id: %v | error: %v",
			c.config.ID, err,
		)
		return
	}
	response, err := c.socket.ReadLine()
	if err != nil {
		log.Errorf("action: receive_end_ack | result: fail | client_id: %v | error: %v",
			c.config.ID, err,
		)
		return
	}
	if response != "ACK\n" {
		log.Errorf("action: receive_end_ack | result: fail | client_id: %v | error: Invalid response: %s",
			c.config.ID, response,
		)
		return
	}
	log.Debug("action: receive_end_ack | result: success")
}

func (c *Client) ReceiveWinners() {
	const IntSize = 4
	const WinnersHeaderId = "AW\x00\x01" // Agency Winners v1
	const WinnersHeaderSize = len(WinnersHeaderId) + IntSize
	header, err := c.socket.ReadBytes(uint32(WinnersHeaderSize))
	if err != nil {
		log.Errorf("action: consulta_ganadores | result: fail | error: %v", err)
		return
	}

	if !bytes.Equal(header[:4], []byte(WinnersHeaderId)) {
		log.Errorf("action: consulta_ganadores | result: fail | error: Invalid header %v", header[:4])
		return
	}

	winnersNum := binary.BigEndian.Uint32(header[4:8])
	_, err = c.socket.ReadBytes(winnersNum * IntSize)
	if err != nil {
		log.Errorf("action: consulta_ganadores | result: fail | error: %v", err)
		return
	}

	// Payload not used, as we only need the amount of winners
	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d", winnersNum)
}

func (c *Client) Start() {
	f, err := os.Open(fmt.Sprintf(DATASET_PATH, c.config.ID))
	if err != nil {
		log.Critical(err)
		return
	}
	c.DatasetFile = f
	batchBuilder := NewBatchBuilder(c.DatasetFile, c.config.MaxBatchItems)

	c.CreateClientSocket()
	c.IdentifyAgency()
	c.SendBetsLoop(batchBuilder)
	c.ReceiveWinners()

	c.Shutdown()
}

func (c *Client) Shutdown() {
	c.CloseClientSocket()

	if c.DatasetFile != nil {
		c.DatasetFile.Close()
		c.DatasetFile = nil
		log.Debugf("action: dataset_file_close | result: success | client_id: %v", c.config.ID)
	}
}

func (c *Client) OnStopSignal() {
	log.Infof("action: shutdown signal received | result: success | client_id: %v", c.config.ID)

	close(c.stopSignal)
	c.CloseClientSocket()
}
