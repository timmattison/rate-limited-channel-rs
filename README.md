# rate-limited-channel-rs

A Rust library that provides a rate-limited channel implementation. This library allows you to create a channel that will only emit values at a specified rate, regardless of how quickly values are sent to it.

## Usage

Add this to your `Cargo.toml`:

```toml
[dependencies]
rate-limited-channel-rs = "0.1.0"
tokio = { version = "1", features = ["full"] }
```

### Example

```rust
use rate_limited_channel_rs::to_rate_limited_channel;
use std::time::Duration;
use tokio::sync::mpsc;

#[tokio::main]
async fn main() {
    // Create a channel with a buffer of 100 values
    let (tx, rx) = mpsc::channel::<i32>(100);
    
    // Create a rate-limited channel that emits at most one value per second
    let mut rate_limited_rx = to_rate_limited_channel(rx, Duration::from_secs(1));
    
    // Spawn a task that sends values as fast as possible
    tokio::spawn(async move {
        for i in 0..10 {
            // These values will be sent as fast as possible
            tx.send(i).await.unwrap();
        }
    });
    
    // Receive values at the rate-limited pace
    for _ in 0..5 {
        if let Some(value) = rate_limited_rx.recv().await {
            println!("Received: {}", value);
            // Will print approximately one value per second
        }
    }
}
```

## How it Works

The rate-limited channel accepts values as fast as they are sent, but only emits them at the specified rate. If multiple values arrive between emitted values, only the most recent value will be emitted when the rate limit allows.

This is useful for scenarios where you want to limit how often an operation occurs while always using the most up-to-date data.
