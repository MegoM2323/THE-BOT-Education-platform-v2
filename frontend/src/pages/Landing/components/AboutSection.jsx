const AboutSection = () => {
  const demoLessons = [
    {
      id: 1,
      subject: '3-4 человека',
      teacher_name: 'Мини группы',
      spots: 4,
      color: '#004231',
    },
    {
      id: 2,
      subject: '2 часа',
      teacher_name: 'Продолжительность занятия',
      spots: 4,
      color: '#004231',
    },
    {
      id: 3,
      subject: '3400/Занятие',
      teacher_name: 'Цена фиксируется на старте',
      spots: 4,
      color: '#004231',
    },
  ];

  return (
    <section className="about-section">
      <div className="container">
        <div className="about-content">
          <div className="about-text">
            <h2 className="section-title">Кто я?</h2>
            <p className="about-description">
              Меня зовут Мирослав Адаменко. Я - увлеченный преподаватель
              олимпиадного программирования. За своей спиной я имею
              более тысячи часов прочитанных лекций и десятки поездок в
              лагеря. Мой опыт дал мне понимание того, как достичь
              максимальной скорости в подготовке к олимпиадам, а также
              сделать процесс наиболее интересным. Результатом этого
              понимания, любви к своему делу и желания сделать лучшее и
              стало решение создать diploma - первую онлайн-школу с
              подобным форматом обучения.
            </p>
            <ul className="about-achievements">
              <li>Лектор олимпиадных школ МФТИ</li>
              <li>Преподаватель выездных интенсивов ВШЭ</li>
              <li>Многократный призер и победитель олимпиад первых и вторых уровней</li>
              <li>Студент ФКН ВШЭ</li>
            </ul>
          </div>
          <div className="about-image">
            <img src="/images/teacher-photo-new.png" alt="Мирослав Адаменко" />
          </div>
        </div>

        <div className="demo-lessons">
          {demoLessons.map((lesson) => (
            <div
              key={lesson.id}
              className="demo-lesson-card"
              style={{
                borderLeft: `4px solid ${lesson.color}`,
                background: `${lesson.color}08`,
              }}
            >
              <div className="demo-lesson-header">
                <span className="demo-lesson-subject">{lesson.subject}</span>
              </div>
              <span className="demo-lesson-teacher">{lesson.teacher_name}</span>
            </div>
          ))}
        </div>

        <div className="consult-block">
          <p className="consult-text">
            <strong>Все начинается с бесплатной консультации.</strong> На ней мы определим ваш уровень, цели и сроки на их достижение.
            Составим индивидуальный план: темы, которые нужно подтянуть и в которые нужно углубиться, лекции,
            которые стоит прослушать и семинары, которые стоит посетить.
          </p>
        </div>
      </div>
    </section>
  );
};

export default AboutSection;
