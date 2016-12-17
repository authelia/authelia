
var Jwt = require('../../src/lib/jwt');
var sinon = require('sinon');
var sinonPromise = require('sinon-promise');
sinonPromise(sinon);

var autoResolving = sinon.promise().resolves();

describe('test jwt', function() {
  it('should sign and verify the token', function() {
    var data = {user: 'user'};
    var secret = 'secret';
    var jwt = new Jwt(secret);
    var token = jwt.sign(data, '1m');
    return jwt.verify(token);
  });

  it('should verify and fail on wrong token', function() {
    var jwt = new Jwt('secret');
    return jwt.verify('wrong token').fail(autoResolving);
  });

  it('should fail after expiry', function(done) {
    var clock = sinon.useFakeTimers(0);
    var data = {user: 'user'};
    var jwt = new Jwt('secret');
    var token = jwt.sign(data, '1m');
    clock.tick(1000 * 61); // 61 seconds
    jwt.verify(token).fail(function() { done(); });
    clock.restore();
  });
});

