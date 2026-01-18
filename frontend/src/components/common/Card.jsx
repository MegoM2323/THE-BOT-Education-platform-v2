import './Card.css';

export const Card = ({ children, className = '', onClick, hover = false, style, ...rest }) => {
  const classes = [
    'card',
    hover && 'card-hover',
    onClick && 'card-clickable',
    className,
  ]
    .filter(Boolean)
    .join(' ');

  return (
    <div className={classes} onClick={onClick} style={style} {...rest}>
      {children}
    </div>
  );
};

export default Card;
