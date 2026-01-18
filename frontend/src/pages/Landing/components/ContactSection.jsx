import { useState, useRef, useEffect } from "react";
import { FiCheckCircle } from "react-icons/fi";
import { toast } from "react-toastify";
import apiClient from "../../../api/client.js";

const ContactSection = () => {
  const [formData, setFormData] = useState({
    name: "",
    phone: "",
    telegram: "",
    email: "",
  });

  const [errors, setErrors] = useState({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);

  const resetTimerRef = useRef(null);

  const validateField = (name, value) => {
    switch (name) {
      case "name":
        if (!value || value.trim().length < 2) {
          return "Введите ваше имя";
        }
        return "";

      case "phone": {
        const phoneDigits = value.replace(/\D/g, "");
        if (phoneDigits.length < 10) {
          return "Введите корректный номер телефона";
        }
        return "";
      }

      case "telegram":
        if (!value || !value.startsWith("@")) {
          return "Введите корректный Telegram (начинается с @)";
        }
        return "";

      case "email":
        if (value && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
          return "Введите корректный email";
        }
        return "";

      default:
        return "";
    }
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));

    if (errors[name]) {
      setErrors((prev) => ({
        ...prev,
        [name]: "",
      }));
    }
  };

  const handleBlur = (e) => {
    const { name, value } = e.target;
    const error = validateField(name, value);
    if (error) {
      setErrors((prev) => ({
        ...prev,
        [name]: error,
      }));
    }
  };

  const validateForm = () => {
    const newErrors = {};

    newErrors.name = validateField("name", formData.name);
    newErrors.phone = validateField("phone", formData.phone);
    newErrors.telegram = validateField("telegram", formData.telegram);
    if (formData.email) {
      newErrors.email = validateField("email", formData.email);
    }

    Object.keys(newErrors).forEach((key) => {
      if (!newErrors[key]) delete newErrors[key];
    });

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!validateForm()) {
      toast.error("Пожалуйста, исправьте ошибки в форме");
      return;
    }

    setIsSubmitting(true);

    try {
      const payload = {
        name: formData.name.trim(),
        phone: formData.phone.trim(),
        telegram: formData.telegram.trim(),
        email: formData.email ? formData.email.trim() : "",
      };

      const response = await apiClient.post("/trial-requests", payload);

      if (response) {
        setIsSuccess(true);
        toast.success("Заявка успешно отправлена! Скоро с вами свяжусь");

        resetTimerRef.current = setTimeout(() => {
          setFormData({ name: "", phone: "", telegram: "", email: "" });
          setIsSuccess(false);
          resetTimerRef.current = null;
        }, 3000);
      }
    } catch (error) {
      if (error?.response) {
        const errorMessage =
          error.response?.data?.message ||
          error.response?.data?.detail ||
          "Ошибка при отправке заявки. Попробуйте позже";
        toast.error(errorMessage);
      } else if (error?.request) {
        toast.error(
          "Не удалось связаться с сервером. Проверьте интернет-соединение",
        );
      } else {
        toast.error(error?.message || "Произошла ошибка. Попробуйте позже");
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  useEffect(() => {
    return () => {
      if (resetTimerRef.current) {
        clearTimeout(resetTimerRef.current);
      }
    };
  }, []);

  if (isSuccess) {
    return (
      <section className="contact-section" id="contact">
        <div className="container">
          <div className="contact-form-success" data-aos="zoom-in">
            <FiCheckCircle className="success-icon" />
            <h3 className="success-title">Заявка отправлена!</h3>
            <p className="success-message">
              Спасибо за вашу заявку. Я свяжусь с вами в ближайшее время.
            </p>
          </div>
        </div>
      </section>
    );
  }

  return (
    <section className="contact-section" id="contact">
      <div className="container">
        <div
          className="contact-form-wrapper"
          data-aos="fade-up"
          data-aos-delay="50"
        >
          <form className="contact-form" onSubmit={handleSubmit}>
            <div className="form-group">
              <label htmlFor="name" className="form-label">
                Имя
              </label>
              <input
                type="text"
                id="name"
                name="name"
                value={formData.name}
                onChange={handleChange}
                onBlur={handleBlur}
                className={`form-input ${errors.name ? "error" : ""}`}
                placeholder="Как к вам обращаться?"
                disabled={isSubmitting}
                required
              />
              {errors.name && (
                <span className="error-message">{errors.name}</span>
              )}
            </div>

            <div className="form-group">
              <label htmlFor="phone" className="form-label">
                Телефон
              </label>
              <input
                type="tel"
                id="phone"
                name="phone"
                value={formData.phone}
                onChange={handleChange}
                onBlur={handleBlur}
                className={`form-input ${errors.phone ? "error" : ""}`}
                placeholder="(+7 (999) 123-45-67)"
                disabled={isSubmitting}
                required
              />
              {errors.phone && (
                <span className="error-message">{errors.phone}</span>
              )}
            </div>

            <div className="form-group">
              <label htmlFor="telegram" className="form-label">
                Telegram
              </label>
              <input
                type="text"
                id="telegram"
                name="telegram"
                value={formData.telegram}
                onChange={handleChange}
                onBlur={handleBlur}
                className={`form-input ${errors.telegram ? "error" : ""}`}
                placeholder="@username"
                disabled={isSubmitting}
                required
              />
              {errors.telegram && (
                <span className="error-message">{errors.telegram}</span>
              )}
            </div>

            <div className="form-group">
              <label htmlFor="email" className="form-label">
                Email (необязательно)
              </label>
              <input
                type="email"
                id="email"
                name="email"
                value={formData.email}
                onChange={handleChange}
                onBlur={handleBlur}
                className={`form-input ${errors.email ? "error" : ""}`}
                placeholder="example@mail.com"
                disabled={isSubmitting}
              />
              {errors.email && (
                <span className="error-message">{errors.email}</span>
              )}
            </div>

            <button
              type="submit"
              className="submit-button contact-submit"
              disabled={isSubmitting}
            >
              {isSubmitting ? (
                <>
                  <div className="spinner"></div>
                  Отправка...
                </>
              ) : (
                <>
                  Отправить заявку
                </>
              )}
            </button>
          </form>
        </div>
      </div>
    </section>
  );
};

export default ContactSection;
