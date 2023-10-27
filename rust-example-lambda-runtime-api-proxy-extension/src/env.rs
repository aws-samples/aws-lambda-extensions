//
// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0
//

//! Access the ENV for the Extension (and Proxy)
//!
//! Utilities and other helper functions for thread-safe access and lazy initializers
//!

use once_cell::sync::OnceCell;

/// Sandbox's Runtime API endpoint
static LAMBDA_RUNTIME_API: OnceCell<String> = OnceCell::new();

/// Lambda Runtime API Proxy (LRAP), this endpoint
static LRAP_API: OnceCell<String> = OnceCell::new();

/// Latch in the API endpoints defined in ENV variables
///
#[allow(dead_code)]
pub fn latch_runtime_env() {
    use std::env::var;

    let aws_lambda_runtime_api =
        match var("LRAP_RUNTIME_API_ENDPOINT").or_else(|_| var("AWS_LAMBDA_RUNTIME_API")) {
            Ok(v) => v,
            Err(_) => panic!("LRAP_RUNTIME_API_ENDPOINT or AWS_LAMBDA_RUNTIME_API not found"),
        };

    // Latch in the ORIGIN we should proxy to the application
    LAMBDA_RUNTIME_API.set(aws_lambda_runtime_api.clone())
        .expect("Expected that mutate_runtime_env() has not been called before, but AWS_LAMBDA_RUNTIME_API was already set");

    let listener_port = var("LRAP_LISTENER_PORT")
        .ok()
        .and_then(|v| v.parse::<u16>().ok())
        .or(Some(crate::DEFAULT_PROXY_PORT))
        .unwrap();

    let lrap_api = format!("127.0.0.1:{}", listener_port);

    LRAP_API.set(lrap_api.clone()).expect("aws_lambda_runtime_api_proxy_rs::env::LRAP_API was previously initialized and should not be");
}

/// Gets the original AWS_LAMBDA_RUNTIME_API.
///
#[allow(dead_code)]
pub fn sandbox_runtime_api() -> &'static str {
    match LAMBDA_RUNTIME_API.get() {
        Some(val) => val,
        None => {
            latch_runtime_env();
            LAMBDA_RUNTIME_API.get().expect(
                "Error in setting and mutating AWS_LAMBDA_RUNTIME_API environment variables.",
            )
        }
    }
}

/// Gets the new LRAP_API.
///
pub fn lrap_api() -> &'static str {
    match LRAP_API.get() {
        Some(val) => val,
        None => {
            latch_runtime_env();
            LRAP_API.get().expect("Error in setting and mutating AWS_LAMBDA_RUNTIME_API dependent LRAP_API host:port.")
        }
    }
}
