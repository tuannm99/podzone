import React from 'react';
import { Result, Button } from 'antd';
import { Link } from 'react-router-dom';

const Page404 = () => {
  return (
    <Result
      status="404"
      title="404 Page Not Found"
      extra={
        <Link to="/">
          <Button type="primary">Back to Home</Button>
        </Link>
      }
    />
  );
};

export default Page404;
