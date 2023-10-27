//
// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0
//

use std::{
    default::Default,
    io::Error,
    pin::Pin,
    sync::Arc,
    task::{Context, Poll},
};

use tokio::io::{AsyncRead, AsyncWrite};

/// Buffer to store chunks of data read from a stream not to exceed a limit
///
/// Implements [`AsyncWrite`] and [`AsyncRead`]
pub struct LimitedBuffer {
    /// max size of the buffer
    capacity: usize,

    /// vec of data pages
    buffer: Vec<Vec<u8>>,

    /// length of data stored across all pages
    content_len: usize,
}

#[allow(dead_code)]
impl LimitedBuffer {
    pub fn new(capacity: usize) -> Self {
        Self {
            capacity,
            buffer: Default::default(),
            content_len: 0,
        }
    }

    /// Get the content length of the buffer
    pub fn len(&self) -> usize {
        self.content_len
    }

    /// Get the capacity of the
    pub fn capacity(&self) -> usize {
        self.capacity
    }

    /// space remaining
    pub fn remaining(&self) -> usize {
        self.capacity - self.content_len
    }

    pub fn into_arc(self) -> Arc<LimitedBuffer> {
        Arc::new(self)
    }
}

impl AsyncWrite for LimitedBuffer {
    fn poll_write(
        self: Pin<&mut Self>,
        _cx: &mut Context<'_>,
        buf: &[u8],
    ) -> Poll<Result<usize, Error>> {
        let this = Pin::into_inner(self);

        let limit_len = this.capacity - this.content_len;

        if limit_len == 0 {
            return Poll::Ready(Ok(0));
        }

        let copy_len = std::cmp::min(this.capacity - this.content_len, buf.len());
        this.buffer.push(buf[0..(copy_len)].to_vec());
        this.content_len += copy_len;

        return Poll::Ready(Ok(copy_len));
    }

    /// flush is a no-op
    fn poll_flush(
        self: std::pin::Pin<&mut Self>,
        _cx: &mut std::task::Context<'_>,
    ) -> std::task::Poll<Result<(), std::io::Error>> {
        Poll::Ready(Ok(()))
    }

    /// no resources to free
    fn poll_shutdown(
        self: std::pin::Pin<&mut Self>,
        _cx: &mut std::task::Context<'_>,
    ) -> std::task::Poll<Result<(), std::io::Error>> {
        Poll::Ready(Ok(()))
    }
}

/// Reader interface the [`LimitedBuffer`]
///
pub struct LimitedBufferReader {
    /// the buffer to read from
    buffer: Arc<LimitedBuffer>,

    /// data page index
    page_index: usize,

    /// offset within the data page
    page_offset: usize,

    /// logical byte offset
    byte_offset: usize,
}

#[allow(dead_code)]
impl LimitedBufferReader {
    pub fn new(buffer: Arc<LimitedBuffer>) -> Self {
        Self {
            buffer,
            page_index: 0,
            page_offset: 0,
            byte_offset: 0,
        }
    }

    /// Get the logical offset of the cursor position in the buffer
    pub fn pos(&self) -> usize {
        self.byte_offset
    }

    /// Get the remaining length of the buffer
    pub fn remaining(&self) -> usize {
        self.buffer.content_len - self.byte_offset
    }

    /// Move the cursor to a byte offset within the buffer
    pub fn seek_to(&mut self, offset: usize) {
        let mut advance = offset;
        self.page_index = 0;
        self.page_offset = 0;

        // if offset is larger than buffer content, position to the end of buffer
        if offset >= self.buffer.content_len {
            self.page_index = self.buffer.buffer.len();
            self.page_offset = 0;
            self.byte_offset = self.buffer.content_len;
            return;
        }

        self.page_index = 0;
        self.page_offset = 0;
        self.byte_offset = 0;

        // position to beginning of buffer
        if offset == 0 {
            return;
        }

        // iteratively traverse the pages to the logical offset
        loop {
            let page_len = self.buffer.buffer[self.page_index].len();
            // if advance is larger than current page
            if advance > page_len {
                self.page_index += 1;
                self.page_offset = 0;
                advance -= page_len;
            } else {
                self.page_offset = advance;
                self.byte_offset = offset;
                return;
            }
        }
    }
}

impl AsyncRead for LimitedBufferReader {
    fn poll_read(
        self: Pin<&mut Self>,
        _cx: &mut Context<'_>,
        buf: &mut tokio::io::ReadBuf<'_>,
    ) -> Poll<std::io::Result<()>> {
        let this = Pin::into_inner(self);

        // no data
        if this.buffer.buffer.len() == 0 {
            return Poll::Ready(Ok(()));
        }

        // page_index is pointing to EOF
        if this.page_index == this.buffer.buffer.len() {
            return Poll::Ready(Ok(()));
        }

        let remaining = buf.remaining();
        let copy_len = std::cmp::min(
            remaining,
            this.buffer.buffer[this.page_index].len() - this.page_offset,
        );
        buf.put_slice(
            &this.buffer.buffer[this.page_index][this.page_offset..(this.page_offset + copy_len)],
        );

        this.page_offset += copy_len; // advance read pointer
        this.byte_offset += copy_len; // advance logical position

        if this.page_offset == this.buffer.buffer[this.page_index].len() {
            // advance to next page
            this.page_index += 1;
            this.page_offset = 0;
        }

        return Poll::Ready(Ok(()));
    }
}

