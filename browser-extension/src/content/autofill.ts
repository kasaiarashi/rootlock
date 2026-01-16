// Content script for form detection and autofill

console.log('RootLock content script loaded');

// Detect login forms on the page
function detectLoginForms(): HTMLFormElement[] {
  const forms = Array.from(document.querySelectorAll('form'));
  
  return forms.filter((form) => {
    const inputs = form.querySelectorAll('input');
    let hasPassword = false;
    let hasUsername = false;
    
    inputs.forEach((input) => {
      if (input.type === 'password') {
        hasPassword = true;
      }
      if (
        input.type === 'email' ||
        input.type === 'text' ||
        input.name.toLowerCase().includes('user') ||
        input.name.toLowerCase().includes('email') ||
        input.id.toLowerCase().includes('user') ||
        input.id.toLowerCase().includes('email')
      ) {
        hasUsername = true;
      }
    });
    
    return hasPassword && hasUsername;
  });
}

// Add RootLock button to password fields
function addAutofillButton(passwordField: HTMLInputElement) {
  // Check if button already exists
  if (passwordField.dataset.rootlockButton === 'true') {
    return;
  }
  
  passwordField.dataset.rootlockButton = 'true';
  
  // Create button element
  const button = document.createElement('button');
  button.type = 'button';
  button.className = 'rootlock-autofill-btn';
  button.innerHTML = '🔐';
  button.title = 'Fill with RootLock';
  
  // Style the button
  Object.assign(button.style, {
    position: 'absolute',
    right: '8px',
    top: '50%',
    transform: 'translateY(-50%)',
    width: '24px',
    height: '24px',
    border: 'none',
    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    borderRadius: '4px',
    cursor: 'pointer',
    fontSize: '12px',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    zIndex: '10000',
    boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
  });
  
  // Add click handler
  button.addEventListener('click', async (e) => {
    e.preventDefault();
    e.stopPropagation();
    
    // Request credentials from background script
    const response = await chrome.runtime.sendMessage({
      type: 'GET_CREDENTIALS',
      url: window.location.href,
      domain: window.location.hostname,
    });
    
    console.log('Credentials response:', response);
    
    // TODO: Show credential selection UI
    alert('Autofill coming soon! Open RootLock desktop app to manage passwords.');
  });
  
  // Position the password field relatively if needed
  const fieldPosition = window.getComputedStyle(passwordField).position;
  if (fieldPosition === 'static') {
    passwordField.style.position = 'relative';
  }
  
  // Insert button after password field
  passwordField.parentElement?.style.position === 'relative' 
    ? passwordField.parentElement.appendChild(button)
    : passwordField.insertAdjacentElement('afterend', button);
}

// Initialize autofill detection
function init() {
  // Detect forms on page load
  const loginForms = detectLoginForms();
  console.log('Found login forms:', loginForms.length);
  
  loginForms.forEach((form) => {
    const passwordFields = form.querySelectorAll<HTMLInputElement>('input[type="password"]');
    passwordFields.forEach(addAutofillButton);
  });
  
  // Watch for dynamically added forms
  const observer = new MutationObserver(() => {
    const newForms = detectLoginForms();
    newForms.forEach((form) => {
      const passwordFields = form.querySelectorAll<HTMLInputElement>('input[type="password"]');
      passwordFields.forEach(addAutofillButton);
    });
  });
  
  observer.observe(document.body, {
    childList: true,
    subtree: true,
  });
}

// Wait for DOM to be ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', init);
} else {
  init();
}

// Listen for messages from background script
chrome.runtime.onMessage.addListener((message, _sender, sendResponse) => {
  if (message.type === 'INSERT_PASSWORD') {
    const passwordField = document.querySelector<HTMLInputElement>('input[type="password"]');
    if (passwordField) {
      passwordField.value = message.password;
      passwordField.dispatchEvent(new Event('input', { bubbles: true }));
      passwordField.dispatchEvent(new Event('change', { bubbles: true }));
    }
    sendResponse({ success: true });
  }
  return true;
});

export {};
