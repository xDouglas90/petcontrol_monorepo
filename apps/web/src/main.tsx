import React from 'react';
import ReactDOM from 'react-dom/client';
import { QueryClientProvider } from '@tanstack/react-query';
import { RouterProvider } from '@tanstack/react-router';

import './styles/globals.css';

import { queryClientForWeb, router } from '@/router';

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClientForWeb()}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </React.StrictMode>,
);
