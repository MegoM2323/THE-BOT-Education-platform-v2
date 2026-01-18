import React, { useState } from 'react';
import { toast } from 'react-toastify';

const ContactSection = () => {
  const [formData, setFormData] = useState({
    name: '',
    phone: '',
    telegram: '',
    email: ''
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const validateForm = () => {
    if (formData.name.length < 2) {
      toast.error('Имя должно содержать минимум 2 символа');
      return false;
    }
    if (formData.phone.replace(/\D/g, '').length < 10) {
      toast.error('Телефон должен содержать минимум 10 цифр');
      return false;
    }
    if (!formData.telegram.startsWith('@') || formData.telegram.length < 3) {
      toast.error('Telegram должен начинаться с @ и содержать минимум 3 символа');
      return false;
    }
    return true;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!validateForm()) return;

    setIsSubmitting(true);
    try {
      const response = await fetch('/api/v1/landing/trial-requests', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData)
      });

      if (!response.ok) throw new Error('Failed to submit');

      setIsSuccess(true);
      toast.success('Заявка успешно отправлена! Скоро с вами свяжемся');
      setTimeout(() => {
        setFormData({ name: '', phone: '', telegram: '', email: '' });
        setIsSuccess(false);
      }, 3000);
    } catch (error) {
      toast.error('Ошибка при отправке заявки');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <section className="contact-section" id="contact">
      <div className="container">
        <h2 className="section-title-center">Подать заявку на пробное занятие</h2>
        <p className="section-description">
          Заполните форму и мы свяжемся с вами для записи на бесплатное пробное занятие
        </p>
        <form className="contact-form" onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="name">Имя *</label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleChange}
              placeholder="Ваше имя"
              required
              maxLength={100}
            />
          </div>
          <div className="form-group">
            <label htmlFor="phone">Телефон *</label>
            <input
              type="tel"
              id="phone"
              name="phone"
              value={formData.phone}
              onChange={handleChange}
              placeholder="+7 (999) 123-45-67"
              required
            />
          </div>
          <div className="form-group">
            <label htmlFor="telegram">Telegram *</label>
            <input
              type="text"
              id="telegram"
              name="telegram"
              value={formData.telegram}
              onChange={handleChange}
              placeholder="@username"
              required
              maxLength={50}
            />
          </div>
          <div className="form-group">
            <label htmlFor="email">Email (необязательно)</label>
            <input
              type="email"
              id="email"
              name="email"
              value={formData.email}
              onChange={handleChange}
              placeholder="your@email.com"
            />
          </div>
          <button
            type="submit"
            className="btn btn-primary submit-btn"
            disabled={isSubmitting}
          >
            {isSubmitting ? 'Отправка...' : isSuccess ? 'Отправлено' : 'Отправить заявку'}
          </button>
        </form>
      </div>
    </section>
  );
};

export default ContactSection;
