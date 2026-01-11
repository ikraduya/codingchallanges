#!/bin/bash

set -x

# make

curl -X POST http://localhost:8080/ \
-H "Content-Type: application/json" \
-d '{"url": "https://www.examples.com"}'
echo ""

curl -X POST http://localhost:8080/ \
-H "Content-Type: application/json" \
-d '{"wrong key": "https://www.example.com"}' -i
echo ""


curl http://localhost:8080/cdb4d8 -i
curl http://localhost:8080/cdb4d9 -i

curl -X DELETE http://localhost:8080/cdb4d8 -i
curl -X DELETE http://localhost:8080/cdb4d9 -i

