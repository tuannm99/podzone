import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, Layout } from 'antd';

import LoginPage from './pages/auth/LoginPage';
import RegisterPage from './pages/auth/RegisterPage';
import HomePage from './pages/private/HomePage';
import AppLayout from './layouts/layout';

const PrivateRoute = ({ children }) => {
    const isAuthenticated = false;
    return isAuthenticated ? children : <Navigate to="/auth/login" />;
};

const App = () => {
    return (
        <ConfigProvider>
            <Layout>
                <Layout.Content>
                    <Routes>
                        <Route path="/auth/login" element={<LoginPage />} />
                        <Route path="/auth/register" element={<RegisterPage />} />

                        <Route
                            path="/home"
                            element={
                                <AppLayout>
                                    <HomePage />
                                </AppLayout>
                            }
                        />
                        <Route
                            path="/"
                            element={
                                <PrivateRoute>
                                    <AppLayout>
                                        <HomePage />
                                    </AppLayout>
                                </PrivateRoute>
                            }
                        />
                    </Routes>
                </Layout.Content>
            </Layout>
        </ConfigProvider>
    );
};

export default App;
