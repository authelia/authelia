
var sinon = require('sinon');
var server = require('../../src/lib/server');
var assert = require('assert');

describe('test server configuration', function() {
  it('should set cookie scope to domain set in the config', function() {
    var config = {};
    config.session_domain = 'example.com';
    config.notifier = {
      gmail: {
        user: 'user@example.com',
        pass: 'password'
      }
    }

    transporter = {};
    transporter.sendMail = sinon.stub().yields();

    var nodemailer = {};
    nodemailer.createTransport = sinon.spy(function() {
      return transporter;
    });

    var deps = {};
    deps.nedb = require('nedb');
    deps.nodemailer = nodemailer;
    deps.session = sinon.spy(function() {
      return function(req, res, next) { next(); };
    });

    server.run(config, undefined, deps);

    assert(deps.session.calledOnce);
    assert.equal(deps.session.getCall(0).args[0].cookie.domain, 'example.com');
  });  
});
