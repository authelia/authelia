
import * as Assert from "assert";
import * as Express from "express";
import { AuthenticationSessionHandler } from "../../AuthenticationSessionHandler";
import { AuthenticationSession } from "../../../../types/AuthenticationSession";
import ExpressMock = require("../../stubs/express.spec");
import { ServerVariables } from "../../ServerVariables";
import { ServerVariablesMockBuilder, ServerVariablesMock } from "../../ServerVariablesMockBuilder.spec";
import { HEADER_X_ORIGINAL_URL } from "../../constants";
import Get from "./Get";
import { ImportMock } from 'ts-mock-imports';
import * as GetBasicAuth from "./GetBasicAuth";
import * as GetSessionCookie from "./GetSessionCookie";
import { NotAuthorizedError, NotAuthenticatedError } from "../../Exceptions";


describe("routes/verify/get", function () {
  let req: Express.Request;
  let res: ExpressMock.ResponseMock;
  let mocks: ServerVariablesMock;
  let vars: ServerVariables;
  let authSession: AuthenticationSession;

  beforeEach(function () {
    req = ExpressMock.RequestMock();
    res = ExpressMock.ResponseMock();
    req.query = {
      redirect: "undefined"
    };
    AuthenticationSessionHandler.reset(req as any);
    req.headers[HEADER_X_ORIGINAL_URL] = "https://secret.example.com/";
    const s = ServerVariablesMockBuilder.build(false);
    mocks = s.mocks;
    vars = s.variables;
    authSession = AuthenticationSessionHandler.get(req as any, vars.logger);
  });

  describe("with basic auth", function () {
    it('should allow access to user', async function() {
      req.headers['proxy-authorization'] = 'zglfzeljfzelmkj';
      const mock = ImportMock.mockOther(GetBasicAuth, "default", () => Promise.resolve());
      await Get(vars)(req, res as any);
      Assert(res.send.calledWithExactly());
      Assert(res.status.calledWithExactly(204))
      mock.restore();
    });
  });

  describe("with session cookie", function () {
    it('should allow access to user', async function() {
      const mock = ImportMock.mockOther(GetSessionCookie, "default", () => Promise.resolve());
      await Get(vars)(req, res as any);
      Assert(res.send.calledWithExactly());
      Assert(res.status.calledWithExactly(204))
      mock.restore();
    });
  });

  describe('Deny access', function() {
    it('should deny access to user on NotAuthorizedError', async function() {
      req.headers['proxy-authorization'] = 'zglfzeljfzelmkj';
      const mock = ImportMock.mockOther(GetBasicAuth, "default", () => Promise.reject(new NotAuthorizedError('No!')));
      await Get(vars)(req, res as any);
      Assert(res.status.calledWith(403));
      mock.restore();
    });

    it('should deny access to user on NotAuthenticatedError', async function() {
      req.headers['proxy-authorization'] = 'zglfzeljfzelmkj';
      const mock = ImportMock.mockOther(GetBasicAuth, "default", () => Promise.reject(new NotAuthenticatedError('No!')));
      await Get(vars)(req, res as any);
      Assert(res.status.calledWith(401));
      mock.restore();
    });

    it('should deny access to user on any exception', async function() {
      req.headers['proxy-authorization'] = 'zglfzeljfzelmkj';
      const mock = ImportMock.mockOther(GetBasicAuth, "default", () => Promise.reject(new Error('No!')));
      await Get(vars)(req, res as any);
      Assert(res.status.calledWith(401));
      mock.restore();
    });
  })

  describe('Kubernetes ingress controller', function() {
    it('should redirect user to login portal', async function() {
      req.headers['proxy-authorization'] = 'zglfzeljfzelmkj';
      req.query.rd = 'https://login.example.com/';
      const mock = ImportMock.mockOther(GetBasicAuth, "default", () => Promise.reject(new NotAuthenticatedError('No!')));
      await Get(vars)(req, res as any);
      Assert(res.redirect.calledWith('https://login.example.com/?rd=https://secret.example.com/'));
      mock.restore();
    });
  });
});

