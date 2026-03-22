import socket

class ClientConnection:
    def __init__(self, socket: socket.socket, addr):
        self.socket = socket
        self.addr = addr
        self._read_buff = b''

    def ReadLine(self):
        while True:
            packet = self.socket.recv(4096)
            if not packet:
                raise Exception("Connection closed")

            self._read_buff += packet
            if b'\n' in packet:
                break

        line, self._read_buff = self._read_buff.split(b'\n', 1)
        return line.rstrip().decode('utf-8')
    
    def SendString(self, data: str):
        self.socket.sendall(data.encode('utf-8'))

    def disconnect(self):
        self.socket.close()
        self.addr = None