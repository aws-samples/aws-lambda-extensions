# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

echo "[function] Running init code outside handler"

function handler () {
  EVENT_DATA=$1
  echo "[function] handler receiving invocation: '$EVENT_DATA'"
  sleep 1
  RESPONSE="Echoing request: '$EVENT_DATA'"
  echo $RESPONSE
}