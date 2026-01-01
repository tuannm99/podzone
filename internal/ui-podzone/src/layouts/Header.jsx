import React from 'react';
import { Button, Layout, Space, Typography } from 'antd';
import { Link, useLocation } from 'react-router-dom';
import { logout } from '../services/auth';
import { tokenStorage } from '../services/tokenStorage';

const { Text } = Typography;

const Header = () => {
  const location = useLocation();
  const token = tokenStorage.getToken();
  const user = tokenStorage.getUser();
  const isAuthed = !!token;

  const isAuthPage =
    location.pathname.startsWith('/auth/login') ||
    location.pathname.startsWith('/auth/register');

  return (
    <Layout.Header
      style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: '0 24px',
        background: 'linear-gradient(to right, #1890ff, #096dd9)',
        boxShadow: '0 2px 8px rgba(0,0,0,0.15)',
      }}
    >
      <Link to="/" style={{ textDecoration: 'none' }}>
        <div
          style={{
            fontSize: '20px',
            fontWeight: 'bold',
            color: 'white',
            textShadow: '1px 1px 2px rgba(0,0,0,0.2)',
          }}
        >
          PODZONE
        </div>
      </Link>

      <Space size={12}>
        {isAuthed && (
          <Text style={{ color: 'white', opacity: 0.9 }}>
            {user?.username ? `Hi, ${user.username}` : ''}
          </Text>
        )}

        {isAuthed ? (
          <Button
            type="primary"
            ghost
            size="middle"
            style={{
              borderColor: 'white',
              color: 'white',
            }}
            onClick={logout}
          >
            Logout
          </Button>
        ) : (
          !isAuthPage && (
            <Link to="/auth/login">
              <Button
                type="primary"
                size="middle"
                style={{
                  background: 'white',
                  color: '#1890ff',
                  border: 'none',
                }}
              >
                Login
              </Button>
            </Link>
          )
        )}
      </Space>
    </Layout.Header>
  );
};

export default Header;
