
var replies = require('../lib/replies');
var assert = require('assert');
var sinon = require('sinon');
var sinonPromise = require('sinon-promise');
sinonPromise(sinon);

var autoResolving = sinon.promise().resolves();

function create_res_mock() {
  var status_mock = sinon.mock();
  var send_mock = sinon.mock();
  var set_mock = sinon.mock();

  return {
    status: status_mock,
    send: send_mock,
    set: set_mock
  };
}

describe('test jwt', function() {
  it('should authenticate with success', function() {
    var res_mock = create_res_mock();
    var username = 'username';

    replies.authentication_succeeded(res_mock, username);

    assert(res_mock.status.calledWith(200));
    assert(res_mock.set.calledWith({'X-Remote-User': username }));
  });

  it('should reply successfully when already authenticated', function() {
    var res_mock = create_res_mock();
    var username = 'username';

    replies.already_authenticated(res_mock, username);

    assert(res_mock.status.calledWith(200));
    assert(res_mock.set.calledWith({'X-Remote-User': username }));
  });

  it('should reply with failed authentication', function() {
    var res_mock = create_res_mock();
    var username = 'username';

    replies.authentication_failed(res_mock, username);

    assert(res_mock.status.calledWith(401));
  });
});

