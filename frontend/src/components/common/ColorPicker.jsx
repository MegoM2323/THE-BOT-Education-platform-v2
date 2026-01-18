import { useState, useEffect } from 'react';
import './ColorPicker.css';

/**
 * Color picker component with preset palette and hex input
 * @param {string} value - Current color value (hex format #RRGGBB)
 * @param {function} onChange - Callback when color changes
 * @param {boolean} disabled - Whether the picker is disabled
 */
export const ColorPicker = ({ value = '#2563eb', onChange, disabled = false }) => {
  const [hexInput, setHexInput] = useState(value);
  const [showPicker, setShowPicker] = useState(false);

  // Preset color palette
  const colorPresets = [
    { color: '#660000', name: 'Тёмно-красный' },
    { color: '#666600', name: 'Коричневый' },
    { color: '#006600', name: 'Тёмно-зелёный' },
    { color: '#006666', name: 'Бирюзовый' },
    { color: '#000066', name: 'Тёмно-синий' },
    { color: '#660066', name: 'Фиолетовый' },
    { color: '#2563eb', name: 'Синий' },
    { color: '#000000', name: 'Чёрный' },
  ];

  // Sync hex input with prop value
  useEffect(() => {
    setHexInput(value);
  }, [value]);

  /**
   * Validate hex color format
   */
  const isValidHex = (hex) => {
    return /^#[0-9A-Fa-f]{6}$/.test(hex);
  };

  /**
   * Handle color selection from palette
   */
  const handlePresetClick = (color) => {
    setHexInput(color);
    onChange?.(color);
  };

  /**
   * Handle hex input change
   */
  const handleHexInputChange = (e) => {
    let input = e.target.value.trim();

    // Auto-add # prefix if missing
    if (input && !input.startsWith('#')) {
      input = '#' + input;
    }

    setHexInput(input);

    // Validate and notify parent
    if (isValidHex(input)) {
      onChange?.(input);
    }
  };

  /**
   * Handle hex input blur - reset to valid color if invalid
   */
  const handleHexInputBlur = () => {
    if (!isValidHex(hexInput)) {
      setHexInput(value);
    }
  };

  return (
    <div className="color-picker">
      <label className="color-picker-label">Цвет</label>

      <div className="color-picker-container">
        {/* Color preview square */}
        <button
          type="button"
          className="color-preview"
          style={{ backgroundColor: isValidHex(hexInput) ? hexInput : value }}
          onClick={() => !disabled && setShowPicker(!showPicker)}
          disabled={disabled}
          aria-label="Выберите цвет"
        >
          <div className="color-preview-inner" />
        </button>

        {/* Hex input */}
        <input
          type="text"
          className="color-hex-input"
          value={hexInput}
          onChange={handleHexInputChange}
          onBlur={handleHexInputBlur}
          placeholder="#2563eb"
          maxLength={7}
          disabled={disabled}
          aria-label="Введите hex код цвета"
        />

        {/* Toggle picker button */}
        <button
          type="button"
          className="color-picker-toggle"
          onClick={() => !disabled && setShowPicker(!showPicker)}
          disabled={disabled}
          aria-label="Открыть палитру"
        >
          {showPicker ? '▲' : '▼'}
        </button>
      </div>

      {/* Color palette dropdown */}
      {showPicker && !disabled && (
        <div className="color-palette">
          <div className="color-palette-header">Выберите цвет</div>
          <div className="color-palette-grid">
            {colorPresets.map(({ color, name }) => (
              <button
                key={color}
                type="button"
                className={`color-palette-item ${hexInput.toUpperCase() === color.toUpperCase() ? 'active' : ''}`}
                style={{ backgroundColor: color }}
                onClick={() => handlePresetClick(color)}
                title={name}
                aria-label={name}
              >
                {hexInput.toUpperCase() === color.toUpperCase() && (
                  <span className="color-palette-check">✓</span>
                )}
              </button>
            ))}
          </div>
          <div className="color-palette-footer">
            <small>Или введите hex код цвета (например, #2563eb)</small>
          </div>
        </div>
      )}

      {/* Validation error */}
      {hexInput && !isValidHex(hexInput) && (
        <div className="color-picker-error">
          Неверный формат. Используйте #RRGGBB (например, #2563eb)
        </div>
      )}
    </div>
  );
};

export default ColorPicker;
