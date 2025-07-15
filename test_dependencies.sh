#!/bin/bash

# Test script for the cerebro-mcp-server tools

echo "Testing cerebro-mcp-server tools..."

# Start the server in the background
export HTTP_MODE=true
./cerebro-mcp-server &
SERVER_PID=$!

# Wait for server to start
sleep 2

echo "Server started with PID: $SERVER_PID"

# Test the project details endpoint
echo "Testing details endpoint..."
response=$(curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"tool": "project_get_details", "arguments": {"project_permalink": "classic"}}' \
  -s -w "%{http_code}")

http_code="${response: -3}"
response_body="${response%???}"

if [ "$http_code" != "200" ]; then
    echo "ERROR: Details endpoint returned HTTP $http_code"
    echo "Response: $response_body"
    kill $SERVER_PID
    exit 1
fi

echo "✓ Details endpoint returned HTTP 200"
echo "$response_body" | jq .


# Test the dependencies endpoint
echo "Testing dependencies endpoint..."
response=$(curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"tool": "project_get_dependencies", "arguments": {"project_permalink": "classic"}}' \
  -s -w "%{http_code}")

http_code="${response: -3}"
response_body="${response%???}"

if [ "$http_code" != "200" ]; then
    echo "ERROR: Dependencies endpoint returned HTTP $http_code"
    echo "Response: $response_body"
    kill $SERVER_PID
    exit 1
fi

echo "✓ Dependencies endpoint returned HTTP 200"
echo "$response_body" | jq .

# Kill the server
kill $SERVER_PID

echo "✓ All tests passed successfully!"
