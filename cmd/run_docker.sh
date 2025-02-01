#!/bin/bash
docker run -d -p 5433:5432 --name db -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} -v $(pwd)/db_init:/docker-entrypoint-initdb.d postgres