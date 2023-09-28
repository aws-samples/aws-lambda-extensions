//
// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0
//

//! Hold global-state of timing metrics for Application processing event and LRAP extension latency
//!
use std::time::Instant;

use once_cell::sync::OnceCell;
use parking_lot::Mutex;

static INIT_START: OnceCell<Instant> = OnceCell::new();
static APP_START: OnceCell<Instant> = OnceCell::new();

static EVENT_START: Mutex<Option<Instant>> = Mutex::new(None);

pub fn init_start() {
    INIT_START.set(Instant::now()).unwrap();
}
pub fn app_start() {
    APP_START.set(Instant::now()).unwrap();
}

#[allow(dead_code)]
pub fn get_next_event() {
    match *EVENT_START.lock() {
        None => {
            eprintln!(
                "[LRAP] LRAP init     : {} us",
                APP_START
                    .get()
                    .unwrap()
                    .duration_since(*INIT_START.get().unwrap())
                    .as_micros()
            );
            eprintln!(
                "[LRAP] App  init     : {} us",
                APP_START.get().unwrap().elapsed().as_micros()
            );
        }
        Some(event_start) => {
            eprintln!(
                "[LRAP] App run time  : {} us",
                event_start.elapsed().as_micros()
            );
        }
    }
}

#[allow(dead_code)]
pub fn event_start() {
    EVENT_START.lock().replace(Instant::now());
}
