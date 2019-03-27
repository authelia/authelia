import CheckAuthorizations from "./CheckAuthorizations";
import AuthorizerStub from "../../authorization/AuthorizerStub.spec";
import { Level } from "../../authentication/Level";
import { Level as AuthorizationLevel } from "../../authorization/Level";
import * as Assert from "assert";
import { NotAuthenticatedError, NotAuthorizedError } from "../../Exceptions";

describe('routes/verify/CheckAuthorizations', function() {
  describe('bypass policy', function() {
    it('should allow an anonymous user', function() {
      const authorizer = new AuthorizerStub();
      authorizer.authorizationMock.returns(AuthorizationLevel.BYPASS);
      CheckAuthorizations(authorizer, "public.example.com", "/index.html", undefined,
        undefined, "127.0.0.1", Level.NOT_AUTHENTICATED);
    });
  
    it('should allow an authenticated user (1FA)', function() {
      const authorizer = new AuthorizerStub();
      authorizer.authorizationMock.returns(AuthorizationLevel.BYPASS);
      CheckAuthorizations(authorizer, "public.example.com", "/index.html", "john",
        ["group1", "group2"], "127.0.0.1",  Level.ONE_FACTOR);
    });
  
    it('should allow an authenticated user (2FA)', function() {
      const authorizer = new AuthorizerStub();
      authorizer.authorizationMock.returns(AuthorizationLevel.BYPASS);
      CheckAuthorizations(authorizer, "public.example.com", "/index.html", "john",
        ["group1", "group2"], "127.0.0.1",  Level.TWO_FACTOR);
    });
  });

  describe('one_factor policy', function() {
    it('should not allow an anonymous user', function() {
      const authorizer = new AuthorizerStub();
      authorizer.authorizationMock.returns(AuthorizationLevel.ONE_FACTOR);
      Assert.throws(() => { CheckAuthorizations(authorizer, "public.example.com", "/index.html", undefined,
        undefined, "127.0.0.1",  Level.NOT_AUTHENTICATED) }, NotAuthenticatedError);
    });
  
    it('should allow an authenticated user (1FA)', function() {
      const authorizer = new AuthorizerStub();
      authorizer.authorizationMock.returns(AuthorizationLevel.ONE_FACTOR);
      CheckAuthorizations(authorizer, "public.example.com", "/index.html", "john",
        ["group1", "group2"], "127.0.0.1",  Level.ONE_FACTOR);
    });
  
    it('should allow an authenticated user (2FA)', function() {
      const authorizer = new AuthorizerStub();
      authorizer.authorizationMock.returns(AuthorizationLevel.ONE_FACTOR);
      CheckAuthorizations(authorizer, "public.example.com", "/index.html", "john",
        ["group1", "group2"], "127.0.0.1",  Level.TWO_FACTOR);
    });
  });

  describe('two_factor policy', function() {
    it('should not allow an anonymous user', function() {
      const authorizer = new AuthorizerStub();
      authorizer.authorizationMock.returns(AuthorizationLevel.TWO_FACTOR);
      Assert.throws(() => CheckAuthorizations(authorizer, "public.example.com", "/index.html", undefined,
        undefined, "127.0.0.1",  Level.NOT_AUTHENTICATED), NotAuthenticatedError);
    });
  
    it('should not allow an authenticated user (1FA)', function() {
      const authorizer = new AuthorizerStub();
      authorizer.authorizationMock.returns(AuthorizationLevel.TWO_FACTOR);
      Assert.throws(() => CheckAuthorizations(authorizer, "public.example.com", "/index.html", "john",
        ["group1", "group2"], "127.0.0.1",  Level.ONE_FACTOR), NotAuthenticatedError);
    });
  
    it('should allow an authenticated user (2FA)', function() {
      const authorizer = new AuthorizerStub();
      authorizer.authorizationMock.returns(AuthorizationLevel.TWO_FACTOR);
      CheckAuthorizations(authorizer, "public.example.com", "/index.html", "john",
        ["group1", "group2"], "127.0.0.1",  Level.TWO_FACTOR);
    });
  });

  describe('deny policy', function() {
    it('should not allow an anonymous user', function() {
      const authorizer = new AuthorizerStub();
      authorizer.authorizationMock.returns(AuthorizationLevel.DENY);
      Assert.throws(() => CheckAuthorizations(authorizer, "public.example.com", "/index.html", undefined,
        undefined, "127.0.0.1",  Level.NOT_AUTHENTICATED), NotAuthenticatedError);
    });
  
    it('should not allow an authenticated user (1FA)', function() {
      const authorizer = new AuthorizerStub();
      authorizer.authorizationMock.returns(AuthorizationLevel.DENY);
      Assert.throws(() => CheckAuthorizations(authorizer, "public.example.com", "/index.html", "john",
        ["group1", "group2"], "127.0.0.1",  Level.ONE_FACTOR), NotAuthorizedError);
    });
  
    it('should not allow an authenticated user (2FA)', function() {
      const authorizer = new AuthorizerStub();
      authorizer.authorizationMock.returns(AuthorizationLevel.DENY);
      Assert.throws(() => CheckAuthorizations(authorizer, "public.example.com", "/index.html", "john",
        ["group1", "group2"], "127.0.0.1",  Level.TWO_FACTOR), NotAuthorizedError);
    });
  });
});