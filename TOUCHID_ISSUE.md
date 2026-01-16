# Touch ID Prompt Issue

## Issue
Touch ID unlock is working but not showing the proper biometric prompt. Currently shows system password prompt instead of fingerprint dialog.

## Current Behavior
- Clicking "Unlock with Touch ID" button successfully retrieves password
- macOS shows system password dialog instead of Touch ID prompt
- Password is retrieved from keychain without biometric authentication

## Expected Behavior
- macOS should show native Touch ID dialog: "Unlock RootLock vault with Touch ID"
- User places finger on Touch ID sensor
- Authentication succeeds → vault unlocks
- Authentication fails → show error

## Technical Details

Current implementation uses AppleScript with LocalAuthentication framework:
```applescript
set theContext to current application's LAContext's alloc()'s init()
set authResult to theContext's evaluatePolicy:2 localizedReason:"..." reply:(missing value) error:(theError)
```

### Problem
`evaluatePolicy` is **asynchronous** in the LocalAuthentication framework. The AppleScript version executes synchronously, which may not properly trigger the Touch ID UI in all cases.

## Potential Solutions

### Option 1: Use security-framework Rust crate (Recommended)
- Use `security-framework` crate with proper async handling
- Access `SecAccessControl` with `kSecAccessControlUserPresence` flag
- Handle `LAContext.evaluatePolicy` properly with callbacks

### Option 2: Use ObjC/Swift bridge
- Create native Swift/ObjC module
- Call LocalAuthentication framework directly
- Expose via Tauri commands

### Option 3: Improve AppleScript approach
- Use `with administrator privileges` or system events
- Add proper error handling for async operations
- Test on multiple macOS versions

### Option 4: Use Native Module
- Create a small native macOS helper binary
- Link against LocalAuthentication.framework
- Call from Rust via command execution

## Files Involved
- `desktop/src-tauri/src/biometric.rs`
- `desktop/src-tauri/Cargo.toml`
- `desktop/src/contexts/AuthContext.tsx`

## References
- [LocalAuthentication Framework](https://developer.apple.com/documentation/localauthentication)
- [security-framework crate](https://docs.rs/security-framework/)
- [Keychain Services](https://developer.apple.com/documentation/security/keychain_services)
- [LAContext evaluatePolicy](https://developer.apple.com/documentation/localauthentication/lacontext/1514176-evaluatepolicy)

## Priority
Medium - Feature works but UX is not ideal. Users can still unlock with password.

## Labels
bug, enhancement, macos, biometric, touch-id

## Workaround
Users can unlock with master password. The biometric feature works technically (retrieves password correctly), just doesn't show the ideal Touch ID UI.
