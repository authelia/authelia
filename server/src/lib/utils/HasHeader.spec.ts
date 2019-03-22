import * as Express from "express";
import HasHeader from "./HasHeader";
import { RequestMock } from "../stubs/express.spec";
import * as Assert from "assert";

describe('utils/HasHeader', function() {
  let req: Express.Request;
  beforeEach(() => {
    req = RequestMock();
  });

  it('should return the header if it exists', function() {
    req.headers["x-target-url"] = 'www.example.com';
    Assert(HasHeader(req, 'x-target-url'));
  });

  it('should return undefined if header does not exist', function() {
    Assert(!HasHeader(req, 'x-target-url'));
  });
});