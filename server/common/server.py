import socket
import logging
import signal

from common.client_connection import ClientConnection
from common.bet_serializer import BetSerializer
from common.utils import has_won, load_bets, store_bets


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self.connections: list[ClientConnection] = []
        self.agencies_ready = set()
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
                self.connections.append(client_conn)
                self.__handle_client_connection(client_conn)
            except OSError:
                if not self._is_running:
                    break
                else:
                    raise

            if len(self.agencies_ready) >= 5:
                self._start_draw()

    def __handle_client_connection(self, client_conn: ClientConnection):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            client_conn.id = client_conn.read_identity_v1()
            logging.debug(f'action: get_agency_id | result: success | ip: {client_conn.addr[0]} | value: {client_conn.id}')
        except (OSError, RuntimeError) as e:
            logging.error(f"action: get_agency_id | result: fail | error: {e}")
            self._disconnect_client(client_conn)
            return
            
        while(True):
            try:
                batch_size = client_conn.read_bet_batch_v1()
                logging.debug(f'action: receive_batch_size | result: success | ip: {client_conn.addr[0]} | value: {batch_size}')
            except (OSError, RuntimeError) as e:
                logging.error(f"action: receive_batch_size | result: fail | error: {e}")
                self._disconnect_client(client_conn)
                return
            
            # A zero size batch size means Bet transmission ended
            if batch_size == 0:
                assert(client_conn.id > 0)
                self.agencies_ready.add(client_conn.id)
                client_conn.send_ack()
                logging.debug(f'action: send_ack | result: success | ip: {client_conn.addr[0]}')
                return

            try:
                bets = []
                for _ in range(batch_size):
                    raw_bet = client_conn.read_bet_v1()
                    bet = BetSerializer.deserialize(client_conn.id, raw_bet)
                    bets.append(bet)
                logging.info(f'action: apuesta_recibida | result: success | cantidad: {batch_size}')
                store_bets(bets)
                client_conn.send_ack()
                logging.debug(f'action: send_ack | result: success | ip: {client_conn.addr[0]}')
            except (OSError, RuntimeError) as e:
                logging.error(f"action: receive_bet | result: fail | error: {e}")
                logging.error(f"action: apuesta_recibida | result: fail | cantidad: {batch_size}")
                self._disconnect_client(client_conn)
                return

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
        
    def _disconnect_client(self, client_conn: ClientConnection):
        client_conn.disconnect()
        logging.info(f'action: client_socket_close | result: success | id: {client_conn.id}')
        self.connections.remove(client_conn)
    
    def _start_draw(self):
        logging.info('action: sorteo | result: success')

        # Collect winners by agency id
        agency_winners: dict[int, list] = {}
        for bet in load_bets():
            if has_won(bet):
                agency_winners.setdefault(bet.agency, []).append(int(bet.document))
        
        for conn in self.connections:
            if conn.id <= 0:
                continue
            conn.send_winners_v1(agency_winners.get(conn.id, []))
        
    def _on_stop_signal(self, _signum: int, _frame: str) -> None:
        logging.info('action: shutdown signal received | result: success')
        self._is_running = False
        for conn in self.connections:
            self._disconnect_client(conn)
        self._server_socket.close()
        logging.info('action: server socket close | result: success')
