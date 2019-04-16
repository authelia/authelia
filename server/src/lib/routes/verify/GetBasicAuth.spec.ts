import * as Express from "express";
import { ServerVariables } from "../../ServerVariables";
import * as ExpressMock from "../../stubs/express.spec";
import { ServerVariablesMockBuilder, ServerVariablesMock } from "../../ServerVariablesMockBuilder.spec";
import { HEADER_X_ORIGINAL_URL } from "../../constants";
import { Level } from "../../authorization/Level";
import GetBasicAuthModule from "./GetBasicAuth";
import * as CheckAuthorizations from "./CheckAuthorizations";
import { ImportMock } from 'ts-mock-imports';
import AssertRejects from "../../utils/AssertRejects";

describe('routes/verify/GetBasicAuth', function() {
  let req: Express.Request;
  let res: ExpressMock.ResponseMock;
  let vars: ServerVariables;
  let mocks: ServerVariablesMock;

  beforeEach(function() {
    req = ExpressMock.RequestMock();
    res = ExpressMock.ResponseMock();
    req.headers[HEADER_X_ORIGINAL_URL] = 'https://secure.example.com';
    const sv = ServerVariablesMockBuilder.build();
    vars = sv.variables;
    mocks = sv.mocks;
  })

  it('should fail on invalid format of token', async function() {
    req.headers['proxy-authorization'] = 'Basic abc';
    AssertRejects(async () => GetBasicAuthModule(req, res as any, vars));
  });

  it('should fail decoded token is not of form user:pass', function() {
    req.headers['proxy-authorization'] = 'Basic aGVsbG93b3JsZAo=';
    AssertRejects(async () => GetBasicAuthModule(req, res as any, vars));
  });

  it('should fail when credentials are wrong', function() {
    req.headers['proxy-authorization'] = 'Basic aGVsbG8xOndvcmxkCg==';
    mocks.usersDatabase.checkUserPasswordStub.rejects(new Error('Bad credentials'));
    AssertRejects(async () => await GetBasicAuthModule(req, res as any, vars));
  });

  it('should fail when authorizations are not sufficient', function() {
    req.headers['proxy-authorization'] = 'Basic aGVsbG8xOndvcmxkCg==';
    const mock = ImportMock.mockOther(CheckAuthorizations, 'default', () => { throw new Error('Not enough permissions.')});
    mocks.usersDatabase.checkUserPasswordStub.resolves({
      email: 'john@example.com',
      groups: ['group1', 'group2'],
    });
    AssertRejects(async () => await GetBasicAuthModule(req, res as any, vars));
    mock.restore();
  });

  it('should succeed when user is authenticated and authorizations are sufficient', async function() {
    req.headers['proxy-authorization'] = 'Basic aGVsbG8xOndvcmxkCg==';
    const mock = ImportMock.mockOther(CheckAuthorizations, 'default', () => Level.TWO_FACTOR);
    mocks.usersDatabase.checkUserPasswordStub.resolves({
      email: 'john@example.com',
      groups: ['group1', 'group2'],
    });
    await GetBasicAuthModule(req, res as any, vars);
    mock.restore();
  });

})