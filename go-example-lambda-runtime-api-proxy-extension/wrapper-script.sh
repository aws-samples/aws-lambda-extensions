#!/bin/bash
args=("$@")
export AWS_LAMBDA_RUNTIME_API="127.0.0.1:${LRAP_LISTENER_PORT:-9009}"
exec "${args[@]}"