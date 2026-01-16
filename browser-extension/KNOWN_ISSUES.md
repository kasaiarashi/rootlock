# Known Issues

## Non-Critical JSON Parse Error on Build

**Symptom:**
When running `npm run build` or `npm run dev`, you see:
```
SyntaxError: Unexpected token '<', "<head><tit"... is not valid JSON
    at JSON.parse (<anonymous>)
    at IncomingMessage.<anonymous> (vite-plugin-web-extension/dist/index.js:859:20)
```

**Impact:** 
**None** - This is a cosmetic error only. The extension builds successfully despite this error.

**Cause:**
The `vite-plugin-web-extension` package tries to check for updates by fetching a URL, but receives HTML instead of JSON. This is an upstream bug in the plugin.

**Verification:**
Check the `dist/` folder after build - all files are present and the extension works correctly.

**Workaround:**
Ignore the error. Look for the line before it:
```
✓ All steps completed.
```

If you see that, your build was successful.

**Files to check after build:**
```
dist/
├── src/
│   ├── popup/
│   │   ├── index.html ✓
│   │   └── index.js   ✓
│   ├── background/
│   │   └── service-worker.js ✓
│   └── content/
│       └── autofill.js ✓
├── manifest.json ✓
└── assets/ ✓
```

All these files should be present and the extension will load correctly in Chrome.
