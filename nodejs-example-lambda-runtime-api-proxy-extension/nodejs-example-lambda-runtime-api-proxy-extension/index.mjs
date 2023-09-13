#!/usr/bin/env node

// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

import { ExtensionsApiClient } from './extensions-api-client.mjs'
import { RuntimeApiProxy } from './runtime-api-proxy.mjs';

console.log('[LRAP:index] starting...');

process.on('SIGINT', () => process.exit(0));
process.on('SIGTERM', () => process.exit(0));

new RuntimeApiProxy().start();
new ExtensionsApiClient().bootstrap();
