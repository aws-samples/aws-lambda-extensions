//
// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0
//

//! Lambda Extension application to proxy Lambda Runtime API requests between
//! the Application Runtime and the Lambda host.
//!
//! This extension uses Hyper, Tokio, and Rust futures
//!
//!

#[allow(unused_imports)]
use std::{
    convert::Infallible,
    io::{Read, Write},
    net::SocketAddr,
    process::Stdio,
    sync::Arc,
};

#[allow(unused_imports)]
use hyper::{Body, Request, Response, Server};

use tokio::{self};

/// ENV references to API endpoints (host:port)
mod env;

/// Routes for Lambda Runtime API
mod route;

/// Common utilities
pub mod util;

pub(crate) mod sandbox;
pub(crate) mod stats;

/// Name to register with the Lambda Extension API.
///
/// NOTE: this must be the same as the
/// entrypoint script destination in the Lambda layer (eg, **extensions/lrap**)
pub const EXTENSION_NAME: &str = "lrap";

/// Default port to listen on, overriden by LRAP_LISTENER_PORT environment variable
///
/// NOTE: this must be the same port as listed in **opt/wrapper** script that launches
/// the Application runtime with the modified `AWS_LAMBDA_RUNTIME_API` env variable.
pub const DEFAULT_PROXY_PORT: u16 = 9009;

pub static LAMBDA_RUNTIME_API_VERSION: &str = "2018-06-01";

/// Implement the Runtime API Proxy for Lambda:
///
/// 1. create a hyper server on the LRAP endpoint
///
/// 2. create a Tower service for the Lambda Runtime API to serve HTTP requests
///
/// 3. register as an Extension, allowing Application runtime to begin initializing
///
/// 4. request `next` event from Extension API, fulfilling lifecycle contract
///   
///
#[tokio::main]
async fn main() {
    stats::init_start();

    println!(
        "[LRAP] start; path={}",
        std::env::current_exe().unwrap().to_str().unwrap()
    );
    println!(
        "[LRAP] commandline arguments: {}",
        std::env::args()
            .map(|v| format!("\"{}\"", v))
            .collect::<Vec<String>>()
            .join(", ")
    );

    env::latch_runtime_env();

    let addr: SocketAddr = env::lrap_api()
        .parse()
        .expect("Invalid IP specification from Lambda Runtime API endpoint");
    println!("[LRAP] listening on {}", addr);

    // bind the server to the Lambda Runtime API Router service
    let server = Server::bind(&addr).serve(route::make_route().into_service());

    // launch the Proxy server task
    let server_join_handle = tokio::spawn(server);

    // Initialize the extension and continually get next extension event.
    // We ignore extension events because all LRAP capability is in the Proxy.
    tokio::task::spawn(async {
        sandbox::extension::register().await;
        // Lambda Application runtime will start once our extension is registered
        stats::app_start();

        loop {
            // Lambda Extension API requires we wait for next extension event
            sandbox::extension::get_next().await;
        }
    });

    match server_join_handle
        .await
        .expect("Failed to join the server task")
    {
        Err(e) => {
            eprintln!("[LRAP] Hyper server error: {}", e);
        }
        Ok(_) => { /* never reached */ }
    }
}