#[cfg(test)]
mod test {
    use super::{LimitedBuffer, LimitedBufferReader};

    use std::future::Future;
    use std::sync::Arc;

    use tokio::io::{AsyncReadExt, AsyncWriteExt};
    use tokio::runtime;

    fn async_invoke<F: Future>(f: F) -> <F as Future>::Output {
        runtime::Builder::new_current_thread()
            .enable_all()
            .build()
            .expect("Cannot create Tokio runtime")
            .block_on(f)
    }

    #[test]
    pub fn create() {
        let buffer = super::LimitedBuffer::new(20);
        assert!(buffer.len() == 0, "Buffer reports having data");
        assert!(
            buffer.remaining() == 20,
            "Buffer reports being partially filled"
        );
    }

    #[test]
    pub fn write() {
        async_invoke(async {
            let mut buffer = super::LimitedBuffer::new(20);
            buffer.write("one".as_bytes()).await.unwrap();
            buffer.write("two".as_bytes()).await.unwrap();

            assert!(buffer.len() == 6, "buffer len should be 6 with 'onetwo'");
        })
    }

    #[test]
    pub fn readback() {
        async_invoke(async {
            let mut buffer = super::LimitedBuffer::new(20);

            buffer.write("one".as_bytes()).await.unwrap();
            buffer.write("two".as_bytes()).await.unwrap();
            buffer
                .write("678901234567890ABCDE".as_bytes())
                .await
                .unwrap();
            assert_eq!(
                buffer.len(),
                20,
                "buffer len should be 20; truncated to capacity"
            );

            let mut reader = super::LimitedBufferReader::new(Arc::new(buffer));
            let mut buf = [0 as u8; 5];

            let bytes_read = reader.read(&mut buf).await.unwrap();
            assert_eq!(bytes_read, 3, "Should have read first page");
            assert_eq!(&buf[0..3], "one".as_bytes(), "Should have read first page");
            assert_eq!(reader.pos(), 3, "Should have read first page");

            let bytes_read = reader.read(&mut buf).await.unwrap();
            assert_eq!(bytes_read, 3, "Should have read second page");
            assert_eq!(&buf[0..3], "two".as_bytes(), "Should have read second page");
            assert_eq!(reader.pos(), 6, "Should have read second page");

            let bytes_read = reader.read(&mut buf).await.unwrap();
            assert_eq!(bytes_read, 5, "Should have read 5 bytes of third page");
            assert_eq!(
                buf,
                "67890".as_bytes(),
                "Should have read first 5 bytes of third page"
            );
            assert_eq!(
                reader.pos(),
                11,
                "Should have read first 5 bytes of third page"
            );

            let bytes_read = reader.read(&mut buf).await.unwrap();
            assert_eq!(bytes_read, 5, "Should have read 5 bytes of third page");
            assert_eq!(
                buf,
                "12345".as_bytes(),
                "Should have read first 5 bytes of third page"
            );
            assert_eq!(
                reader.pos(),
                16,
                "Should have read first 5 bytes of third page"
            );

            let bytes_read = reader.read(&mut buf).await.unwrap();
            assert_eq!(bytes_read, 4, "Should have read 5 bytes of third page");
            assert_eq!(
                &buf[0..4],
                "6789".as_bytes(),
                "Should have read first 5 bytes of third page"
            );
            assert_eq!(
                reader.pos(),
                20,
                "Should have read first 5 bytes of third page"
            );

            // seek to position
            reader.seek_to(4);
            assert_eq!(reader.pos(), 4, "Did not seek to pos=4 or report correctly");
            let bytes_read = reader.read(&mut buf).await.unwrap();
            assert_eq!(bytes_read, 2, "Should have read 2 bytes of second page");
            assert_eq!(
                &buf[0..2],
                "wo".as_bytes(),
                "Should have read first 2 bytes of second page"
            );
            assert_eq!(reader.pos(), 6, "Should have read thru second page");
        })
    }

    #[test]
    pub fn as_stream_bytes() {
        use hyper::Body;
        use tokio_util::io::ReaderStream;

        async_invoke(async {
            let buffer = LimitedBuffer::new(20);

            {
                let reader = LimitedBufferReader::new(Arc::new(buffer));
                let _body = Body::wrap_stream(ReaderStream::new(reader));
            }
        })
    }
}
