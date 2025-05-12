import React from 'react';
import { Button, Layout } from 'antd';
import { Link } from 'react-router-dom';

const { Header: AntHeader } = Layout;

const Header = () => {
    return (
        <AntHeader
            style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                padding: '0 24px',
                background: 'linear-gradient(to right, #1890ff, #096dd9)',
                boxShadow: '0 2px 8px rgba(0,0,0,0.15)',
            }}
        >
            <div
                style={{
                    fontSize: '24px',
                    fontWeight: 'bold',
                    color: 'white',
                    textShadow: '1px 1px 2px rgba(0,0,0,0.2)',
                }}
            >
                Admin App
            </div>
            <div>
                <Button
                    type="primary"
                    ghost
                    size="middle"
                    style={{
                        marginRight: '12px',
                        borderColor: 'white',
                        color: 'white',
                    }}
                >
                    Logout
                </Button>
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
            </div>
        </AntHeader>
    );
};

export default Header;
