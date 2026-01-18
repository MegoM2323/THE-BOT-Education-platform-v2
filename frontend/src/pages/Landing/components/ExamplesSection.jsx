import { formatDate } from '../../../utils/dateFormat.js';

const ExamplesSection = () => {
  // Вычисляем сегодняшнюю и вчерашнюю дату
  const today = new Date();
  const yesterday = new Date(today);
  yesterday.setDate(yesterday.getDate() - 1);

  const linkUrl = 'https://t.me/m/3mWQebLbMjUy';

  // Функция для рендеринга текста с буквами как ссылками
  const renderTextWithLinks = (text) => {
    // Сначала убираем URL в скобках
    const cleanedText = text.replace(/\s*\(https:\/\/t\.me\/m\/3mWQebLbMjUy\)/g, '');
    
    // Разбиваем текст на части, где буквы (A-Z) должны стать ссылками
    const parts = [];
    let lastIndex = 0;
    // Ищем заглавные латинские буквы как отдельные символы в контексте списка
    const letterPattern = /([A-Z])/g;
    let match;

    while ((match = letterPattern.exec(cleanedText)) !== null) {
      const charBefore = match.index > 0 ? cleanedText[match.index - 1] : '';
      const charAfter = match.index < cleanedText.length - 1 ? cleanedText[match.index + 1] : '';
      
      // Проверяем, что буква стоит в контексте списка (после пробела, запятой, двоеточия или в начале строки,
      // и перед пробелом, запятой или концом строки)
      const isInListContext = 
        (match.index === 0 || /[\s:,]/.test(charBefore)) &&
        (/[\s,]/.test(charAfter) || match.index === cleanedText.length - 1);
      
      if (isInListContext) {
        // Добавляем текст до буквы
        if (match.index > lastIndex) {
          parts.push({ type: 'text', content: cleanedText.substring(lastIndex, match.index) });
        }
        // Добавляем букву как ссылку
        parts.push({ type: 'link', content: match[1], url: linkUrl });
        lastIndex = match.index + 1;
      }
    }
    // Добавляем оставшийся текст
    if (lastIndex < cleanedText.length) {
      parts.push({ type: 'text', content: cleanedText.substring(lastIndex) });
    }

    // Если не найдено совпадений, возвращаем исходный текст
    if (parts.length === 0) {
      return <span>{cleanedText}</span>;
    }

    return parts.map((part, index) => {
      if (part.type === 'link') {
        return (
          <a key={index} href={part.url} target="_blank" rel="noopener noreferrer">
            {part.content}
          </a>
        );
      }
      return <span key={index}>{part.content}</span>;
    });
  };

  const homeworkExamples = [
    {
      id: 1,
      subject: 'Дерево отрезков',
      date: formatDate(today), // Всегда сегодняшняя дата
      teacher: 'Мирослав Адаменко',
      text: 'Задачи с урока: A (https://t.me/m/3mWQebLbMjUy), B (https://t.me/m/3mWQebLbMjUy), C (https://t.me/m/3mWQebLbMjUy)\nДЗ: D (https://t.me/m/3mWQebLbMjUy), E (https://t.me/m/3mWQebLbMjUy), F (https://t.me/m/3mWQebLbMjUy)',
      files: [{ name: 'segtree.pdf' }],
    },
    {
      id: 2,
      subject: 'Сканирующая прямая',
      date: formatDate(yesterday), // Всегда вчерашняя дата
      teacher: 'Мирослав Адаменко',
      text: 'Одномерный сканлайн для практики: A (https://t.me/m/3mWQebLbMjUy), B (https://t.me/m/3mWQebLbMjUy), C (https://t.me/m/3mWQebLbMjUy), D (https://t.me/m/3mWQebLbMjUy)\nЗадачи с урока: E (https://t.me/m/3mWQebLbMjUy), F (https://t.me/m/3mWQebLbMjUy)\nПопытаться решить самим: G (https://t.me/m/3mWQebLbMjUy), H (https://t.me/m/3mWQebLbMjUy), I (https://t.me/m/3mWQebLbMjUy)',
      files: [{ name: 'scanline.pdf' }],
    },
  ];

  return (
    <section className="examples-section" id="examples">
      <div className="container">
        <p className="examples-intro">
          После каждого из занятий вы получаете домашнее задание, кропотливо подобранное именно для вас, а также
          конспект, охватывающий все, что было на уроке.
        </p>

        <div className="homework-examples-grid">
          {homeworkExamples.map((hw) => (
            <div key={hw.id} className="homework-example homework-example-tile">
              <div className="homework-example-header">
                <div className="homework-example-subject">{hw.subject}</div>
                <div className="homework-example-date">Занятие от {hw.date}</div>
              </div>

              <div className="homework-example-teacher">
                <span className="teacher-label">Преподаватель:</span>
                <span className="teacher-name">{hw.teacher}</span>
              </div>

              <div className="homework-example-content">
                <p className="homework-example-text" style={{ whiteSpace: 'pre-line' }}>
                  {renderTextWithLinks(hw.text)}
                </p>
              </div>

              <div className="homework-example-files">
                <div className="files-label">Прикрепленные файлы:</div>
                {hw.files.map((file, index) => (
                  <a
                    key={index}
                    href={`/files/${file.name}`}
                    download
                    className="file-link-demo"
                  >
                    <svg className="file-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                        d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                    </svg>
                    {file.displayName || file.name}
                  </a>
                ))}
              </div>
            </div>
          ))}
        </div>

        <p className="examples-footer">
          PDF конспекты и материалы доступны для скачивания
        </p>
      </div>
    </section>
  );
};

export default ExamplesSection;
