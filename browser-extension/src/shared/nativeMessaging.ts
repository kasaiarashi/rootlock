// Native messaging utilities for communicating with RootLock desktop app

const NATIVE_HOST_NAME = 'com.rootlock.native';

interface NativeRequest {
  type: string;
  requestId: string;
  url?: string;
  domain?: string;
  data?: any;
}

interface NativeResponse {
  requestId: string;
  success: boolean;
  data?: any;
  error?: string;
}

let port: chrome.runtime.Port | null = null;
let messageHandlers: Map<string, (response: NativeResponse) => void> = new Map();

export function connectToNativeHost(): boolean {
  if (port) {
    return true; // Already connected
  }

  try {
    console.log('[Native Messaging] Connecting to host:', NATIVE_HOST_NAME);
    port = chrome.runtime.connectNative(NATIVE_HOST_NAME);

    port.onMessage.addListener((message: NativeResponse) => {
      console.log('[Native Messaging] Received response:', message);
      
      const handler = messageHandlers.get(message.requestId);
      if (handler) {
        handler(message);
        messageHandlers.delete(message.requestId);
      }
    });

    port.onDisconnect.addListener(() => {
      console.log('[Native Messaging] Disconnected from host');
      if (chrome.runtime.lastError) {
        console.error('[Native Messaging] Error:', chrome.runtime.lastError.message);
      }
      port = null;
      messageHandlers.clear();
    });

    console.log('[Native Messaging] Connected successfully');
    return true;
  } catch (error) {
    console.error('[Native Messaging] Failed to connect:', error);
    return false;
  }
}

export function sendNativeMessage(request: Omit<NativeRequest, 'requestId'>): Promise<NativeResponse> {
  return new Promise((resolve, reject) => {
    if (!port) {
      if (!connectToNativeHost()) {
        reject(new Error('Failed to connect to native host'));
        return;
      }
    }

    const requestId = crypto.randomUUID();
    const fullRequest: NativeRequest = {
      ...request,
      requestId,
    };

    console.log('[Native Messaging] Sending request:', fullRequest);

    messageHandlers.set(requestId, (response: NativeResponse) => {
      if (response.success) {
        resolve(response);
      } else {
        reject(new Error(response.error || 'Unknown error'));
      }
    });

    try {
      port!.postMessage(fullRequest);
    } catch (error) {
      messageHandlers.delete(requestId);
      reject(error);
    }

    // Timeout after 10 seconds
    setTimeout(() => {
      if (messageHandlers.has(requestId)) {
        messageHandlers.delete(requestId);
        reject(new Error('Request timeout'));
      }
    }, 10000);
  });
}

export async function getVaultStatus() {
  const response = await sendNativeMessage({ type: 'getStatus' });
  return response.data;
}

export async function getCredentials(url: string, domain: string) {
  const response = await sendNativeMessage({
    type: 'getCredentials',
    url,
    domain,
  });
  return response.data;
}

export async function generatePassword() {
  const response = await sendNativeMessage({ type: 'generatePassword' });
  return response.data;
}

// Initialize connection when background script loads
export function initializeNativeMessaging() {
  console.log('[Native Messaging] Initializing...');
  connectToNativeHost();
}
