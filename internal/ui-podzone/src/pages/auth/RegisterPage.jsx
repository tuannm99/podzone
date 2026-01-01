import React from 'react';
import { Button, Form, Input, Typography, Card } from 'antd';
import { Link, useNavigate } from 'react-router-dom';
import { register } from '../../services/auth';
import { toast } from 'react-toastify';

const RegisterPage = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();

  const onFinish = async (values) => {
    const { success, data } = await register({
      username: values.username,
      email: values.email,
      password: values.password,
    });

    if (!success) {
      toast.error(data?.message || 'Register failed');
      return;
    }

    // token/user đã được set trong service register()
    navigate('/home', { replace: true });
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
        <Typography.Title
          level={2}
          style={{ textAlign: 'center', marginBottom: '24px' }}
        >
          Register
        </Typography.Title>

        <Form form={form} layout="vertical" onFinish={onFinish}>
          <Form.Item
            name="username"
            label="Username"
            rules={[{ required: true, message: 'Please input your username!' }]}
          >
            <Input size="large" placeholder="Enter your username" />
          </Form.Item>

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
                  return Promise.reject(
                    new Error('The passwords do not match!'),
                  );
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
