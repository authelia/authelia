
import sinon = require("sinon");
import IdentityValidator = require("../../src/lib/IdentityValidator");
import exceptions = require("../../src/lib/Exceptions");
import assert = require("assert");
import winston = require("winston");
import Promise = require("bluebird");
import express = require("express");
import BluebirdPromise = require("bluebird");

import ExpressMock = require("./mocks/express");
import UserDataStoreMock = require("./mocks/UserDataStore");
import NotifierMock = require("./mocks/Notifier");
import IdentityValidatorMock = require("./mocks/IdentityValidator");


describe("test identity check process", function () {
  let req: ExpressMock.RequestMock;
  let res: ExpressMock.ResponseMock;
  let userDataStore: UserDataStoreMock.UserDataStore;
  let notifier: NotifierMock.NotifierMock;
  let app: express.Application;
  let app_get: sinon.SinonStub;
  let app_post: sinon.SinonStub;
  let identityValidable: IdentityValidatorMock.IdentityValidableMock;

  beforeEach(function () {
    req = ExpressMock.RequestMock();
    res = ExpressMock.ResponseMock();

    userDataStore = UserDataStoreMock.UserDataStore();
    userDataStore.issue_identity_check_token = sinon.stub();
    userDataStore.issue_identity_check_token.returns(Promise.resolve());
    userDataStore.consume_identity_check_token = sinon.stub();
    userDataStore.consume_identity_check_token.returns(Promise.resolve({ userid: "user" }));

    notifier = NotifierMock.NotifierMock();
    notifier.notify = sinon.stub().returns(Promise.resolve());

    req.headers = {};
    req.session = {};
    req.session.auth_session = {};

    req.query = {};
    req.app = {};
    req.app.get = sinon.stub();
    req.app.get.withArgs("logger").returns(winston);
    req.app.get.withArgs("user data store").returns(userDataStore);
    req.app.get.withArgs("notifier").returns(notifier);

    app = express();
    app_get = sinon.stub(app, "get");
    app_post = sinon.stub(app, "post");

    identityValidable = IdentityValidatorMock.IdentityValidableMock();
  });

  afterEach(function () {
    app_get.restore();
    app_post.restore();
  });

  it("should register a POST and GET endpoint", function () {
    const endpoint = "/test";
    const icheck_interface = {};

    IdentityValidator.IdentityValidator.setup(app, endpoint, identityValidable, userDataStore as any, winston);

    assert(app_get.calledOnce);
    assert(app_get.calledWith(endpoint));

    assert(app_post.calledOnce);
    assert(app_post.calledWith(endpoint));
  });

  describe("test POST", test_post_handler);
  describe("test GET", test_get_handler);

  function test_post_handler() {
    it("should send 403 if pre check rejects", function (done) {
      const endpoint = "/protected";

      identityValidable.preValidation.returns(Promise.reject("No access"));
      IdentityValidator.IdentityValidator.setup(app, endpoint, identityValidable, userDataStore as any, winston);

      res.send = sinon.spy(function () {
        assert.equal(res.status.getCall(0).args[0], 403);
        done();
      });

      const handler = app_post.getCall(0).args[1];
      handler(req, res);
    });

    it("should send 400 if email is missing in provided identity", function (done) {
      const endpoint = "/protected";
      const identity = { userid: "abc" };

      identityValidable.preValidation.returns(Promise.resolve(identity));
      IdentityValidator.IdentityValidator.setup(app, endpoint, identityValidable, userDataStore as any, winston);

      res.send = sinon.spy(function () {
        assert.equal(res.status.getCall(0).args[0], 400);
        done();
      });

      const handler = app_post.getCall(0).args[1];
      handler(req, res);
    });

    it("should send 400 if userid is missing in provided identity", function (done) {
      const endpoint = "/protected";
      const identity = { email: "abc@example.com" };

      identityValidable.preValidation.returns(Promise.resolve(identity));
      IdentityValidator.IdentityValidator.setup(app, endpoint, identityValidable, userDataStore as any, winston);

      res.send = sinon.spy(function () {
        assert.equal(res.status.getCall(0).args[0], 400);
        done();
      });
      const handler = app_post.getCall(0).args[1];
      handler(req, res);
    });

    describe("should issue a token, send an email and return 204", () => {
      function contains(str: string, pattern: string): boolean {
        return str.indexOf(pattern) > -1;
      }

      it("with x-original-uri", function(done) {
        const endpoint = "/protected";
        const identity = { userid: "user", email: "abc@example.com" };
        req.headers.host = "localhost";
        req.headers["x-original-uri"] = "/auth/test";

        identityValidable.preValidation.returns(Promise.resolve(identity));
        IdentityValidator.IdentityValidator.setup(app, endpoint, identityValidable, userDataStore as any, winston);

        res.send = sinon.spy(function () {
          assert.equal(res.status.getCall(0).args[0], 204);
          assert(notifier.notify.calledOnce);
                    console.log(notifier.notify.getCall(0).args[2]);
          assert(contains(notifier.notify.getCall(0).args[2], "https://localhost/auth/test?identity_token="));
          assert(userDataStore.issue_identity_check_token.calledOnce);
          assert.equal(userDataStore.issue_identity_check_token.getCall(0).args[0], "user");
          assert.equal(userDataStore.issue_identity_check_token.getCall(0).args[3], 240000);
          done();
        });
        const handler = app_post.getCall(0).args[1];
        handler(req, res);
      });

      it("without x-original-uri", function(done) {
        const endpoint = "/protected";
        const identity = { userid: "user", email: "abc@example.com" };
        req.headers.host = "localhost";

        identityValidable.preValidation.returns(Promise.resolve(identity));
        IdentityValidator.IdentityValidator.setup(app, endpoint, identityValidable, userDataStore as any, winston);

        res.send = sinon.spy(function () {
          assert.equal(res.status.getCall(0).args[0], 204);
          assert(notifier.notify.calledOnce);
          assert(contains(notifier.notify.getCall(0).args[2], "https://localhost?identity_token="));
          assert(userDataStore.issue_identity_check_token.calledOnce);
          assert.equal(userDataStore.issue_identity_check_token.getCall(0).args[0], "user");
          assert.equal(userDataStore.issue_identity_check_token.getCall(0).args[3], 240000);
          done();
        });
        const handler = app_post.getCall(0).args[1];
        handler(req, res);
      });
    });
  }

  function test_get_handler() {
    it("should send 403 if no identity_token is provided", function (done) {
      const endpoint = "/protected";

      IdentityValidator.IdentityValidator.setup(app, endpoint, identityValidable, userDataStore as any, winston);

      res.send = sinon.spy(function () {
        assert.equal(res.status.getCall(0).args[0], 403);
        done();
      });
      const handler = app_get.getCall(0).args[1];
      handler(req, res);
    });

    it("should render template if identity_token is provided and still valid", function (done) {
      req.query.identity_token = "token";
      const endpoint = "/protected";
      identityValidable.templateName.returns("template");

      IdentityValidator.IdentityValidator.setup(app, endpoint, identityValidable, userDataStore as any, winston);

      res.render = sinon.spy(function (template: string) {
        assert.equal(template, "template");
        done();
      });
      const handler = app_get.getCall(0).args[1];
      handler(req, res);
    });

    it("should return 403 if identity_token is provided but invalid", function (done) {
      req.query.identity_token = "token";
      const endpoint = "/protected";

      identityValidable.templateName.returns("template");
      userDataStore.consume_identity_check_token
        .returns(Promise.reject("Invalid token"));

      IdentityValidator.IdentityValidator.setup(app, endpoint, identityValidable, userDataStore as any, winston);

      res.send = sinon.spy(function (template: string) {
        assert.equal(res.status.getCall(0).args[0], 403);
        done();
      });
      const handler = app_get.getCall(0).args[1];
      handler(req, res);
    });

    it("should set the identity_check session object even if session does not exist yet", function (done) {
      req.query.identity_token = "token";
      const endpoint = "/protected";

      req.session = {};
      identityValidable.templateName.returns("template");

      IdentityValidator.IdentityValidator.setup(app, endpoint, identityValidable, userDataStore as any, winston);

      res.render = sinon.spy(function (template: string) {
        assert.equal(req.session.auth_session.identity_check.userid, "user");
        assert.equal(template, "template");
        done();
      });
      const handler = app_get.getCall(0).args[1];
      handler(req, res);
    });
  }
});
