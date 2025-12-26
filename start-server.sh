#!/bin/bash
# Start the Backend (Server)
echo "Starting Backend..."
cd server
go run cmd/server/main.go
