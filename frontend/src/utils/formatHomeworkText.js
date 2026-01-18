import React from 'react';

export const formatHomeworkText = (text) => {
  if (!text || typeof text !== 'string' || text.trim() === '') {
    return null;
  }

  const urlRegex = /https?:\/\/[^\s]+/g;
  const result = [];

  const lines = text.split('\n');

  lines.forEach((line, lineIndex) => {
    const parts = [];
    let lastIndex = 0;
    let match;

    const regex = new RegExp(urlRegex);
    while ((match = regex.exec(line)) !== null) {
      if (match.index > lastIndex) {
        parts.push(line.substring(lastIndex, match.index));
      }
      parts.push(
        <a key={`link-${lineIndex}-${match.index}`} href={match[0]} target="_blank" rel="noopener noreferrer">
          {match[0]}
        </a>
      );
      lastIndex = regex.lastIndex;
    }

    if (lastIndex < line.length) {
      parts.push(line.substring(lastIndex));
    }

    if (parts.length === 0) {
      parts.push('');
    }

    result.push(...parts);

    if (lineIndex < lines.length - 1) {
      result.push(<br key={`br-${lineIndex}`} />);
    }
  });

  return result;
};

export default formatHomeworkText;
