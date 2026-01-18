const LessonTypesSection = () => {
  return (
    <section className="lesson-types-section" id="lesson-types">
      <div className="container">
        <h2 className="section-title">
          Какие бывают занятия и как они проходят?
        </h2>

        <div className="lesson-types-content">
          <div className="lesson-types-text">
            <div className="lesson-type">
              <h3>Лекции</h3>
              <p>
                Разбираем каждую из тем предельно глубоко,
                покрывая все аспекты. После понимания материала
                отрабатываем навыки реализации - преподаватель
                помогает написать идеальный и рабочий код.
              </p>
            </div>

            <div className="lesson-type">
              <h3>Семинары</h3>
              <p>
                Решаем реальные задачи по пройденным темам,
                увеличивая насмотренность и учась видеть идеи в
                реальных задачах. Материал не забывается, а
                фиксируется, получая регулярное подкрепление.
              </p>
            </div>
          </div>

          <div className="lesson-types-image">
            <img src="/images/prep-screen.png" alt="Скриншот занятия с преподавателем" />
            <span className="image-caption">отработка практики с преподавателем</span>
          </div>
        </div>
      </div>
    </section>
  );
};

export default LessonTypesSection;
