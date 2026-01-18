import './Spinner.css';

export const Spinner = ({ size = 'md', className = '' }) => {
  const classes = ['spinner', `spinner-${size}`, className]
    .filter(Boolean)
    .join(' ');

  return <div className={classes} role="status" aria-label="Загрузка"></div>;
};

export default Spinner;
