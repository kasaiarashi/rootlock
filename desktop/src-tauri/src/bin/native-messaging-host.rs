use serde::{Deserialize, Serialize};
use std::io::{self, Read, Write};

#[derive(Debug, Deserialize)]
struct ExtensionRequest {
    #[serde(rename = "type")]
    request_type: String,
    #[serde(rename = "requestId")]
    request_id: String,
    url: Option<String>,
    domain: Option<String>,
}

#[derive(Debug, Serialize)]
struct HostResponse {
    #[serde(rename = "requestId")]
    request_id: String,
    success: bool,
    data: Option<serde_json::Value>,
    error: Option<String>,
}

// Native messaging protocol: messages are prefixed with 4-byte length (little-endian)
fn read_message() -> io::Result<String> {
    let mut length_bytes = [0u8; 4];
    io::stdin().read_exact(&mut length_bytes)?;
    
    let length = u32::from_le_bytes(length_bytes) as usize;
    
    let mut buffer = vec![0u8; length];
    io::stdin().read_exact(&mut buffer)?;
    
    String::from_utf8(buffer).map_err(|e| io::Error::new(io::ErrorKind::InvalidData, e))
}

fn write_message(message: &str) -> io::Result<()> {
    let length = message.len() as u32;
    let length_bytes = length.to_le_bytes();
    
    io::stdout().write_all(&length_bytes)?;
    io::stdout().write_all(message.as_bytes())?;
    io::stdout().flush()?;
    
    Ok(())
}

fn handle_request(request: ExtensionRequest) -> HostResponse {
    match request.request_type.as_str() {
        "getCredentials" => {
            // TODO: Get credentials from vault for this URL/domain
            // For now, return mock data
            let credentials = serde_json::json!([
                {
                    "id": "1",
                    "name": "GitHub Account",
                    "username": "user@example.com",
                    "url": request.url.unwrap_or_default(),
                }
            ]);
            
            HostResponse {
                request_id: request.request_id,
                success: true,
                data: Some(credentials),
                error: None,
            }
        }
        
        "getStatus" => {
            // TODO: Check if vault is unlocked
            let status = serde_json::json!({
                "connected": true,
                "locked": false,
                "version": "0.1.0"
            });
            
            HostResponse {
                request_id: request.request_id,
                success: true,
                data: Some(status),
                error: None,
            }
        }
        
        "generatePassword" => {
            // TODO: Generate password using crypto utilities
            let password = serde_json::json!({
                "password": "GeneratedPassword123!@#"
            });
            
            HostResponse {
                request_id: request.request_id,
                success: true,
                data: Some(password),
                error: None,
            }
        }
        
        _ => {
            HostResponse {
                request_id: request.request_id,
                success: false,
                data: None,
                error: Some(format!("Unknown request type: {}", request.request_type)),
            }
        }
    }
}

fn main() {
    // Log to stderr (stdout is used for messaging)
    eprintln!("RootLock native messaging host started");
    
    loop {
        match read_message() {
            Ok(message_str) => {
                eprintln!("Received message: {}", message_str);
                
                match serde_json::from_str::<ExtensionRequest>(&message_str) {
                    Ok(request) => {
                        let response = handle_request(request);
                        
                        match serde_json::to_string(&response) {
                            Ok(response_str) => {
                                if let Err(e) = write_message(&response_str) {
                                    eprintln!("Failed to write response: {}", e);
                                    break;
                                }
                            }
                            Err(e) => {
                                eprintln!("Failed to serialize response: {}", e);
                            }
                        }
                    }
                    Err(e) => {
                        eprintln!("Failed to parse request: {}", e);
                    }
                }
            }
            Err(e) => {
                eprintln!("Failed to read message: {}", e);
                break;
            }
        }
    }
    
    eprintln!("RootLock native messaging host stopped");
}
