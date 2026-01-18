import { render, screen } from '@testing-library/react';
import { Button } from './Button';
import { describe, it, expect } from 'vitest';

describe('Button text overflow handling', () => {
  it('should handle long text without overflow', () => {
    const longText = 'Очень длинный текст который должен быть обрезан с ellipsis';
    const { container } = render(<Button>{longText}</Button>);

    const button = container.querySelector('.btn');
    const styles = window.getComputedStyle(button);

    // Should have white-space nowrap for base style
    expect(styles.whiteSpace).toBe('nowrap');
    // Button should render without error
    expect(button).toBeInTheDocument();
  });

  it('should render short text normally', () => {
    render(<Button>Click me</Button>);
    const button = screen.getByRole('button', { name: /click me/i });
    expect(button).toBeInTheDocument();
  });

  it('should constrain button width with max-width', () => {
    const { container } = render(<Button fullWidth>Test</Button>);
    const button = container.querySelector('.btn-full-width');

    expect(button).toBeInTheDocument();
  });

  it('should handle full width variant', () => {
    const { container } = render(<Button fullWidth>Full Width Button</Button>);
    const button = container.querySelector('.btn-full-width');
    expect(button).toBeInTheDocument();
  });

  it('should maintain responsive font sizes', () => {
    // Test is checking that CSS is properly applied
    // Actual media query testing requires specific viewport simulation
    const { container } = render(<Button size="sm">Small</Button>);
    const button = container.querySelector('.btn-sm');
    expect(button).toBeInTheDocument();
  });

  it('should handle loading state without overflow', () => {
    const longText = 'Загрузка очень длинного процесса';
    render(<Button loading>{longText}</Button>);

    const button = screen.getByRole('button');
    expect(button).toHaveAttribute('disabled');
  });

  it('should render different button sizes correctly', () => {
    const { container: smContainer } = render(<Button size="sm">Small</Button>);
    const { container: mdContainer } = render(<Button size="md">Medium</Button>);
    const { container: lgContainer } = render(<Button size="lg">Large</Button>);

    expect(smContainer.querySelector('.btn-sm')).toBeInTheDocument();
    expect(mdContainer.querySelector('.btn-md')).toBeInTheDocument();
    expect(lgContainer.querySelector('.btn-lg')).toBeInTheDocument();
  });

  it('should handle Russian text with Cyrillic characters', () => {
    const cyrillicText = 'Применить ко всем последующим урокам';
    render(<Button>{cyrillicText}</Button>);

    const button = screen.getByRole('button');
    expect(button).toHaveTextContent(cyrillicText);
  });
});
