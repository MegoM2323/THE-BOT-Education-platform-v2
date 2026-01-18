import { Button } from "../components/ui/button";
import { Card } from "../components/ui/card";
import { BookOpen, Users, Target, MessageCircle, TrendingUp, Shield } from "lucide-react";
import { Link } from "react-router-dom";

const Landing = () => {
  return (
    <div className="min-h-screen">
      {/* Hero Section */}
      <section className="container mx-auto px-4 py-20 md:py-32">
        <div className="max-w-4xl mx-auto text-center space-y-8 animate-slide-up">
          <h1 className="bg-gradient-to-r from-primary via-primary-glow to-accent bg-clip-text text-transparent">
            Образовательная платформа нового поколения
          </h1>
          <p className="text-xl md:text-2xl text-muted-foreground max-w-2xl mx-auto">
            Персонализированное обучение с трекингом прогресса, прямой связью с преподавателями и тьюторами
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link to="/login">
              <Button type="button" size="lg" className="gradient-primary shadow-glow hover:opacity-90 transition-opacity">
                Подать заявку
              </Button>
            </Link>
            <Link to="/login">
              <Button type="button" size="lg" variant="outline">
                Личный кабинет
              </Button>
            </Link>
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

      {/* Application CTA Section */}
      <section id="apply" className="py-20 bg-muted/30">
        <div className="container mx-auto px-4">
          <div className="max-w-2xl mx-auto text-center">
            <h2 className="text-3xl font-bold mb-4">Подать заявку на обучение</h2>
            <p className="text-lg text-muted-foreground mb-8">
              Заполните форму заявки, и наши специалисты свяжутся с вами в ближайшее время.
              После одобрения заявки вы получите данные для входа в систему через Telegram.
            </p>
            <Link to="/login">
              <Button type="button" size="lg" className="gradient-primary shadow-glow">
                Подать заявку
              </Button>
            </Link>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t py-8 bg-card/50">
        <div className="container mx-auto px-4">
          <div className="flex flex-col md:flex-row justify-between items-center gap-4">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 gradient-primary rounded-lg flex items-center justify-center">
                <BookOpen className="w-5 h-5 text-primary-foreground" />
              </div>
              <span className="font-bold">THE BOT</span>
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
