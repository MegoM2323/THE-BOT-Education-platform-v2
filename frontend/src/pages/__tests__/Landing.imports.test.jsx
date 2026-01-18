import { describe, it, expect } from 'vitest';
import HeroSection from '../Landing/components/HeroSection';
import FeaturesSection from '../Landing/components/FeaturesSection';
import RolesSection from '../Landing/components/RolesSection';
import CallToActionSection from '../Landing/components/CallToActionSection';
import FooterSection from '../Landing/components/FooterSection';
import Landing from '../Landing';

describe('Landing Page - Component Imports', () => {
  it('should import HeroSection successfully', () => {
    expect(HeroSection).toBeDefined();
    expect(typeof HeroSection).toBe('function');
  });

  it('should import FeaturesSection successfully', () => {
    expect(FeaturesSection).toBeDefined();
    expect(typeof FeaturesSection).toBe('function');
  });

  it('should import RolesSection successfully', () => {
    expect(RolesSection).toBeDefined();
    expect(typeof RolesSection).toBe('function');
  });

  it('should import CallToActionSection successfully', () => {
    expect(CallToActionSection).toBeDefined();
    expect(typeof CallToActionSection).toBe('function');
  });

  it('should import FooterSection successfully', () => {
    expect(FooterSection).toBeDefined();
    expect(typeof FooterSection).toBe('function');
  });

  it('should import Landing page successfully', () => {
    expect(Landing).toBeDefined();
    expect(typeof Landing).toBe('function');
  });

  it('Landing can be used as default export', () => {
    expect(Landing).toBeDefined();
  });

  it('all components should be callable', () => {
    expect(() => HeroSection()).not.toThrow();
    expect(() => FeaturesSection()).not.toThrow();
    expect(() => RolesSection()).not.toThrow();
    expect(() => CallToActionSection()).not.toThrow();
    expect(() => FooterSection()).not.toThrow();
  });

  it('components return React elements', () => {
    const heroElement = HeroSection();
    const featuresElement = FeaturesSection();
    const rolesElement = RolesSection();
    const ctaElement = CallToActionSection();
    const footerElement = FooterSection();

    expect(heroElement).not.toBeNull();
    expect(featuresElement).not.toBeNull();
    expect(rolesElement).not.toBeNull();
    expect(ctaElement).not.toBeNull();
    expect(footerElement).not.toBeNull();
  });
});
