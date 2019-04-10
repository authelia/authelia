import * as Express from "express";
import * as ExpressMock from "../../stubs/express.spec";
import * as Sinon from "sinon";
import * as Assert from "assert";
import CheckInactivity from "./CheckInactivity";
import { AuthenticationSession } from "../../../../types/AuthenticationSession";
import { Configuration } from "../../configuration/schema/Configuration";
import { RequestLoggerStub } from "../../logging/RequestLoggerStub.spec";
import { Level } from "../../authentication/Level";


describe('routes/verify/VerifyInactivity', function() {
  let req: Express.Request;
  let authSession: AuthenticationSession;
  let configuration: Configuration;
  let logger: RequestLoggerStub;

  beforeEach(function() {
    req = ExpressMock.RequestMock();
    authSession = {
      authentication_level: Level.TWO_FACTOR,
    } as any;
    configuration = {
      session: {
        domain: 'example.com',
        secret: 'abc',
        inactivity: 1000,
      },
      authentication_backend: {
        file: {
          path: 'abc'
        }
      }
    }
    logger = new RequestLoggerStub();
  });

  it('should not throw if user is not authenticated', function() {
    authSession.authentication_level = Level.NOT_AUTHENTICATED;
    CheckInactivity(req, authSession, configuration, logger);
  });

  it('should not throw if inactivity timeout is disabled', function() {
    delete configuration.session.inactivity;
    CheckInactivity(req, authSession, configuration, logger);
  });

  it('should not throw if keep me logged in has been checked', function() {
    authSession.keep_me_logged_in = true;
    CheckInactivity(req, authSession, configuration, logger);
  });

  it('should not throw if the inactivity timeout has not timed out', function() {
    this.clock = Sinon.useFakeTimers();
    authSession.last_activity_datetime = new Date().getTime();
    this.clock.tick(200);
    CheckInactivity(req, authSession, configuration, logger);
    this.clock.restore();
  });

  it('should throw if the inactivity timeout has timed out', function() {
    this.clock = Sinon.useFakeTimers();
    authSession.last_activity_datetime = new Date().getTime();
    this.clock.tick(2000);
    Assert.throws(() => CheckInactivity(req, authSession, configuration, logger));
    this.clock.restore();
  });
});