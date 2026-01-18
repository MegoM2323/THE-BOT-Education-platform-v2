/**
 * Detect if browser is Safari (any platform)
 * @returns {boolean} true if Safari
 */
export function isSafari() {
  const ua = navigator.userAgent;
  const vendor = navigator.vendor;
  return /Safari/.test(ua) && /Apple/.test(vendor) && !/Chrome|CriOS/.test(ua);
}

/**
 * Detect if platform is iOS (iPhone, iPad, iPod)
 * @returns {boolean} true if iOS
 */
export function isIOS() {
  const ua = navigator.userAgent;
  const platform = navigator.platform;
  const touchPoints = navigator.maxTouchPoints;

  if (/iPhone|iPad|iPod/.test(ua) || /iPhone|iPad|iPod/.test(platform)) {
    return true;
  }

  if (platform === 'MacIntel' && touchPoints > 1) {
    return true;
  }

  return false;
}

/**
 * Detect if browser is Safari on iOS
 * @returns {boolean} true if Safari on iOS
 */
export function isSafariOnIOS() {
  return isSafari() && isIOS();
}

/**
 * Open Telegram bot using optimal method
 * @param {string} botUrl - Telegram bot URL
 * @returns {object} { method, redirected?, blocked? }
 */
export function openTelegramBot(botUrl) {
  if (isSafariOnIOS()) {
    window.location.href = botUrl;
    return { method: 'redirect', redirected: true };
  }

  if (isSafari()) {
    const popup = window.open(botUrl, '_blank', 'noopener,noreferrer');
    return { method: 'popup', blocked: !popup };
  }

  const popup = window.open(botUrl, '_blank', 'noopener,noreferrer');
  return { method: 'popup', blocked: !popup };
}

export default { isSafari, isIOS, isSafariOnIOS, openTelegramBot };
