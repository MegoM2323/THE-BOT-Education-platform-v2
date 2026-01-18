import { render, screen } from '@testing-library/react';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { LessonCard } from "../LessonCard.jsx";
import '@testing-library/jest-dom';

describe('LessonCard Color Implementation', () => {
  let mockOnBook;

  beforeEach(() => {
    mockOnBook = vi.fn();
  });

  const createMockLesson = (overrides = {}) => ({
    id: 1,
    subject: 'Mathematics',
    teacher_name: 'John Doe',
    start_time: new Date(Date.now() + 48 * 60 * 60 * 1000).toISOString(),
    end_time: new Date(Date.now() + 49 * 60 * 60 * 1000).toISOString(),
    max_students: 10,
    current_students: 5,
    description: 'Basic algebra lessons',
    color: '#ff6b6b',
    ...overrides,
  });

  it('applies lesson.color to borderLeft style', () => {
    const mockLesson = createMockLesson();
    const { container } = render(
      <LessonCard
        lesson={mockLesson}
        onBook={mockOnBook}
      />
    );

    const card = container.querySelector('.lesson-card');

    // Check that the color is applied
    expect(card).toHaveStyle(`border-left: 4px solid ${mockLesson.color}`);
  });

  it('applies lesson.color to background with opacity', () => {
    const mockLesson = createMockLesson();
    const { container } = render(
      <LessonCard
        lesson={mockLesson}
        onBook={mockOnBook}
      />
    );

    const card = container.querySelector('.lesson-card');

    // Background should use lesson color with 08 (4% opacity)
    expect(card).toHaveStyle(`background: ${mockLesson.color}08`);
  });

  it('uses default color when lesson.color is undefined', () => {
    const mockLesson = createMockLesson({ color: undefined });

    const { container } = render(
      <LessonCard
        lesson={mockLesson}
        onBook={mockOnBook}
      />
    );

    const card = container.querySelector('.lesson-card');

    // Should use default dark red color
    expect(card).toHaveStyle('border-left: 4px solid #660000');
    expect(card).toHaveStyle('background: #66000008');
  });

  it('uses default color when lesson.color is null', () => {
    const mockLesson = createMockLesson({ color: null });

    const { container } = render(
      <LessonCard
        lesson={mockLesson}
        onBook={mockOnBook}
      />
    );

    const card = container.querySelector('.lesson-card');

    // Should use default dark red color
    expect(card).toHaveStyle('border-left: 4px solid #660000');
    expect(card).toHaveStyle('background: #66000008');
  });

  it('uses default color when lesson.color is empty string', () => {
    const mockLesson = createMockLesson({ color: '' });

    const { container } = render(
      <LessonCard
        lesson={mockLesson}
        onBook={mockOnBook}
      />
    );

    const card = container.querySelector('.lesson-card');

    // Should use default dark red color
    expect(card).toHaveStyle('border-left: 4px solid #660000');
    expect(card).toHaveStyle('background: #66000008');
  });

  it('applies custom hex colors correctly', () => {
    const testColors = ['#00ff00', '#0000ff', '#ffff00', '#ff00ff'];

    testColors.forEach((color) => {
      const mockLesson = createMockLesson({ color });
      const { container, unmount } = render(
        <LessonCard
          lesson={mockLesson}
          onBook={mockOnBook}
        />
      );

      const card = container.querySelector('.lesson-card');

      expect(card).toHaveStyle(`border-left: 4px solid ${color}`);
      expect(card).toHaveStyle(`background: ${color}08`);

      unmount();
    });
  });

  it('CSS allows inline border-left style to override default', () => {
    const mockLesson = createMockLesson();
    const { container } = render(
      <LessonCard
        lesson={mockLesson}
        onBook={mockOnBook}
      />
    );

    const card = container.querySelector('.lesson-card');
    const computedStyle = window.getComputedStyle(card);

    // Verify that the inline style is applied (not overridden by CSS)
    // Browsers convert hex colors to rgb, so we check for the color in the computed style
    const borderLeft = computedStyle.borderLeft;
    expect(borderLeft).toMatch(/rgb\(255,\s*107,\s*107\)|#ff6b6b/i);
  });

  it('CSS allows background to be set via inline style', () => {
    const mockLesson = createMockLesson();
    const { container } = render(
      <LessonCard
        lesson={mockLesson}
        onBook={mockOnBook}
      />
    );

    const card = container.querySelector('.lesson-card');

    // Verify background style is applied
    expect(card).toHaveAttribute('style');
    const styleAttr = card.getAttribute('style');
    expect(styleAttr).toContain('background');
  });

  it('maintains color on booked card state', () => {
    const mockLesson = createMockLesson();
    const { container } = render(
      <LessonCard
        lesson={mockLesson}
        onBook={mockOnBook}
        isBooked={true}
      />
    );

    const card = container.querySelector('.lesson-card');

    // Color should still be applied even when booked
    expect(card).toHaveStyle(`border-left: 4px solid ${mockLesson.color}`);
    expect(card).toHaveStyle(`background: ${mockLesson.color}15`);
  });

  it('Card component receives style prop correctly', () => {
    const mockLesson = createMockLesson();
    const { container } = render(
      <LessonCard
        lesson={mockLesson}
        onBook={mockOnBook}
      />
    );

    const card = container.querySelector('.lesson-card');

    // Verify style attribute exists
    expect(card).toHaveAttribute('style');

    const styleAttr = card.getAttribute('style');
    // Check for the color format - browsers convert hex to rgb
    expect(styleAttr).toMatch(/border-left:\s*4px\s+solid\s+rgb\(255,\s*107,\s*107\)|border-left:\s*4px\s+solid\s+#ff6b6b/i);
    expect(styleAttr).toMatch(/background:\s*rgba\(255,\s*107,\s*107|background:\s*#ff6b6b08/i);
  });

  it('different lessons have different colors applied', () => {
    const lesson1 = createMockLesson({ id: 1, color: '#ff0000' });
    const lesson2 = createMockLesson({ id: 2, color: '#00ff00' });

    const { container: container1, unmount: unmount1 } = render(
      <LessonCard lesson={lesson1} onBook={mockOnBook} />
    );

    const card1 = container1.querySelector('.lesson-card');
    expect(card1).toHaveStyle('border-left: 4px solid #ff0000');

    unmount1();

    const { container: container2, unmount: unmount2 } = render(
      <LessonCard lesson={lesson2} onBook={mockOnBook} />
    );

    const card2 = container2.querySelector('.lesson-card');
    expect(card2).toHaveStyle('border-left: 4px solid #00ff00');

    unmount2();
  });

  it('color persists during animation states', () => {
    const mockLesson = createMockLesson();
    const { container, rerender } = render(
      <LessonCard
        lesson={mockLesson}
        onBook={mockOnBook}
        isOperating={false}
      />
    );

    let card = container.querySelector('.lesson-card');
    expect(card).toHaveStyle(`border-left: 4px solid ${mockLesson.color}`);

    // Rerender with operating state
    rerender(
      <LessonCard
        lesson={mockLesson}
        onBook={mockOnBook}
        isOperating={true}
      />
    );

    card = container.querySelector('.lesson-card');
    expect(card).toHaveStyle(`border-left: 4px solid ${mockLesson.color}`);
  });

  it('rgb color format is supported', () => {
    const mockLesson = createMockLesson({ color: 'rgb(255, 107, 107)' });

    const { container } = render(
      <LessonCard
        lesson={mockLesson}
        onBook={mockOnBook}
      />
    );

    const card = container.querySelector('.lesson-card');

    // Should apply rgb color
    expect(card).toHaveStyle(`border-left: 4px solid ${mockLesson.color}`);
  });

  it('rgba color format with opacity is supported', () => {
    const mockLesson = createMockLesson({ color: 'rgba(255, 107, 107, 0.8)' });

    const { container } = render(
      <LessonCard
        lesson={mockLesson}
        onBook={mockOnBook}
      />
    );

    const card = container.querySelector('.lesson-card');

    // Should apply rgba color
    expect(card).toHaveStyle(`border-left: 4px solid ${mockLesson.color}`);
  });

  it('hsl color format is supported', () => {
    const mockLesson = createMockLesson({ color: 'hsl(0, 100%, 71%)' });

    const { container } = render(
      <LessonCard
        lesson={mockLesson}
        onBook={mockOnBook}
      />
    );

    const card = container.querySelector('.lesson-card');

    // Should apply hsl color
    expect(card).toHaveStyle(`border-left: 4px solid ${mockLesson.color}`);
  });
});
