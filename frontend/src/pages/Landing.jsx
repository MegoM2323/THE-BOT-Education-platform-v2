import { useState, useRef, useEffect } from "react";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "../components/ui/card";
import { BookOpen, Users, Target, MessageCircle, TrendingUp, Shield, CheckCircle, Loader2, Send } from "lucide-react";
import apiClient from "../api/client.js";

// Компонент формы заявки
const ApplicationForm = () => {
  const [formData, setFormData] = useState({
    name: "",
    phone: "",
    telegram: "",
    email: "",
  });
  const [errors, setErrors] = useState({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);
  const [submitError, setSubmitError] = useState("");
  const resetTimerRef = useRef(null);

  const validateField = (name, value) => {
    switch (name) {
      case "name":
        if (!value || value.trim().length < 2) {
          return "Введите ваше имя (минимум 2 символа)";
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
        if (!value || !value.startsWith("@") || value.length < 2) {
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
    setSubmitError("");
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
      setSubmitError("Пожалуйста, исправьте ошибки в форме");
      return;
    }

    setIsSubmitting(true);
    setSubmitError("");

    try {
      const payload = {
        name: formData.name.trim(),
        phone: formData.phone.trim(),
        telegram: formData.telegram.trim(),
        email: formData.email ? formData.email.trim() : "",
      };

      await apiClient.post("/trial-requests", payload);

      setIsSuccess(true);

      resetTimerRef.current = setTimeout(() => {
        setFormData({ name: "", phone: "", telegram: "", email: "" });
        setIsSuccess(false);
        resetTimerRef.current = null;
      }, 5000);
    } catch (error) {
      const errorMessage =
        error?.data?.error?.message ||
        error?.message ||
        "Ошибка при отправке заявки. Попробуйте позже";
      setSubmitError(errorMessage);
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
      <Card className="max-w-md mx-auto border-success/30 bg-success/5">
        <CardContent className="pt-8 pb-8 text-center space-y-4">
          <div className="w-16 h-16 gradient-secondary rounded-full flex items-center justify-center mx-auto">
            <CheckCircle className="w-8 h-8 text-secondary-foreground" />
          </div>
          <h3 className="text-2xl font-bold text-foreground">Заявка отправлена!</h3>
          <p className="text-muted-foreground">
            Спасибо за вашу заявку. Я свяжусь с вами в ближайшее время.
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="max-w-md mx-auto">
      <CardHeader className="text-center">
        <CardTitle className="text-2xl">Подать заявку</CardTitle>
        <CardDescription>
          Заполните форму, и я свяжусь с вами для обсуждения деталей
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label htmlFor="name" className="text-sm font-medium text-foreground">
              Имя <span className="text-destructive">*</span>
            </label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleChange}
              onBlur={handleBlur}
              className={`w-full h-10 px-3 rounded-md border bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring transition-colors ${
                errors.name ? "border-destructive focus:ring-destructive" : "border-input"
              }`}
              placeholder="Как к вам обращаться?"
              disabled={isSubmitting}
            />
            {errors.name && (
              <p className="text-sm text-destructive">{errors.name}</p>
            )}
          </div>

          <div className="space-y-2">
            <label htmlFor="phone" className="text-sm font-medium text-foreground">
              Телефон <span className="text-destructive">*</span>
            </label>
            <input
              type="tel"
              id="phone"
              name="phone"
              value={formData.phone}
              onChange={handleChange}
              onBlur={handleBlur}
              className={`w-full h-10 px-3 rounded-md border bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring transition-colors ${
                errors.phone ? "border-destructive focus:ring-destructive" : "border-input"
              }`}
              placeholder="+7 (999) 123-45-67"
              disabled={isSubmitting}
            />
            {errors.phone && (
              <p className="text-sm text-destructive">{errors.phone}</p>
            )}
          </div>

          <div className="space-y-2">
            <label htmlFor="telegram" className="text-sm font-medium text-foreground">
              Telegram <span className="text-destructive">*</span>
            </label>
            <input
              type="text"
              id="telegram"
              name="telegram"
              value={formData.telegram}
              onChange={handleChange}
              onBlur={handleBlur}
              className={`w-full h-10 px-3 rounded-md border bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring transition-colors ${
                errors.telegram ? "border-destructive focus:ring-destructive" : "border-input"
              }`}
              placeholder="@username"
              disabled={isSubmitting}
            />
            {errors.telegram && (
              <p className="text-sm text-destructive">{errors.telegram}</p>
            )}
          </div>

          <div className="space-y-2">
            <label htmlFor="email" className="text-sm font-medium text-foreground">
              Email <span className="text-muted-foreground text-xs">(необязательно)</span>
            </label>
            <input
              type="email"
              id="email"
              name="email"
              value={formData.email}
              onChange={handleChange}
              onBlur={handleBlur}
              className={`w-full h-10 px-3 rounded-md border bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring transition-colors ${
                errors.email ? "border-destructive focus:ring-destructive" : "border-input"
              }`}
              placeholder="example@mail.com"
              disabled={isSubmitting}
            />
            {errors.email && (
              <p className="text-sm text-destructive">{errors.email}</p>
            )}
          </div>

          {submitError && (
            <div className="p-3 rounded-md bg-destructive/10 border border-destructive/30">
              <p className="text-sm text-destructive">{submitError}</p>
            </div>
          )}

          <Button
            type="submit"
            className="w-full gradient-primary shadow-glow hover:opacity-90 transition-opacity"
            size="lg"
            disabled={isSubmitting}
          >
            {isSubmitting ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Отправка...
              </>
            ) : (
              <>
                <Send className="w-4 h-4 mr-2" />
                Отправить заявку
              </>
            )}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
};

const Landing = () => {
  return (
    <div className="min-h-screen">
      {/* Hero Section */}
      <section className="container mx-auto px-4 py-20 md:py-32">
        <div className="max-w-4xl mx-auto text-center space-y-8 animate-slide-up">
          <h1 className="text-gradient-thebot">
            Образовательная платформа нового поколения
          </h1>
          <p className="text-xl md:text-2xl text-muted-foreground max-w-2xl mx-auto">
            Персонализированное обучение с трекингом прогресса, прямой связью с преподавателями и тьюторами
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <a href="#apply">
              <Button
                type="button"
                size="lg"
                className="gradient-primary shadow-glow hover:opacity-90 transition-opacity"
              >
                Подать заявку
              </Button>
            </a>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section id="features" className="py-20 bg-muted/30">
        <div className="container mx-auto px-4">
          <h2 className="text-center mb-12">Возможности платформы</h2>
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {features.map((feature, index) => (
              <Card key={index} className="p-6 hover:shadow-lg transition-shadow">
                <div className="w-12 h-12 gradient-primary rounded-lg flex items-center justify-center mb-4">
                  <feature.icon className="w-6 h-6 text-primary-foreground" />
                </div>
                <h3 className="text-xl mb-2">{feature.title}</h3>
                <p className="text-muted-foreground">{feature.description}</p>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Roles Section */}
      <section id="roles" className="py-20">
        <div className="container mx-auto px-4">
          <h2 className="text-center mb-12">Для всех участников образовательного процесса</h2>
          <div className="grid md:grid-cols-2 gap-8 max-w-5xl mx-auto">
            {roles.map((role, index) => (
              <Card key={index} className="p-8 hover:border-primary transition-colors">
                <div className="flex items-start gap-4">
                  <div className="w-14 h-14 gradient-secondary rounded-xl flex items-center justify-center flex-shrink-0">
                    <role.icon className="w-7 h-7 text-secondary-foreground" />
                  </div>
                  <div>
                    <h3 className="text-2xl mb-3">{role.title}</h3>
                    <ul className="space-y-2 text-muted-foreground">
                      {role.features.map((feature, idx) => (
                        <li key={idx} className="flex items-start gap-2">
                          <span className="text-primary mt-1">•</span>
                          <span>{feature}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Application Form Section */}
      <section id="apply" className="py-20 bg-muted/30">
        <div className="container mx-auto px-4">
          <div className="text-center mb-12">
            <h2 className="text-3xl font-bold mb-4">Оставить заявку на обучение</h2>
            <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
              Заполните форму заявки, и я свяжусь с вами в ближайшее время.
              После одобрения заявки вы получите данные для входа в систему через Telegram.
            </p>
          </div>
          <ApplicationForm />
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t py-8 bg-card/50">
        <div className="container mx-auto px-4">
          <div className="flex flex-col md:flex-row justify-between items-center gap-4">
            <div className="flex items-center gap-2">
              <div className="w-10 h-10 flex items-center justify-center flex-shrink-0">
                <svg width="40" height="40" viewBox="0 0 40 40" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <defs>
                    <linearGradient id="footerLogoGradient" x1="0%" y1="0%" x2="100%" y2="0%">
                      <stop offset="0%" stopColor="#8B5CF6" />
                      <stop offset="100%" stopColor="#A78BFA" />
                    </linearGradient>
                  </defs>
                  <rect width="40" height="40" rx="8" fill="url(#footerLogoGradient)" />
                  <path d="M12 14C12 13.4477 12.4477 13 13 13H20C20.5523 13 21 13.4477 21 14V26C21 26.5523 20.5523 27 20 27H13C12.4477 27 12 26.5523 12 26V14Z" stroke="white" strokeWidth="2" fill="none" strokeLinecap="round" strokeLinejoin="round"/>
                  <path d="M19 14C19 13.4477 19.4477 13 20 13H27C27.5523 13 28 13.4477 28 14V26C28 26.5523 27.5523 27 27 27H20C19.4477 27 19 26.5523 19 26V14Z" stroke="white" strokeWidth="2" fill="none" strokeLinecap="round" strokeLinejoin="round"/>
                  <path d="M20 20L20 27" stroke="white" strokeWidth="2" strokeLinecap="round"/>
                  <circle cx="20" cy="24" r="1.5" fill="white"/>
                </svg>
              </div>
              <span className="font-bold text-gradient-thebot">THE BOT</span>
            </div>
            <p className="text-sm text-muted-foreground">
              © 2024 THE BOT. Все права защищены.
            </p>
          </div>
        </div>
      </footer>
    </div>
  );
};

const features = [
  {
    icon: Target,
    title: "Трекинг прогресса",
    description: "Визуализация достижений и движения к образовательной цели в реальном времени"
  },
  {
    icon: MessageCircle,
    title: "Прямая связь",
    description: "Общение с преподавателями и тьюторами напрямую через платформу"
  },
  {
    icon: TrendingUp,
    title: "Персональные отчёты",
    description: "Подробная аналитика по каждому ученику для родителей и тьюторов"
  },
  {
    icon: BookOpen,
    title: "Материалы и задания",
    description: "Удобная система публикации и проверки домашних заданий"
  },
  {
    icon: Users,
    title: "Многоролевая система",
    description: "Отдельные кабинеты для учеников, родителей, преподавателей и тьюторов"
  },
  {
    icon: Shield,
    title: "Безопасность данных",
    description: "Надёжная защита личной информации и образовательных материалов"
  }
];

const roles = [
  {
    icon: Users,
    title: "Ученики",
    features: [
      "Доступ к образовательным материалам",
      "Трекер прогресса с геймификацией",
      "Возможность задавать вопросы преподавателям",
      "Уведомления о проверке домашних заданий"
    ]
  },
  {
    icon: Users,
    title: "Родители",
    features: [
      "Отчёты о прогрессе детей",
      "Связь с тьютором",
      "Доступ к образовательным целям",
      "Уведомления о важных событиях"
    ]
  },
  {
    icon: Users,
    title: "Преподаватели",
    features: [
      "Создание и публикация материалов",
      "Проверка домашних заданий",
      "Формирование отчётов по предмету",
      "Ответы на вопросы учеников"
    ]
  },
  {
    icon: Users,
    title: "Тьюторы",
    features: [
      "Управление группой учеников",
      "Создание финальных отчётов",
      "Координация с преподавателями",
      "Консультации родителей"
    ]
  }
];

export default Landing;
