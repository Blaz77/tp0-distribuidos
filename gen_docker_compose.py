import sys

class DockerComposeGenerator:
    _SERVER_SECTION = """  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net
"""

    _CLIENT_TEMPLATE = """
  [[name]]:
    container_name: [[name]]
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=[[id]]
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server
"""

    _NETWORKS_SECTION = """
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
"""

    @staticmethod
    def _generate_client(id) -> str:
        client_section = DockerComposeGenerator._CLIENT_TEMPLATE
        client_section = client_section.replace("[[id]]", str(id))
        client_section = client_section.replace("[[name]]", f"client{id}")
        return client_section
    
    @staticmethod
    def _generate_content(num_clients):
        content = "name: tp0\nservices:\n"
        content += DockerComposeGenerator._SERVER_SECTION
        for i in range(num_clients):
            content += DockerComposeGenerator._generate_client(i+1)
        content += DockerComposeGenerator._NETWORKS_SECTION
        return content

    def __init__(self, num_clients):
        self.content = DockerComposeGenerator._generate_content(num_clients)
    
    def __repr__(self):
        return self.content
    
    def save_to_disk(self, filename):
        with open(filename, "w") as f:
            f.write(self.content)

def main():
    if len(sys.argv) < 3:
        print("Uso: python3 gen_docker_compose.py <filename> <num_clients>")
        sys.exit(1)

    filename = sys.argv[1]
    num_clients = int(sys.argv[2])
    generator = DockerComposeGenerator(num_clients)
    generator.save_to_disk(filename)

main()