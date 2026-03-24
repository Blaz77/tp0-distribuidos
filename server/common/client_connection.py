import socket

INT_SIZE = 4

class ClientConnection:    
    def __init__(self, socket: socket.socket, addr):
        self.socket = socket
        self.addr = addr
        self.id = -1
        self._read_buff = b''

    def _recv_more(self) -> bytes:
        packet = self.socket.recv(4096)
        self._read_buff += packet
        return packet

    def read_line(self):
        while True:
            packet = self._recv_more(self)
            if not packet:
                raise Exception("Connection closed")
            if b'\n' in packet:
                break

        line, self._read_buff = self._read_buff.split(b'\n', 1)
        return line.rstrip().decode('utf-8')
    
    def send_string(self, data: str):
        self.socket.sendall(data.encode('utf-8'))

    def send_ack(self):
        data = "ACK\n"
        self.send_string(data)

    def send_winners_v1(self, winners: list[int]):
        # Header
        WINNERS_HEADER_ID = b'AW\0\1'
        buffer = bytearray()
        buffer += WINNERS_HEADER_ID
        buffer += len(winners).to_bytes(4, byteorder="big", signed=False)

        # Payload
        for winner in winners:
            buffer += winner.to_bytes(4, byteorder="big", signed=False)

        self.socket.sendall(buffer)

    def _read_header(self, header_id: bytes) -> int:
        HEADER_SIZE = len(header_id) + INT_SIZE
        while True:
            # Receive until we have enough data to work with
            while len(self._read_buff) < HEADER_SIZE:
                if not self._recv_more():
                    raise RuntimeError("Unable to read the packet header")
            
            # Look for the expected header and discard garbage
            header_pos = self._read_buff.find(header_id)
            if header_pos == -1:
                # Keep a possible partial header
                self._read_buff = self._read_buff[-len(header_id):]
                continue
            if header_pos > 0:
                self._read_buff = self._read_buff[header_pos:]

            # Ensure we have the size field
            while len(self._read_buff) < HEADER_SIZE:
                if not self._recv_more():
                    raise RuntimeError("Unable to read the size field")
                
            size = int.from_bytes(self._read_buff[len(header_id):HEADER_SIZE], "big")
            return size
        
    def read_identity_v1(self):
        BET_HEADER_ID = b'AI\0\1'
        HEADER_SIZE = len(BET_HEADER_ID) + INT_SIZE
        agency_id = self._read_header(BET_HEADER_ID)
        self._read_buff = self._read_buff[HEADER_SIZE:]
        return agency_id

    def read_bet_batch_v1(self):
        BATCH_HEADER_ID = b'BB\0\1'
        HEADER_SIZE = len(BATCH_HEADER_ID) + INT_SIZE
        batch_size = self._read_header(BATCH_HEADER_ID)
        self._read_buff = self._read_buff[HEADER_SIZE:]
        return batch_size

    def read_bet_v1(self):
        BET_HEADER_ID = b'AB\0\1'
        HEADER_SIZE = len(BET_HEADER_ID) + INT_SIZE
        payload_size = self._read_header(BET_HEADER_ID)
        total_size = HEADER_SIZE + payload_size
        while True:
            # Ensure we have the entire expected packet
            while len(self._read_buff) < total_size:
                if not self._recv_more():
                    raise RuntimeError("Unable to finish reading the packet payload")
            
            # Save payload and remove from buffer
            payload = self._read_buff[HEADER_SIZE:total_size]
            self._read_buff = self._read_buff[total_size:]
            return payload


    def disconnect(self):
        self.socket.close()
        self.addr = None