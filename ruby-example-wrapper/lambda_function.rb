# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

require 'json'

def lambda_handler(event:, context:)
  begin
    my_string = 'my string literal'
    my_string << 'modifying string literal'
  rescue FrozenError => ex
    puts ex
  end

  { statusCode: 200, body: JSON.generate(my_string) }
end