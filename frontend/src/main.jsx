import React from 'react';
import ReactDOM from 'react-dom/client';
import { QueryClientProvider } from '@tanstack/react-query';
import App from './App.jsx';
import { queryClient } from './lib/queryClient.js';
import { initializeAuth } from './utils/authInit.js';
import './index.css';
import './styles/global.css';

// Инициализируем аутентификацию перед рендерингом приложения
initializeAuth();

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>
  </React.StrictMode>
);
