# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

function handler () {
  EVENT_DATA=$1
  RESPONSE="Echoing request: '$EVENT_DATA'"
  echo $(curl -XGET -m 5 "http://localhost:2772/")
  echo $(cat /tmp/test.txt)
  echo $RESPONSE
}