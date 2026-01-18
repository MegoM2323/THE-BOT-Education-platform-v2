import { useState } from 'react';
import { useSubjects } from '../../hooks/useSubjects.js';
import Button from "../common/Button.jsx";
import Spinner from '../common/Spinner.jsx';
import './SubjectsManager.css';

/**
 * Component for managing teacher's subjects
 * Displays current subjects and allows viewing all available subjects
 */
export const SubjectsManager = () => {
  const { useMySubjects } = useSubjects();
  const { subjects, loading, error, refetch } = useMySubjects();
  const [showAllSubjects, setShowAllSubjects] = useState(false);

  if (loading) {
    return (
      <div className="subjects-manager-loading">
        <Spinner size="sm" />
        <span className="loading-text">Загрузка предметов...</span>
      </div>
    );
  }

  if (error) {
    const errorMessage = error instanceof Error ? error.message : (error?.message || String(error));
    return (
      <div className="subjects-manager-error" role="alert">
        <p className="error-message">
          <strong>Ошибка загрузки предметов:</strong> {errorMessage}
        </p>
        <Button variant="secondary" size="sm" onClick={() => refetch()}>
          Повторить попытку
        </Button>
      </div>
    );
  }

  return (
    <div className="subjects-manager">
      <div className="subjects-header">
        <h3 className="subjects-title">Мои предметы</h3>
        {subjects.length > 0 && (
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowAllSubjects(!showAllSubjects)}
          >
            {showAllSubjects ? 'Скрыть все предметы' : 'Показать все предметы'}
          </Button>
        )}
      </div>

      {subjects.length === 0 ? (
        <div className="subjects-empty">
          <p className="empty-message">У вас пока нет назначенных предметов</p>
          <p className="empty-hint">Обратитесь к администратору для назначения предметов</p>
        </div>
      ) : (
        <div className="subjects-list">
          {subjects.map((subject) => (
            <div key={subject.id} className="subject-card">
              <div className="subject-icon">📚</div>
              <div className="subject-info">
                <h4 className="subject-name">{subject.name}</h4>
                {subject.description && (
                  <p className="subject-description">{subject.description}</p>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {showAllSubjects && <AllSubjectsList />}
    </div>
  );
};

/**
 * Component to display all available subjects (read-only)
 */
const AllSubjectsList = () => {
  const { useAllSubjects } = useSubjects();
  const { subjects, loading, error } = useAllSubjects();

  if (loading) {
    return (
      <div className="all-subjects-loading">
        <Spinner size="sm" />
        <span className="loading-text">Загрузка списка предметов...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="all-subjects-error">
        <p className="error-message">Не удалось загрузить список предметов</p>
      </div>
    );
  }

  return (
    <div className="all-subjects-section">
      <h4 className="all-subjects-title">Все доступные предметы</h4>
      <div className="all-subjects-grid">
        {subjects.map((subject) => (
          <div key={subject.id} className="all-subject-item">
            <span className="all-subject-icon">📖</span>
            <span className="all-subject-name">{subject.name}</span>
          </div>
        ))}
      </div>
      {subjects.length === 0 && (
        <p className="all-subjects-empty">Нет доступных предметов</p>
      )}
    </div>
  );
};

export default SubjectsManager;
