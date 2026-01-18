/**
 * Color utility functions for text contrast detection
 * Implements WCAG contrast ratio calculation for accessibility
 */
import { logger } from '../utils/logger.js';

/**
 * Convert hex color to RGB values
 * Handles both #RRGGBB and RRGGBB formats
 *
 * @param {string} hex - Hex color code (with or without #)
 * @returns {{r: number, g: number, b: number}} RGB values (0-255)
 */
export function hexToRgb(hex) {
  // Remove # if present
  const cleanHex = hex.replace('#', '');

  // Handle invalid input
  if (cleanHex.length !== 6) {
    logger.warn(`Invalid hex color: ${hex}, using default #000000`);
    return hexToRgb('#000000');
  }

  const r = parseInt(cleanHex.substring(0, 2), 16);
  const g = parseInt(cleanHex.substring(2, 4), 16);
  const b = parseInt(cleanHex.substring(4, 6), 16);

  return { r, g, b };
}

/**
 * Calculate relative luminance using WCAG formula
 * Formula: L = 0.2126 * R + 0.7152 * G + 0.0722 * B
 * where R, G, B are the linearized RGB values
 *
 * @param {number} r - Red value (0-255)
 * @param {number} g - Green value (0-255)
 * @param {number} b - Blue value (0-255)
 * @returns {number} Relative luminance (0-1)
 */
export function calculateLuminance(r, g, b) {
  // Convert RGB values to 0-1 range
  const [rNorm, gNorm, bNorm] = [r, g, b].map(val => {
    const normalized = val / 255;

    // Linearize the RGB values (gamma correction)
    // WCAG formula for sRGB
    return normalized <= 0.03928
      ? normalized / 12.92
      : Math.pow((normalized + 0.055) / 1.055, 2.4);
  });

  // Calculate relative luminance using WCAG coefficients
  // These coefficients reflect human eye sensitivity to different colors
  const luminance = 0.2126 * rNorm + 0.7152 * gNorm + 0.0722 * bNorm;

  return luminance;
}

/**
 * Calculate contrast ratio between two colors
 * WCAG formula: (L1 + 0.05) / (L2 + 0.05)
 * where L1 is the lighter color and L2 is the darker
 *
 * @param {number} luminance1 - Luminance of first color
 * @param {number} luminance2 - Luminance of second color
 * @returns {number} Contrast ratio (1-21)
 */
export function getContrastRatio(luminance1, luminance2) {
  const lighter = Math.max(luminance1, luminance2);
  const darker = Math.min(luminance1, luminance2);

  return (lighter + 0.05) / (darker + 0.05);
}

/**
 * Get optimal text color (black or white) for given background color
 * Uses relative luminance to determine which text color provides better contrast
 *
 * WCAG AA requirements:
 * - Normal text: 4.5:1 contrast ratio
 * - Large text: 3:1 contrast ratio
 *
 * @param {string} hexColor - Background color in hex format
 * @returns {string} "#000000" for dark backgrounds, "#FFFFFF" for light backgrounds
 */
export function getContrastTextColor(hexColor) {
  // Handle null/undefined
  if (!hexColor) {
    hexColor = '#000000'; // Default to black
  }

  // Convert hex to RGB
  const { r, g, b } = hexToRgb(hexColor);

  // Calculate luminance
  const luminance = calculateLuminance(r, g, b);

  // Calculate contrast ratios with white and black text
  const whiteLuminance = 1; // White has maximum luminance
  const blackLuminance = 0; // Black has minimum luminance

  const contrastWithWhite = getContrastRatio(luminance, whiteLuminance);
  const contrastWithBlack = getContrastRatio(luminance, blackLuminance);

  // Choose the color with better contrast
  // If background is bright (high luminance), use black text
  // If background is dark (low luminance), use white text
  return contrastWithBlack > contrastWithWhite ? '#000000' : '#FFFFFF';
}

/**
 * Verify if contrast ratio meets WCAG AA standards
 *
 * @param {string} bgColor - Background color hex
 * @param {string} fgColor - Foreground (text) color hex
 * @param {string} size - Text size: 'normal' or 'large'
 * @returns {{passes: boolean, ratio: number, required: number}}
 */
export function verifyContrastRatio(bgColor, fgColor, size = 'normal') {
  const bgRgb = hexToRgb(bgColor);
  const fgRgb = hexToRgb(fgColor);

  const bgLuminance = calculateLuminance(bgRgb.r, bgRgb.g, bgRgb.b);
  const fgLuminance = calculateLuminance(fgRgb.r, fgRgb.g, fgRgb.b);

  const ratio = getContrastRatio(bgLuminance, fgLuminance);

  // WCAG AA requirements
  const required = size === 'large' ? 3 : 4.5;

  return {
    passes: ratio >= required,
    ratio: Math.round(ratio * 100) / 100,
    required,
  };
}

/**
 * Get suggested text color with verification
 * Returns color information and contrast details
 *
 * @param {string} bgColor - Background color hex
 * @returns {{color: string, luminance: number, contrast: number, passes: boolean}}
 */
export function getSuggestedTextColor(bgColor) {
  const textColor = getContrastTextColor(bgColor);
  const bgRgb = hexToRgb(bgColor);
  const bgLuminance = calculateLuminance(bgRgb.r, bgRgb.g, bgRgb.b);

  const verification = verifyContrastRatio(bgColor, textColor, 'normal');

  return {
    color: textColor,
    luminance: Math.round(bgLuminance * 1000) / 1000,
    contrast: verification.ratio,
    passes: verification.passes,
  };
}

/**
 * Convert hex color to RGBA string
 *
 * @param {string} hex - Hex color code (with or without #)
 * @param {number} alpha - Alpha value (0-1)
 * @returns {string} RGBA color string
 */
export function hexToRgba(hex, alpha) {
  const { r, g, b } = hexToRgb(hex);
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}
