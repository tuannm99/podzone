import React from 'react';
import { Button, Form, Input, Typography, Card, Divider } from 'antd';
import { GoogleOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';
import { ADMIN_API_URL } from '../../services/baseurl';

const { Title } = Typography;

const LoginPage = () => {
    const [form] = Form.useForm();

    const onFinish = (values) => {
        console.log('Login values:', values);
    };

    const handleGoogleLogin = () => {
        const loginUrl = `${ADMIN_API_URL || 'http://localhost:8080'}/auth/v1/google/login`;
        window.location.href = loginUrl;
    };

    return (
        <div
            style={{
                display: 'flex',
                justifyContent: 'center',
                alignItems: 'center',
                minHeight: '100vh',
                padding: '16px',
            }}
        >
            <Card style={{ width: '100%', maxWidth: '400px' }}>
                <Title level={2} style={{ textAlign: 'center', marginBottom: '24px' }}>
                    Login
                </Title>

                <Form form={form} layout="vertical" onFinish={onFinish}>
                    <Form.Item
                        name="email"
                        label="Email"
                        rules={[
                            { required: true, message: 'Please input your email!' },
                            { type: 'email', message: 'Please enter a valid email!' },
                        ]}
                    >
                        <Input size="large" placeholder="Enter your email" />
                    </Form.Item>

                    <Form.Item
                        name="password"
                        label="Password"
                        rules={[{ required: true, message: 'Please input your password!' }]}
                    >
                        <Input.Password size="large" placeholder="Enter your password" />
                    </Form.Item>

                    <Form.Item>
                        <Button type="primary" htmlType="submit" block size="large">
                            Login
                        </Button>
                    </Form.Item>
                </Form>

                <Divider>Or</Divider>

                <Button icon={<GoogleOutlined />} block size="large" onClick={handleGoogleLogin}>
                    Continue with Google
                </Button>

                <div style={{ marginTop: '24px', textAlign: 'center' }}>
                    Don't have an account? <Link to="/auth/register">Register now</Link>
                </div>
            </Card>
        </div>
    );
};

export default LoginPage;
