import './SkeletonLoader.css';

/**
 * Skeleton Loader Component
 * Shows animated placeholder while content is loading
 *
 * @param {string} variant - Type of skeleton: 'text', 'title', 'card', 'table-row', 'avatar', 'calendar-cell'
 * @param {number} count - Number of skeleton items to render (default: 1)
 * @param {string} width - Custom width
 * @param {string} height - Custom height
 * @param {string} className - Additional CSS classes
 */
export const SkeletonLoader = ({
  variant = 'text',
  count = 1,
  width,
  height,
  className = ''
}) => {
  const skeletons = Array.from({ length: count }, (_, i) => i);

  const getSkeletonClass = () => {
    const baseClass = 'skeleton';
    const variantClass = `skeleton-${variant}`;
    return [baseClass, variantClass, className].filter(Boolean).join(' ');
  };

  const style = {
    ...(width && { width }),
    ...(height && { height }),
  };

  return (
    <>
      {skeletons.map((i) => (
        <div
          key={i}
          className={getSkeletonClass()}
          style={style}
          aria-hidden="true"
        />
      ))}
    </>
  );
};

/**
 * Skeleton Card - for lesson cards, user cards, etc.
 */
export const SkeletonCard = ({ className = '' }) => (
  <div className={`skeleton-card ${className}`} aria-hidden="true">
    <SkeletonLoader variant="title" />
    <SkeletonLoader variant="text" count={2} />
    <div className="skeleton-card-footer">
      <SkeletonLoader variant="text" width="60%" />
    </div>
  </div>
);

/**
 * Skeleton Table - for data tables
 */
export const SkeletonTable = ({ rows = 5, columns = 4, className = '' }) => (
  <div className={`skeleton-table ${className}`} aria-hidden="true" role="status" aria-label="Загрузка таблицы">
    <table>
      <thead>
        <tr>
          {Array.from({ length: columns }).map((_, i) => (
            <th key={i}>
              <SkeletonLoader variant="text" width="80%" />
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {Array.from({ length: rows }).map((_, i) => (
          <tr key={i}>
            {Array.from({ length: columns }).map((_, j) => (
              <td key={j}>
                <SkeletonLoader variant="text" />
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  </div>
);

/**
 * Skeleton Calendar - for calendar views
 */
export const SkeletonCalendar = ({ className = '' }) => (
  <div className={`skeleton-calendar ${className}`} aria-hidden="true" role="status" aria-label="Загрузка календаря">
    <div className="skeleton-calendar-header">
      <SkeletonLoader variant="title" width="200px" />
      <div className="skeleton-calendar-nav">
        <SkeletonLoader variant="text" width="80px" count={3} />
      </div>
    </div>
    <div className="skeleton-calendar-grid">
      {Array.from({ length: 7 }).map((_, dayIndex) => (
        <div key={dayIndex} className="skeleton-calendar-day">
          <SkeletonLoader variant="text" width="60px" />
          <div className="skeleton-calendar-cells">
            <SkeletonLoader variant="calendar-cell" count={3} />
          </div>
        </div>
      ))}
    </div>
  </div>
);

/**
 * Skeleton List - for lesson lists, homework lists, etc.
 */
export const SkeletonList = ({ items = 5, className = '' }) => (
  <div className={`skeleton-list ${className}`} aria-hidden="true" role="status" aria-label="Загрузка списка">
    {Array.from({ length: items }).map((_, i) => (
      <SkeletonCard key={i} />
    ))}
  </div>
);

/**
 * Skeleton Dashboard - for dashboard loading
 */
export const SkeletonDashboard = ({ className = '' }) => (
  <div className={`skeleton-dashboard ${className}`} aria-hidden="true" role="status" aria-label="Загрузка панели">
    <div className="skeleton-dashboard-header">
      <SkeletonLoader variant="title" width="300px" />
      <SkeletonLoader variant="text" width="150px" />
    </div>
    <div className="skeleton-dashboard-content">
      <SkeletonList items={3} />
    </div>
  </div>
);

export default SkeletonLoader;
