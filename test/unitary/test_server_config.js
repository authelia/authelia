
var sinon = require('sinon');
var server = require('../../src/lib/server');
var assert = require('assert');

describe('test server configuration', function() {
  var deps;
  var config;

  before(function() {
    config = {};
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

    deps = {};
    deps.nedb = require('nedb');
    deps.winston = sinon.spy();
    deps.nodemailer = nodemailer;
    deps.ldapjs = {};
    deps.ldapjs.createClient = sinon.spy(function() {
      return { on: sinon.spy() };
    });
    deps.session = sinon.spy(function() {
      return function(req, res, next) { next(); };
    });
  });


  it('should set cookie scope to domain set in the config', function() {
    config.session = {};
    config.session.domain = 'example.com';
    config.session.secret = 'secret';
    config.ldap = {};
    config.ldap.url = 'http://ldap';
    server.run(config, deps);

    assert(deps.session.calledOnce);
    assert.equal(deps.session.getCall(0).args[0].cookie.domain, 'example.com');
  });  
});
