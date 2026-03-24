import socket
import logging
import signal

from common.client_connection import ClientConnection
from common.bet_serializer import BetSerializer
from common.utils import store_bets


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._is_running = True

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        signal.signal(signal.SIGTERM, self._on_stop_signal)
        signal.signal(signal.SIGINT, self._on_stop_signal)
        
        while self._is_running:
            try:
                client_conn = self.__accept_new_connection()
                self.__handle_client_connection(client_conn)
            except OSError:
                if not self._is_running:
                    break
                else:
                    raise

    def __handle_client_connection(self, client_conn: ClientConnection):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            batch_size = client_conn.read_bet_batch_v1()
            logging.debug(f'action: receive_batch_size | result: success | ip: {client_conn.addr[0]} | value: {batch_size}')
        except (OSError, RuntimeError) as e:
            logging.error(f"action: receive_batch_size | result: fail | error: {e}")
            client_conn.disconnect()
            logging.info('action: client socket close | result: success')
            return

        try:
            bets = []
            for _ in range(batch_size):
                raw_bet = client_conn.read_bet_v1()
                bet = BetSerializer.deserialize(raw_bet)
                bets.append(bet)
            logging.info(f'action: apuesta_recibida | result: success | cantidad: {batch_size}')
            store_bets(bets)
            client_conn.send_ack()
            logging.debug(f'action: send_ack | result: success | ip: {client_conn.addr[0]}')
        except (OSError, RuntimeError) as e:
            logging.error(f"action: receive_bet | result: fail | error: {e}")
            logging.error(f"action: apuesta_recibida | result: fail | cantidad: {batch_size}")
        finally:
            client_conn.disconnect()
            logging.info('action: client socket close | result: success')

    def __accept_new_connection(self) -> ClientConnection:
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        try:
            c, addr = self._server_socket.accept()
            logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
            return ClientConnection(c, addr)
        except OSError as e:
            if self._is_running:
                logging.error(f"action: accept_connections | result: fail | error: {e}")
            else:
                logging.info(f"action: accept_connections cancelled | result: success")
            raise
        
    
    def _on_stop_signal(self, _signum: int, _frame: str) -> None:
        logging.info('action: shutdown signal received | result: success')
        self._is_running = False
        self._server_socket.close()
        logging.info('action: server socket close | result: success')
