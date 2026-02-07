export const normalizeOperationType = (type) => String(type || '').toLowerCase().trim();

export const OPERATION_TYPE_LABELS = {
  add: 'Начисление',
  deduct: 'Списание',
  refund: 'Возврат',
};

export const REASON_LABELS = {
  'Booking lesson': 'Запись на занятие',
  'Admin booking for student': 'Запись администратором',
  'Booking cancelled': 'Отмена записи',
};

export const translateReason = (reason) => {
  if (!reason) return '—';
  if (REASON_LABELS[reason]) return REASON_LABELS[reason];
  if (reason.startsWith('Lesson booking (recurring')) return 'Запись на повторяющееся занятие';
  if (reason.startsWith('Recurring rollback')) return 'Откат повторяющейся серии';
  if (reason.startsWith('Lesson cancelled during')) return 'Отмена при замене';
  return reason;
};

export const getOperationTypeLabel = (type) => {
  const normalizedType = normalizeOperationType(type);
  return OPERATION_TYPE_LABELS[normalizedType] || type || 'Неизвестно';
};

export const getTransactionSign = (amount) => {
  return amount > 0 ? '+' : '-';
};

export const isPositiveOperation = (operationType) => {
  const normalizedType = normalizeOperationType(operationType);
  return normalizedType === 'add' || normalizedType === 'refund';
};

export const getTransactionSignByType = (operationType) => {
  return isPositiveOperation(operationType) ? '+' : '-';
};
