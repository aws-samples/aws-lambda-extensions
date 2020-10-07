# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

function handler () {
  EVENT_DATA=$1
  echo "[function] Receiving invocation: '$EVENT_DATA'"
  RESPONSE="Echoing request: '$EVENT_DATA'"
  echo $RESPONSE
}