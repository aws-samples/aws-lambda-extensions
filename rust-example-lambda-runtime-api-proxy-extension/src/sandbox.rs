//
// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0
//

//! Interact with the Lambda Runtime API, the service managing this sandbox
//!
//! Includes helpers for sending request for `next` and posting back responses.
//!

use std::sync::Arc;

use hyper::{Body, Error, HeaderMap, Request, Response};

pub async fn next(headers: &HeaderMap, path: &str) -> Result<(Arc<String>, Response<Body>), Error> {
    let uri = hyper::Uri::builder()
        .scheme("http")
        .authority(crate::env::sandbox_runtime_api())
        .path_and_query(path)
        .build()
        .expect("[LRAP] Error building Sandbox Lambda Runtime API endpoint URL");

    let mut req = Request::builder()
        .method("GET")
        .uri(uri)
        .body(Body::empty())
        .expect("Cannot create Sandbox Lambda Runtime API request");

    *req.headers_mut() = headers.clone();

    let response = hyper::Client::new().request(req).await?;

    match response.headers().get("lambda-runtime-aws-request-id") {
        Some(id) => {
            let id = id.to_str().expect("Error parsing Lambda Runtime API request ID");
            Ok((Arc::new(id.to_string()), response))
        },
        // PANIC OK: when Lambda Runtime API does not meet its API contract, we kill the application
        _ => panic!("[LRAP] Sandbox Lambda Runtime API response missing 'lambda-runtime-aws-request-id' header in Lambda Runtime API GET:next response") 
    }
}

/// Send a request through a {hyper::Client}
pub async fn send_request(request: Request<Body>) -> Result<Response<Body>, Error> {
    hyper::Client::new().request(request).await
}

#[allow(dead_code)]
pub async fn create_invoke_result_request(id: &str, body: Body) -> Result<Request<Body>, Error> {
    let uri = hyper::Uri::builder()
        .scheme("http")
        .authority(crate::env::sandbox_runtime_api())
        .path_and_query(format!(
            "/{}/runtime/invocation/{}/response",
            crate::LAMBDA_RUNTIME_API_VERSION,
            id
        ))
        .build()
        .expect("[LRAP] Error building Sandbox Lambda Runtime API endpoint URL");

    Ok(hyper::Request::builder()
        .method("POST")
        .uri(uri)
        .body(body)
        .expect("Cannot create Sandbox Lambda Runtime API request"))
}

/// Lambda Extensions API
///
/// Interact with the Lambda sandbox as a Lambda Extension
///
#[allow(dead_code)]
pub mod extension {
    use hyper::Body;
    use once_cell::sync::OnceCell;

    /// Cannonical Lambda Extensions API version
    ///
    /// Documentation: https://docs.aws.amazon.com/lambda/latest/dg/runtimes-extensions-api.html
    ///
    const EXTENSION_API_VERSION: &str = "2020-01-01";
    static LAMBDA_EXTENSION_IDENTIFIER: OnceCell<String> = OnceCell::new();

    fn find_extension_name() -> String {
        crate::EXTENSION_NAME.to_owned()
    }

    pub(super) fn extension_id() -> &'static String {
        LAMBDA_EXTENSION_IDENTIFIER
            .get()
            .expect("[LRAP:Extension] Lambda Extension Identifier not set!")
    }

    fn make_uri(path: &str) -> hyper::Uri {
        hyper::Uri::builder()
            .scheme("http")
            .authority(crate::env::sandbox_runtime_api())
            .path_and_query(format!("/{}/extension{}", EXTENSION_API_VERSION, path))
            .build()
            .expect("[LRAP:Extension] Error building Lambda Extensions API endpoint URL")
    }

    /// Register the extension with the Lambda Extensions API
    pub async fn register() {
        let uri = make_uri("/register");

        let body = hyper::Body::from(r#"{"events":["INVOKE"]}"#);
        let mut request = hyper::Request::builder()
            .method("POST")
            .uri(uri)
            .body(body)
            .expect("[LRAP:Extension] Cannot create Lambda Extensions API request");

        // Set Lambda Extension Name header
        request.headers_mut().append(
            "Lambda-Extension-Name",
            find_extension_name().try_into().unwrap(),
        );

        let response = super::send_request(request)
            .await
            .expect("[LRAP:Extension] Cannot send Lambda Extensions API request to register");

        let extension_identifier = response
            .headers()
            .get("lambda-extension-identifier")
            .expect("[LRAP:Extension] Lambda Extensions API response missing 'lambda-extension-identifier' header in Lambda Extensions API POST:register response")
            .to_str()
            .unwrap();

        LAMBDA_EXTENSION_IDENTIFIER
            .set(extension_identifier.to_owned())
            .expect("[LRAP:Extension] Error setting Lambda Extensions API request ID");
    }

    /// Get next event from the Lambda Extensions API
    ///
    pub async fn get_next() {
        let uri = make_uri("/event/next");

        let mut request = hyper::Request::builder()
            .method("GET")
            .uri(uri)
            .body(Body::empty())
            .expect("[LRAP:Extension] Cannot create Lambda Extensions API request");

        request.headers_mut().insert(
            "Lambda-Extension-Identifier",
            extension_id().try_into().unwrap(),
        );

        // do not care about result because we get next payload through the Runtime API Proxy
        let _result = super::send_request(request).await;
    }
}
