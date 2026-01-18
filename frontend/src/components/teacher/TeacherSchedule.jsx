import { useState, useEffect } from 'react';
import { useLessons } from '../../hooks/useLessons.js';
import Card from '../common/Card.jsx';
import Button from "../common/Button.jsx";
import Spinner from '../common/Spinner.jsx';
import StudentsList from './StudentsList.jsx';
import { formatDateWithDay, formatTime } from '../../utils/dateFormat.js';
import './TeacherSchedule.css';

export const TeacherSchedule = () => {
  const { myLessons, myLessonsLoading, myLessonsError, fetchMyLessons } = useLessons();
  const [dateFilter, setDateFilter] = useState('upcoming');
  const [selectedLesson, setSelectedLesson] = useState(null);

  useEffect(() => {
    fetchMyLessons();
  }, [fetchMyLessons]);

  const groupedLessons = myLessons.reduce((groups, lesson) => {
    const lessonDate = new Date(lesson.start_time);
    const now = new Date();
    const dateKey = lessonDate.toISOString().split('T')[0];

    // Применить фильтр по дате
    if (dateFilter === 'upcoming' && lessonDate < now) return groups;
    if (dateFilter === 'past' && lessonDate >= now) return groups;

    if (!groups[dateKey]) {
      groups[dateKey] = [];
    }
    groups[dateKey].push(lesson);
    return groups;
  }, {});

  const sortedDates = Object.keys(groupedLessons).sort((a, b) => {
    if (dateFilter === 'upcoming') {
      return new Date(a) - new Date(b);
    } else {
      return new Date(b) - new Date(a);
    }
  });

  if (myLessonsLoading) {
    return (
      <div className="teacher-schedule-loading">
        <Spinner />
      </div>
    );
  }

  if (myLessonsError) {
    return (
      <div className="teacher-schedule" data-testid="teacher-schedule">
        <div className="teacher-schedule-header">
          <h2 className="teacher-schedule-title">Мое расписание</h2>
        </div>
        <div className="teacher-schedule-error">
          <p>Ошибка загрузки расписания: {myLessonsError?.message || myLessonsError}</p>
          <Button variant="primary" onClick={fetchMyLessons}>
            Повторить попытку
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="teacher-schedule" data-testid="teacher-schedule">
      <div className="teacher-schedule-header">
        <h2 className="teacher-schedule-title">Мое расписание</h2>
        <div className="teacher-schedule-filters" data-testid="date-filter">
          <Button
            variant={dateFilter === 'upcoming' ? 'primary' : 'secondary'}
            onClick={() => setDateFilter('upcoming')}
            data-testid="sort-date"
          >
            Предстоящие
          </Button>
          <Button
            variant={dateFilter === 'past' ? 'primary' : 'secondary'}
            onClick={() => setDateFilter('past')}
          >
            Прошедшие
          </Button>
        </div>
      </div>

      {sortedDates.length === 0 ? (
        <div className="teacher-schedule-empty">
          <p>
            {dateFilter === 'upcoming'
              ? 'У вас нет предстоящих занятий'
              : 'У вас нет прошедших занятий'}
          </p>
        </div>
      ) : (
        <div className="teacher-schedule-groups">
          {sortedDates.map((dateKey) => (
            <div key={dateKey} className="teacher-schedule-group">
              <h3 className="teacher-schedule-date">
                {formatDateWithDay(groupedLessons[dateKey][0].start_time)}
              </h3>
              <div className="teacher-schedule-lessons">
                {groupedLessons[dateKey].map((lesson) => {
                  const spotsBooked = lesson.current_students || 0;
                  const spotsTotal = lesson.max_students;
                  const fillPercentage = (spotsBooked / spotsTotal) * 100;

                  return (
                    <Card key={lesson.id} className="teacher-lesson-card" data-testid="my-lesson">
                      <div className="teacher-lesson-header">
                        <span className="teacher-lesson-time">
                          {formatTime(lesson.start_time)} - {formatTime(lesson.end_time)}
                        </span>
                      </div>

                      <div className="teacher-lesson-details">
                        {lesson.description && (
                          <p className="teacher-lesson-description">{lesson.description}</p>
                        )}

                        <div className="teacher-lesson-capacity">
                          <div className="capacity-label">
                            Записано: {spotsBooked} из {spotsTotal}
                          </div>
                          <div className="capacity-bar">
                            <div
                              className={`capacity-fill ${
                                fillPercentage === 100
                                  ? 'full'
                                  : fillPercentage >= 75
                                  ? 'high'
                                  : fillPercentage >= 50
                                  ? 'medium'
                                  : 'low'
                              }`}
                              style={{ width: `${fillPercentage}%` }}
                            />
                          </div>
                        </div>
                      </div>

                      <div className="teacher-lesson-actions">
                        <Button
                          onClick={() => setSelectedLesson(lesson)}
                          variant="secondary"
                          size="small"
                          data-testid="view-details"
                        >
                          Посмотреть учеников
                        </Button>
                      </div>
                    </Card>
                  );
                })}
              </div>
            </div>
          ))}
        </div>
      )}

      <StudentsList
        isOpen={!!selectedLesson}
        onClose={() => setSelectedLesson(null)}
        lesson={selectedLesson}
      />
    </div>
  );
};

export default TeacherSchedule;
