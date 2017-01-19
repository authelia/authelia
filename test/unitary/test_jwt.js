
var Jwt = require('../../src/lib/jwt');
var sinon = require('sinon');

describe('test jwt', function() {
  it('should sign and verify the token', function() {
    var data = {user: 'user'};
    var secret = 'secret';
    var jwt = new Jwt(secret);
    return jwt.sign(data, '1m')
    .then(function(token) {
      return jwt.verify(token);
    });
  });

  it('should verify and fail on wrong token', function() {
    var jwt = new Jwt('secret');
    var token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjoidXNlciIsImlhdCI6MTQ4NDc4NTExMywiZXhwIjoaNDg0Nzg1MTczfQ.yZOZEaMDyOn0tSDiDSPYl4ZP2oL3FQ-Vrzds7hYcNio';
    return jwt.verify(token).catch(function() {
      return Promise.resolve();
    });
  });

  it('should fail after expiry', function() {
    var clock = sinon.useFakeTimers(0);
    var data = { user: 'user' };
    var jwt = new Jwt('secret');
    return jwt.sign(data, '1m')
    .then(function(token) {
      clock.tick(1000 * 61); // 61 seconds
      return jwt.verify(token);
    })
    .catch(function() {
      clock.restore();
      return Promise.resolve();
    });
  });
});

