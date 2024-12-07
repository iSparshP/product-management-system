#!/bin/bash

# scripts/migrate.sh

# Usage: ./scripts/migrate.sh up|down [steps]

COMMAND=$1
STEPS=$2

DATABASE_URL="postgres://youruser:yourpassword@localhost:5432/productdb?sslmode=disable"

migrate -database "$DATABASE_URL" -path migrations/001_create_users_and_products up

# For down migrations, adjust accordingly
# migrate -database "$DATABASE_URL" -path migrations/001_create_users_and_products down
