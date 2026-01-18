/**
 * Preview component showing a sample of affected lessons
 * Used in bulk edit confirmation dialogs
 */
import { formatTime } from '../../utils/dateFormat.js';
import './AffectedLessonsPreview.css';

export const AffectedLessonsPreview = ({ lessons, totalCount }) => {
  if (!lessons || lessons.length === 0) {
    return (
      <div className="affected-lessons-preview">
        <p className="preview-empty">No lessons to preview</p>
      </div>
    );
  }

  const previewLimit = 5;
  const lessonsToShow = lessons.slice(0, previewLimit);
  const remaining = totalCount > previewLimit ? totalCount - previewLimit : 0;

  return (
    <div className="affected-lessons-preview">
      <h4 className="preview-title">Preview of Affected Lessons:</h4>

      <table className="preview-table">
        <thead>
          <tr>
            <th>Date</th>
            <th>Time</th>
            <th>Teacher</th>
            <th>Students</th>
          </tr>
        </thead>
        <tbody>
          {lessonsToShow.map((lesson) => {
            const lessonDate = new Date(lesson.start_time);
            return (
              <tr key={lesson.id}>
                <td>
                  {lessonDate.toLocaleDateString('ru-RU', {
                    month: 'short',
                    day: 'numeric',
                    year: 'numeric',
                  })}
                </td>
                <td>
                  {formatTime(lesson.start_time)} - {formatTime(lesson.end_time)}
                </td>
                <td>{lesson.teacher_name || 'Unknown'}</td>
                <td>
                  {lesson.current_students}/{lesson.max_students}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>

      {remaining > 0 && (
        <p className="preview-more">
          And {remaining} more lesson{remaining !== 1 ? 's' : ''}...
        </p>
      )}
    </div>
  );
};

export default AffectedLessonsPreview;
