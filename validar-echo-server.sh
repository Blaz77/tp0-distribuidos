#!/bin/bash

SERVER_CONTAINER_NAME="server"
SERVER_PORT=12345
NETWORK_NAME="tp0_testing_net"

TEST_STRING="testing_the_server"

# Using alpine because it's lightweight but has netcat
RESULT=$(docker run --rm --network $NETWORK_NAME alpine sh -c "echo $TEST_STRING | nc $SERVER_CONTAINER_NAME $SERVER_PORT")

if [ "$RESULT" = "$TEST_STRING" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi