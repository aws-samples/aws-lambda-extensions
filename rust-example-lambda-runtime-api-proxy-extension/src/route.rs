//
// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0
//

//! Routing for Runtime API requests.  This builds out the services and stitches them together as
//! well as builds routing tables for HTTP methods on resources to proxy the Lambda Runtime API.
//!
//!

use std::sync::Arc;

use httprouter::Router;
use hyper::{Body, Error, Request, Response, Uri};

use crate::{env, sandbox, stats, util::LimitedBuffer};

pub fn make_route<'a>() -> Router<'a> {
    // Route `invocation/next` demonstrates hooks for filtering incoming request events
    // Users can implement a similar patern in `invocation/:id/response` to filter responses
    let router = Router::default()
        .get("/", passthru_proxy)
        .get("/:apiver/runtime/invocation/next", proxy_invocation_next)
        .post("/:apiver/runtime/invocation/:id/response", passthru_proxy)
        .post("/:apiver/runtime/invocation/:id/error", passthru_proxy)
        .not_found(notfound_passthru_proxy);
    router
}

/// Pass-through the request, but log the unhandled path and method
#[allow(dead_code)]
pub async fn notfound_passthru_proxy(req: Request<Body>) -> Result<Response<Body>, Error> {
    eprintln!(
        "[LRAP] Route not found: path={} method={}",
        &req.uri().path(),
        &req.method()
    );
    passthru_proxy(req).await
}

#[allow(dead_code)]
pub async fn passthru_proxy(req: Request<Body>) -> Result<Response<Body>, Error> {
    // possible improvement: replace with resource pool or persistent connection
    let endpoint_client = hyper::Client::new();
    let endpoint_uri: Uri = Uri::builder()
        .scheme("http")
        .authority(env::sandbox_runtime_api())
        .path_and_query(req.uri().path())
        .build()
        .unwrap();

    // remap URI
    let mut endpoint_req: Request<Body> = req.into();
    *endpoint_req.uri_mut() = endpoint_uri.clone();

    let method = endpoint_req.method().clone();

    match endpoint_client.request(endpoint_req).await {
        Ok(res) => Ok(res),
        Err(e) => {
            eprintln!(
                "[LRAP] Error invoking endpoint ({} on {}): {:?}",
                method, endpoint_uri, e
            );
            Ok(Response::builder()
                .status(502)
                .body("502 - Bad Gateway: Lambda Runtime API did not process request".into())
                .unwrap())
        }
    }
}

/// Example of reading the HTTP body into a limited-length buffer for later processing
#[allow(dead_code)]
async fn hyper_body_to_body_buffer(
    size: usize,
    body: hyper::Body,
) -> std::sync::Arc<LimitedBuffer> {
    use futures::stream::StreamExt;
    use tokio_util::io::StreamReader;

    let mut body_buffer = LimitedBuffer::new(size);

    let mapped_stream = body.map(|chunk_result| {
        chunk_result.map_err(|hyper_err| {
            std::io::Error::new(
                std::io::ErrorKind::Other,
                format!("Hyper error: {}", hyper_err),
            )
        })
    });

    let mut reader = StreamReader::new(mapped_stream);
    tokio::io::copy(&mut reader, &mut body_buffer)
        .await
        .unwrap();
    std::sync::Arc::new(body_buffer)
}

/// Get next invocation; provide hooks for skipping bad requests (payload malicious or ill-formed)
///
/// Flow:
///
///          [App Runtime]               [LRAP]                        [Lambda Service]
///               |                         
///               +---- GET next event --->|
///                                        |
///                                 [ proxy request ]-- GET next event ------>|
///                                                                           |                             
///                                                                           |<---- [ INVOKE with payload ]
///                                        |<--------- event payload ---------|
///                                        |                                   
///                          [ if validation fails: DROP event ]                  
///                                        |                                   
///                                        |----------- GET next event ------>|
///                                                                           |<---- [ INVOKE with payload ]
///                                        |<--------- event payload ---------|
///                                        |                                   
///               |<-- event -----[ if validation succeeds: PASS event ]               
///               |   payload             
///               |                         
///           [ appp logic ]                
///               |                         
///               |--response payload ---->|
///                                        |                                   
///                              [ sanitize response ]-- response sanitized ->|
///                                                                           |----->[ synchronous response ]
///                                         
pub async fn proxy_invocation_next(req: Request<Body>) -> Result<Response<Body>, Error> {
    use std::time::Duration;

    'getNext: loop {
        // track either initialization  -or-
        // how long it took to process the event and request next
        //
        stats::get_next_event();

        let (aws_request_id, response) =
            match crate::sandbox::next(req.headers(), req.uri().path()).await {
                Err(e) => {
                    eprintln!(
                        "[LRAP]  Error getting next invocation from Runtime API: {}",
                        e
                    );
                    eprintln!("[LRAP] uri: {}", req.uri());
                    tokio::time::sleep(Duration::from_millis(100)).await;
                    continue 'getNext;
                }
                Ok(response) => response,
            };

        // start the counter on the new event
        stats::event_start();

        match validate_and_mangle_next_event(aws_request_id, response).await {
            Ok(response) => {
                return Ok(response);
            }
            Err(req) => {
                sandbox::send_request(req).await.ok();
                continue 'getNext;
            }
        }
    }
}

/// Process the next invocation event from the Lambda Runtime API
///
/// Event context, payload is in `response`
///
/// On Error, create a [`Request<Body>`] to send to the Runtime API.
///
/// This _could_ be a request to the Runtime API's /runtime/invocation/:id/response to short-cut the Application with a specific code
///
async fn validate_and_mangle_next_event(
    _aws_request_id: Arc<String>,
    response: Response<Body>,
) -> Result<Response<Body>, Request<Body>> {
    // implement event input filtering here

    return Ok(response);
}
