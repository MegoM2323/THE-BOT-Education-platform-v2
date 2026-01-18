import { useState, useEffect, useCallback, useMemo } from 'react';
import { FiInfo } from 'react-icons/fi';
import Spinner from '../common/Spinner.jsx';
import StudentCreditsHistoryModal from '../common/StudentCreditsHistoryModal.jsx';
import { useNotification } from '../../hooks/useNotification.js';
import * as usersAPI from '../../api/users.js';
import * as creditsAPI from '../../api/credits.js';
import './MethodologistCreditsView.css';

export const MethodologistCreditsView = () => {
  const [students, setStudents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [sortColumn, setSortColumn] = useState('name');
  const [sortDirection, setSortDirection] = useState('asc');
  const [showHistoryModal, setShowHistoryModal] = useState(false);
  const [selectedStudentForHistory, setSelectedStudentForHistory] = useState(null);
  const notification = useNotification();

  const fetchStudents = useCallback(async () => {
    try {
      setLoading(true);
      const data = await usersAPI.getStudentsAll();
      const studentsWithCredits = await Promise.all(
        data.map(async (student) => {
          try {
            const credits = await creditsAPI.getUserCredits(student.id);
            return { ...student, credits: credits.balance };
          } catch (error) {
            return { ...student, credits: 0 };
          }
        })
      );
      setStudents(studentsWithCredits);
    } catch (error) {
      notification.error('Ошибка загрузки студентов');
    } finally {
      setLoading(false);
    }
  }, [notification]);

  useEffect(() => {
    fetchStudents();
  }, [fetchStudents]);

  const handleSort = (column) => {
    if (sortColumn === column) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortColumn(column);
      setSortDirection('asc');
    }
  };

  const handleViewHistory = (student) => {
    setSelectedStudentForHistory(student);
    setShowHistoryModal(true);
  };

  const sortedStudents = useMemo(() => {
    const sorted = [...students];

    sorted.sort((a, b) => {
      let aValue, bValue;

      if (sortColumn === 'name') {
        aValue = a.full_name?.toLowerCase() || '';
        bValue = b.full_name?.toLowerCase() || '';
      } else if (sortColumn === 'balance') {
        aValue = a.credits || 0;
        bValue = b.credits || 0;
      }

      if (typeof aValue === 'string') {
        return sortDirection === 'asc'
          ? aValue.localeCompare(bValue, 'ru-RU')
          : bValue.localeCompare(aValue, 'ru-RU');
      }

      return sortDirection === 'asc' ? aValue - bValue : bValue - aValue;
    });

    return sorted;
  }, [students, sortColumn, sortDirection]);

  const SortIndicator = ({ column }) => {
    if (sortColumn !== column) return null;
    return <span className="sort-indicator">{sortDirection === 'asc' ? '↑' : '↓'}</span>;
  };

  if (loading) {
    return (
      <div className="methodologist-credits-loading">
        <Spinner />
      </div>
    );
  }

  return (
    <div className="methodologist-credits-view">
      <div className="methodologist-credits-header">
        <h2 className="methodologist-credits-title">Просмотр кредитов студентов</h2>
      </div>

      <div className="methodologist-credits-table-wrapper">
        <table className="methodologist-credits-table">
          <thead>
            <tr>
              <th
                className={`sortable ${sortColumn === 'name' ? 'active' : ''}`}
                onClick={() => handleSort('name')}
              >
                Имя <SortIndicator column="name" />
              </th>
              <th>Email</th>
              <th
                className={`sortable ${sortColumn === 'balance' ? 'active' : ''}`}
                onClick={() => handleSort('balance')}
              >
                Баланс <SortIndicator column="balance" />
              </th>
              <th>Действия</th>
            </tr>
          </thead>
          <tbody>
            {sortedStudents.length > 0 ? (
              sortedStudents.map((student) => (
                <tr key={student.id}>
                  <td>{student.full_name}</td>
                  <td>{student.email}</td>
                  <td>
                    <span className="balance-badge">
                      {student.credits} кредитов
                    </span>
                  </td>
                  <td className="actions-cell">
                    <button
                      className="history-action-btn"
                      onClick={() => handleViewHistory(student)}
                      aria-label="Просмотреть историю кредитов"
                      data-testid="view-history-button"
                    >
                      <FiInfo size={20} />
                    </button>
                  </td>
                </tr>
              ))
            ) : (
              <tr>
                <td colSpan="4" className="empty-message">
                  Студенты не найдены
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <StudentCreditsHistoryModal
        isOpen={showHistoryModal}
        onClose={() => {
          setShowHistoryModal(false);
          setSelectedStudentForHistory(null);
        }}
        student={selectedStudentForHistory}
      />
    </div>
  );
};

export default MethodologistCreditsView;
