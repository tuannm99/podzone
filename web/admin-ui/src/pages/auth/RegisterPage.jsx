import React from 'react';
import { Button, Form, Input, Typography, Card, Divider } from 'antd';
import { Link } from 'react-router-dom';

const { Title } = Typography;

const RegisterPage = () => {
    const [form] = Form.useForm();

    const onFinish = (values) => {
        console.log('Registration values:', values);
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
                    Register
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

                    <Form.Item
                        name="confirmPassword"
                        label="Confirm Password"
                        dependencies={['password']}
                        rules={[
                            { required: true, message: 'Please confirm your password!' },
                            ({ getFieldValue }) => ({
                                validator(_, value) {
                                    if (!value || getFieldValue('password') === value) {
                                        return Promise.resolve();
                                    }
                                    return Promise.reject(new Error('The passwords do not match!'));
                                },
                            }),
                        ]}
                    >
                        <Input.Password size="large" placeholder="Confirm your password" />
                    </Form.Item>

                    <Form.Item>
                        <Button type="primary" htmlType="submit" block size="large">
                            Register
                        </Button>
                    </Form.Item>
                </Form>

                <div style={{ marginTop: '24px', textAlign: 'center' }}>
                    Already have an account? <Link to="/auth/login">Login now</Link>
                </div>
            </Card>
        </div>
    );
};

export default RegisterPage;
