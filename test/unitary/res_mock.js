
module.exports = create_res_mock;

var sinon = require('sinon');
var sinonPromise = require('sinon-promise');
sinonPromise(sinon);

function create_res_mock() {
  var status_mock = sinon.mock();
  var send_mock = sinon.mock();
  var set_mock = sinon.mock();
  var cookie_mock = sinon.mock();
  var render_mock = sinon.mock();
  var redirect_mock = sinon.mock();

  return {
    status: status_mock,
    send: send_mock,
    set: set_mock,
    cookie: cookie_mock,
    render: render_mock,
    redirect: redirect_mock
  };
}
