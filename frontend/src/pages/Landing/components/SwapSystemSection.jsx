import { useState } from 'react';

const SwapSystemSection = () => {
  const [hoveredId, setHoveredId] = useState(null);

  const calendarLessons = [
    {
      id: 1,
      subject: 'Дерево отрезков',
      time: '09:00',
      teacher: 'Мирослав Адаменко',
      freeSpots: 1,
      maxSpots: 4,
      color: '#004231',
    },
    {
      id: 2,
      subject: 'Сканирующая прямая',
      time: '12:00',
      teacher: 'Мирослав Адаменко',
      freeSpots: 2,
      maxSpots: 4,
      color: '#004231',
    },
    {
      id: 3,
      subject: 'Семинар к РЭ ВСОШ',
      time: '17:00',
      teacher: 'Мирослав Адаменко',
      freeSpots: 1,
      maxSpots: 4,
      color: '#004231',
    },
    {
      id: 4,
      subject: 'Отжиг и его применения',
      time: '19:30',
      teacher: 'Мирослав Адаменко',
      freeSpots: 1,
      maxSpots: 4,
      color: '#004231',
    },
  ];

  return (
    <section className="swap-system-section" id="swap-system">
      <div className="container">
        <h2 className="section-title">Записи на занятия</h2>

        <div className="swap-content">
          <div className="swap-description">
            <p className="description-text">
              Ваше расписание составляется в строгом соответствии
              с вашим индивидуальным планом, темпом и графиком.
              Если приходится пропустить занятие в своей группе -
              просто отмените запись. Методист предложит вам
              альтернативное занятие по пропущенной теме. Также
              вы можете записываться на любые интересные вам
              занятия в два клика, а наши методисты также будут
              предлагать вам наиболее актуальные и полезные для
              вас занятия. Подобная система дает вам предельную
              пластичность расписания.
            </p>
          </div>

          <div className="calendar-demo">
            {calendarLessons.map((lesson) => (
              <div
                key={lesson.id}
                className="calendar-lesson"
                style={{
                  borderLeftColor: lesson.color,
                  backgroundColor: `${lesson.color}${hoveredId === lesson.id ? '35' : '15'}`,
                }}
                onMouseEnter={() => setHoveredId(lesson.id)}
                onMouseLeave={() => setHoveredId(null)}
              >
                <div className="calendar-lesson-subject">{lesson.subject}</div>
                <div className="calendar-lesson-time">{lesson.time}</div>
                <div className="calendar-lesson-teacher">{lesson.teacher}</div>
                <div className="calendar-lesson-spots">
                  Свободно {lesson.freeSpots} из {lesson.maxSpots} мест
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
};

export default SwapSystemSection;
