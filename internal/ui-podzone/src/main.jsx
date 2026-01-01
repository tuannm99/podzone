import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import App from './app';

import 'antd/dist/reset.css';
import 'nprogress/nprogress.css';

import { Bounce, ToastContainer } from 'react-toastify';
import { AuthProvider } from './auth/auth.context';
import RouteTransition from './components/RouteTransition';

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <BrowserRouter>
      <AuthProvider>
        {/* Route change loader */}
        <RouteTransition />

        <App />

        <ToastContainer
          position="top-right"
          autoClose={3000}
          theme="light"
          transition={Bounce}
        />
      </AuthProvider>
    </BrowserRouter>
  </React.StrictMode>,
);
