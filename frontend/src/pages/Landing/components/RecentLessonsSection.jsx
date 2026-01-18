export default function RecentLessonsSection() {
  const today = new Date();
  const yesterday = new Date(today);
  yesterday.setDate(yesterday.getDate() - 1);

  const formatDate = (date) => {
    return date.toLocaleDateString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric'
    });
  };

  const downloadFile = (filename) => {
    try {
      const link = document.createElement('a');
      link.href = `/files/${filename}`;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } catch (error) {
      console.error('Error downloading file:', error);
    }
  };

  return (
    <section className="recent-lessons-section">
      <div className="container">
        <h2 className="section-title">Последние занятия</h2>
        <div className="recent-lessons-grid">
          <div className="lesson-card">
            <div className="lesson-card-header">
              <h3 className="lesson-card-title">Дерево отрезков</h3>
              <span className="lesson-date">Занятие от {formatDate(today)}</span>
            </div>

            <p className="lesson-teacher">Преподаватель: Мирослав Адаменко</p>

            <div className="lesson-section">
              <h4 className="lesson-section-title">Задачи с урока:</h4>
              <div className="task-list">
                <span className="task-link">A</span>
                <span className="task-link">B</span>
                <span className="task-link">C</span>
              </div>
            </div>

            <div className="lesson-section">
              <h4 className="lesson-section-title">ДЗ:</h4>
              <div className="task-list">
                <span className="task-link">D</span>
                <span className="task-link">E</span>
                <span className="task-link">F</span>
              </div>
            </div>

            <div className="attached-files">
              <p className="attached-files-title">Прикреплённые файлы:</p>
              <div className="file-item file-item-clickable" onClick={() => downloadFile('segtree.pdf')}>
                📄 segtree.pdf
              </div>
            </div>
          </div>

          <div className="lesson-card">
            <div className="lesson-card-header">
              <h3 className="lesson-card-title">Сканирующая прямая</h3>
              <span className="lesson-date">Занятие от {formatDate(yesterday)}</span>
            </div>

            <p className="lesson-teacher">Преподаватель: Мирослав Адаменко</p>

            <div className="lesson-section">
              <h4 className="lesson-section-title">Одномерный сканлайн для практики:</h4>
              <div className="task-list">
                <span className="task-link">A</span>
                <span className="task-link">B</span>
                <span className="task-link">C</span>
                <span className="task-link">D</span>
              </div>
            </div>

            <div className="lesson-section">
              <h4 className="lesson-section-title">Задачи с урока:</h4>
              <div className="task-list">
                <span className="task-link">E</span>
                <span className="task-link">F</span>
              </div>
            </div>

            <div className="lesson-section">
              <h4 className="lesson-section-title">Попытаться решить самим:</h4>
              <div className="task-list">
                <span className="task-link">G</span>
                <span className="task-link">H</span>
                <span className="task-link">I</span>
              </div>
            </div>

            <div className="attached-files">
              <p className="attached-files-title">Прикреплённые файлы:</p>
              <div className="file-item file-item-clickable" onClick={() => downloadFile('scanline.pdf')}>
                📄 scanline.pdf
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
