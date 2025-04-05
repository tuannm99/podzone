const jwt = require('jsonwebtoken');

const payload = {
  key: '1', // This must match the key in your consumer config
  // key: 'jwt-key',
  uid: 10001,
  role: 'admin',
  permissions: ['read', 'write'],
};
const secret = '1234';
// const secret = '0+iYUDO3se6TwdwbpxEXVgi2Tw/y3wAoyB/eNB4ubD0=';
const token = jwt.sign(payload, secret, { expiresIn: 3600 });

console.log(token);
