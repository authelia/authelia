import * as Express from "express";
import * as ExpressMock from "../../stubs/express.spec";
import { ImportMock } from 'ts-mock-imports';
import * as CheckAuthorizations from "./CheckAuthorizations";
import * as CheckInactivity from "./CheckInactivity";
import GetSessionCookie from "./GetSessionCookie";
import { ServerVariables } from "../../ServerVariables";
import { ServerVariablesMockBuilder } from "../../ServerVariablesMockBuilder.spec";
import { AuthenticationSession } from "../../../../types/AuthenticationSession";
import AssertRejects from "../../utils/AssertRejects";
import { Level } from "../../authorization/Level";


describe('routes/verify/GetSessionCookie', function() {
  let req: Express.Request;
  let res: ExpressMock.ResponseMock;
  let vars: ServerVariables;
  let authSession: AuthenticationSession;
  
  beforeEach(function() {
    req = ExpressMock.RequestMock();
    res = ExpressMock.ResponseMock();
    const sv = ServerVariablesMockBuilder.build();
    vars = sv.variables;
    authSession = {} as any;
  });

  it("should fail when target url is not provided", async function() {
    AssertRejects(async () => await GetSessionCookie(req, res as any, vars, authSession));
  });

  it("should not let unauthorized users in", async function() {
    req.originalUrl = "https://public.example.com";
    const mock = ImportMock.mockOther(CheckAuthorizations, "default", () => { throw new Error('Not authorized')});
    AssertRejects(async () => await GetSessionCookie(req, res as any, vars, authSession));
    mock.restore();
  });

  it("should not let authorize user after a long period of inactivity", async function() {
    req.originalUrl = "https://public.example.com";
    const checkAuthorizationsMock = ImportMock.mockOther(CheckAuthorizations, "default", () => Level.ONE_FACTOR);
    const checkInactivityMock = ImportMock.mockOther(CheckInactivity, "default", () => { throw new Error('Timed out')});
    AssertRejects(async () => await GetSessionCookie(req, res as any, vars, authSession));
    checkInactivityMock.restore();
    checkAuthorizationsMock.restore();
  });

  it("should let the user in", async function() {
    req.headers['x-original-url'] = "https://public.example.com";

    const checkAuthorizationsMock = ImportMock.mockOther(CheckAuthorizations, "default", () => Level.ONE_FACTOR);
    const checkInactivityMock = ImportMock.mockOther(CheckInactivity, "default", () => {});
    await GetSessionCookie(req, res as any, vars, authSession);
    checkInactivityMock.restore();
    checkAuthorizationsMock.restore();
  });
});