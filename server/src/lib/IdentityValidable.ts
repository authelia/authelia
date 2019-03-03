import Bluebird = require("bluebird");
import Identity = require("../../types/Identity");

// IdentityValidator allows user to go through a identity validation process
// in two steps:
// - Request an operation to be performed (password reset, registration).
// - Confirm operation with email.

export interface IdentityValidable {
  challenge(): string;
  preValidationInit(req: Express.Request): Bluebird<Identity.Identity>;
  postValidationInit(req: Express.Request): Bluebird<void>;

  // Serves a page after identity check request
  preValidationResponse(req: Express.Request, res: Express.Response): void;
  // Serves the page if identity validated
  postValidationResponse(req: Express.Request, res: Express.Response): void;
  mailSubject(): string;
  destinationPath(): string;
}