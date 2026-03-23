import socket

class ClientConnection:
    def __init__(self, socket: socket.socket, addr):
        self.socket = socket
        self.addr = addr
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

    def read_bet_v1(self):
        BET_HEADER_ID = b'AB\0\1'
        INT_SIZE = 4
        HEADER_SIZE = len(BET_HEADER_ID) + INT_SIZE
        while True:
            # Receive until we have enough data to work with
            while len(self._read_buff) < HEADER_SIZE:
                if not self._recv_more():
                    raise RuntimeError("Unable to read the packet header")
            
            # Look for the expected header and discard garbage
            header_pos = self._read_buff.find(BET_HEADER_ID)
            if header_pos == -1:
                # Keep a possible partial header
                self._read_buff = self._read_buff[-len(BET_HEADER_ID):]
                continue
            if header_pos > 0:
                self._read_buff = self._read_buff[header_pos:]

            # Ensure we have the Bet packet size
            while len(self._read_buff) < HEADER_SIZE:
                if not self._recv_more():
                    raise Exception("Unable to read the payload length")
                
            payload_size = int.from_bytes(self._read_buff[len(BET_HEADER_ID):HEADER_SIZE], "big")
            
            # Ensure we have the entire expected packet
            total_size = HEADER_SIZE + payload_size
            while len(self._read_buff) < total_size:
                if not self._recv_more():
                    raise Exception("Unable to finish reading the packet payload")
            
            # Save payload and remove from buffer
            payload = self._read_buff[HEADER_SIZE:total_size]
            self._read_buff = self._read_buff[total_size:]
            return payload


    def disconnect(self):
        self.socket.close()
        self.addr = None