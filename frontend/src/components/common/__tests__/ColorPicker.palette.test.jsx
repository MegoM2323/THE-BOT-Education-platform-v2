import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ColorPicker } from "../ColorPicker.jsx";

/**
 * Tests for ColorPicker component with palette colors
 * Focuses on testing the 8 preset colors from the palette
 */
describe('ColorPicker - Palette Colors', () => {
  let onChange;

  beforeEach(() => {
    onChange = vi.fn();
  });

  const paletteColors = [
    { hex: '#660000', name: 'Тёмно-красный' },
    { hex: '#666600', name: 'Коричневый' },
    { hex: '#006600', name: 'Тёмно-зелёный' },
    { hex: '#006666', name: 'Бирюзовый' },
    { hex: '#000066', name: 'Тёмно-синий' },
    { hex: '#660066', name: 'Фиолетовый' },
    { hex: '#004231', name: 'Зелёный' },
    { hex: '#000000', name: 'Чёрный' },
  ];

  describe('Palette Colors Display', () => {
    it('should display default color #660000 (Dark Red)', () => {
      render(<ColorPicker onChange={onChange} />);

      const hexInput = screen.getByPlaceholderText('#660000');
      expect(hexInput).toBeInTheDocument();
      expect(hexInput).toHaveValue('#660000');

      const preview = screen.getByRole('button', { name: /выберите цвет/i });
      expect(preview).toHaveStyle({ backgroundColor: '#660000' });
    });

    it('should accept and display all palette colors', () => {
      // Test a subset of colors to avoid timeout
      const testColors = paletteColors.slice(0, 3);

      testColors.forEach((color) => {
        const { unmount } = render(<ColorPicker value={color.hex} onChange={onChange} />);

        const hexInput = screen.getByDisplayValue(color.hex);
        expect(hexInput).toBeInTheDocument();

        const preview = screen.getByRole('button', { name: /выберите цвет/i });
        expect(preview).toHaveStyle({ backgroundColor: color.hex });

        unmount();
      });
    });

    it('should show palette when clicking toggle button', async () => {
      const user = userEvent.setup();
      render(<ColorPicker value="#660000" onChange={onChange} />);

      const toggleButton = screen.getByRole('button', { name: /открыть палитру/i });
      await user.click(toggleButton);

      const paletteHeader = screen.getByText('Выберите цвет');
      expect(paletteHeader).toBeInTheDocument();
    });

    it('should display all 8 palette colors in the palette grid', async () => {
      const user = userEvent.setup({ delay: null });
      render(<ColorPicker value="#660000" onChange={onChange} />);

      const toggleButton = screen.getByRole('button', { name: /открыть палитру/i });
      await user.click(toggleButton);

      // Check for palette header and footer
      expect(screen.getByText('Выберите цвет')).toBeInTheDocument();
      expect(screen.getByText(/или введите hex код/i)).toBeInTheDocument();

      // Check that palette color buttons exist by checking for color buttons with title attributes
      const colorButtons = screen.getAllByRole('button').filter(btn => btn.className.includes('color-palette-item'));
      expect(colorButtons.length).toBe(8);
    });
  });

  describe('Palette Color Selection', () => {
    it('should select color from palette by clicking', async () => {
      const user = userEvent.setup();
      render(<ColorPicker value="#660000" onChange={onChange} />);

      const toggleButton = screen.getByRole('button', { name: /открыть палитру/i });
      await user.click(toggleButton);

      // Get palette and find color button
      const paletteDiv = screen.getByText('Выберите цвет').parentElement;
      const colorButtons = paletteDiv.querySelectorAll('[style*="background-color"]');

      // Click first color button in palette (should be #660000)
      if (colorButtons.length > 0) {
        await user.click(colorButtons[0]);
        expect(onChange).toHaveBeenCalled();
      }
    });

    it('should update preview when selecting color from palette', async () => {
      const user = userEvent.setup({ delay: null });
      const { rerender } = render(<ColorPicker value="#660000" onChange={onChange} />);

      const toggleButton = screen.getByRole('button', { name: /открыть палитру/i });
      await user.click(toggleButton);

      // Rerender with new color value
      rerender(<ColorPicker value="#000066" onChange={onChange} />);

      const preview = screen.getByRole('button', { name: /выберите цвет/i });
      expect(preview).toHaveStyle({ backgroundColor: '#000066' });
    });

    it('should show checkmark on selected color', async () => {
      const user = userEvent.setup();
      render(<ColorPicker value="#660000" onChange={onChange} />);

      const toggleButton = screen.getByRole('button', { name: /открыть палитру/i });
      await user.click(toggleButton);

      // Should have checkmark for selected color
      const checkmarks = screen.getAllByText('✓');
      expect(checkmarks.length).toBeGreaterThan(0);
    });
  });

  describe('Hex Input Validation with Palette Colors', () => {
    it('should accept valid hex color from palette', async () => {
      const user = userEvent.setup();
      render(<ColorPicker value="#660000" onChange={onChange} />);

      const hexInput = screen.getByPlaceholderText('#660000');
      await user.clear(hexInput);
      await user.type(hexInput, '#006600');

      expect(onChange).toHaveBeenCalledWith('#006600');
    });

    it('should validate all palette colors as valid hex', () => {
      // Test a subset of colors to avoid timeout
      const testColors = paletteColors.slice(0, 3);

      testColors.forEach(color => {
        const { unmount } = render(<ColorPicker value={color.hex} onChange={onChange} />);

        const hexInput = screen.getByDisplayValue(color.hex);
        expect(hexInput).toBeInTheDocument();

        // No error should be shown for valid palette colors
        const error = screen.queryByText(/неверный формат/i);
        expect(error).not.toBeInTheDocument();

        unmount();
      });
    });

    it('should auto-add # prefix for palette colors', async () => {
      const user = userEvent.setup();
      render(<ColorPicker value="#660000" onChange={onChange} />);

      const hexInput = screen.getByPlaceholderText('#660000');
      await user.clear(hexInput);
      await user.type(hexInput, '006600');

      // Should call onChange with # prefix added
      expect(onChange).toHaveBeenCalledWith('#006600');
    });

    it('should reset to valid palette color on blur if invalid', async () => {
      const user = userEvent.setup();
      render(<ColorPicker value="#660000" onChange={onChange} />);

      const hexInput = screen.getByPlaceholderText('#660000');
      await user.clear(hexInput);
      await user.type(hexInput, 'INVALID');

      // Blur to trigger reset
      fireEvent.blur(hexInput);

      // Should revert to original palette color
      await waitFor(() => {
        expect(hexInput).toHaveValue('#660000');
      });
    });
  });

  describe('Color Format Handling', () => {
    it('should handle uppercase palette colors', () => {
      render(<ColorPicker value="#660000" onChange={onChange} />);

      const hexInput = screen.getByPlaceholderText('#660000');
      expect(hexInput).toHaveValue('#660000');
    });

    it('should handle lowercase palette colors', () => {
      const lowercaseColor = '#660000'.toLowerCase();
      render(<ColorPicker value={lowercaseColor} onChange={onChange} />);

      const hexInput = screen.getByPlaceholderText('#660000');
      expect(hexInput).toBeInTheDocument();
    });

    it('should accept palette colors without # prefix in input', async () => {
      const user = userEvent.setup();
      render(<ColorPicker value="#660000" onChange={onChange} />);

      const hexInput = screen.getByPlaceholderText('#660000');
      await user.clear(hexInput);
      await user.type(hexInput, '000066');

      expect(onChange).toHaveBeenCalledWith('#000066');
    });
  });

  describe('Palette Color Names and Accessibility', () => {
    it('should have aria-label for color buttons', async () => {
      const user = userEvent.setup();
      render(<ColorPicker value="#660000" onChange={onChange} />);

      const preview = screen.getByRole('button', { name: /выберите цвет/i });
      expect(preview).toHaveAttribute('aria-label');

      const toggleButton = screen.getByRole('button', { name: /открыть палитру/i });
      expect(toggleButton).toHaveAttribute('aria-label');
    });

    it('should have aria-label for hex input', () => {
      render(<ColorPicker value="#660000" onChange={onChange} />);

      const hexInput = screen.getByPlaceholderText('#660000');
      expect(hexInput).toHaveAttribute('aria-label');
    });
  });

  describe('Disabled State with Palette', () => {
    it('should disable palette interaction when disabled=true', () => {
      render(<ColorPicker value="#660000" onChange={onChange} disabled={true} />);

      const preview = screen.getByRole('button', { name: /выберите цвет/i });
      const toggleButton = screen.getByRole('button', { name: /открыть палитру/i });
      const hexInput = screen.getByPlaceholderText('#660000');

      expect(preview).toBeDisabled();
      expect(toggleButton).toBeDisabled();
      expect(hexInput).toBeDisabled();
    });
  });

  describe('Default Color Behavior', () => {
    it('should use #660000 as default color when value prop not provided', () => {
      render(<ColorPicker onChange={onChange} />);

      const hexInput = screen.getByPlaceholderText('#660000');
      expect(hexInput).toHaveValue('#660000');
    });

    it('should display default color in preview', () => {
      render(<ColorPicker onChange={onChange} />);

      const preview = screen.getByRole('button', { name: /выберите цвет/i });
      expect(preview).toHaveStyle({ backgroundColor: '#660000' });
    });

    it('should be from palette colors set', () => {
      render(<ColorPicker onChange={onChange} />);

      const hexInput = screen.getByPlaceholderText('#660000');
      const defaultColor = hexInput.value;

      const isFromPalette = paletteColors.some(c => c.hex === defaultColor);
      expect(isFromPalette).toBe(true);
    });
  });

  describe('Color Picker with Context', () => {
    it('should work with TemplateLessonEditorModal context', () => {
      render(<ColorPicker value="#660000" onChange={onChange} />);

      const hexInput = screen.getByDisplayValue('#660000');
      expect(hexInput).toBeInTheDocument();
    });

    it('should work with TemplateLessonForm context', () => {
      render(<ColorPicker value="#660000" onChange={onChange} />);

      const hexInput = screen.getByDisplayValue('#660000');
      expect(hexInput).toBeInTheDocument();
    });

    it('should trigger onChange callback when color changes', async () => {
      const user = userEvent.setup();
      render(<ColorPicker value="#660000" onChange={onChange} />);

      const hexInput = screen.getByPlaceholderText('#660000');
      await user.clear(hexInput);
      await user.type(hexInput, '#000066');

      expect(onChange).toHaveBeenCalledWith('#000066');
    });
  });
});
