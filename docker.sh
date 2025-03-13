#!/bin/bash

# Make script executable with: chmod +x docker.sh

case "$1" in
  build)
    echo "Building Docker image..."
    sudo docker compose build
    ;;
  up)
    echo "Starting containers in detached mode..."
    sudo docker compose up -d
    ;;
  down)
    echo "Stopping containers..."
    sudo docker compose down
    ;;
  logs)
    echo "Showing logs..."
    sudo docker compose logs -f
    ;;
  restart)
    echo "Restarting containers..."
    sudo docker compose restart
    ;;
  rebuild)
    echo "Rebuilding and restarting containers..."
    sudo docker compose down
    sudo docker compose build
    sudo docker compose up -d
    ;;
  *)
    echo "Usage: $0 {build|up|down|logs|restart|rebuild}"
    exit 1
    ;;
esac

exit 0 