import React from 'react';
import { Button, Form, Input, Typography, Card, Divider } from 'antd';
import { GoogleOutlined } from '@ant-design/icons';
import { Link, useNavigate } from 'react-router-dom';
import { loginGG, login } from '../../services/auth';
import { toast } from 'react-toastify';
import { UseAuth } from '../../auth/auth.context';

const { Title } = Typography;

const LoginPage = () => {
  const { setSession } = UseAuth();
  const [form] = Form.useForm();
  const navigate = useNavigate();

  const onFinish = async (values) => {
    const { success, data } = await login({
      username: values.username,
      password: values.password,
    });
    if (!success) {
      toast.error(data.message);
      return;
    }

    setSession({ ...data });

    navigate('/home', { replace: true });
  };

  const handleGoogleLogin = async () => {
    window.location.href = await loginGG();
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
            name="username"
            label="Username or email"
            rules={[
              {
                required: true,
                message: 'Please input your username or email!',
              },
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

        <Button
          icon={<GoogleOutlined />}
          block
          size="large"
          onClick={handleGoogleLogin}
        >
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
