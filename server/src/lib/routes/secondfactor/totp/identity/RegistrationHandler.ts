
import * as Express from "express";
import BluebirdPromise = require("bluebird");

import { Identity } from "../../../../../../types/Identity";
import { IdentityValidable } from "../../../../IdentityValidable";
import Constants = require("../constants");
import ErrorReplies = require("../../../../ErrorReplies");
import { AuthenticationSessionHandler } from "../../../../AuthenticationSessionHandler";
import UserMessages = require("../../../../UserMessages");
import FirstFactorValidator = require("../../../../FirstFactorValidator");
import { IRequestLogger } from "../../../../logging/IRequestLogger";
import { IUserDataStore } from "../../../../storage/IUserDataStore";
import { ITotpHandler } from "../../../../authentication/totp/ITotpHandler";
import { TOTPSecret } from "../../../../../../types/TOTPSecret";
import { TotpConfiguration } from "../../../../configuration/schema/TotpConfiguration";


export default class RegistrationHandler implements IdentityValidable {
  private logger: IRequestLogger;
  private userDataStore: IUserDataStore;
  private totp: ITotpHandler;
  private configuration: TotpConfiguration;

  constructor(logger: IRequestLogger,
    userDataStore: IUserDataStore,
    totp: ITotpHandler, configuration: TotpConfiguration) {
    this.logger = logger;
    this.userDataStore = userDataStore;
    this.totp = totp;
    this.configuration = configuration;
  }

  challenge(): string {
    return Constants.CHALLENGE;
  }

  private retrieveIdentity(req: Express.Request): BluebirdPromise<Identity> {
    const that = this;
    return new BluebirdPromise(function (resolve, reject) {
      const authSession = AuthenticationSessionHandler.get(req, that.logger);
      const userid = authSession.userid;
      const email = authSession.email;

      if (!(userid && email)) {
        return reject(new Error("User ID or email is missing"));
      }

      const identity = {
        email: email,
        userid: userid
      };
      return resolve(identity);
    });
  }

  preValidationInit(req: Express.Request): BluebirdPromise<Identity> {
    const that = this;
    return FirstFactorValidator.validate(req, this.logger)
      .then(function () {
        return that.retrieveIdentity(req);
      });
  }

  preValidationResponse(req: Express.Request, res: Express.Response) {
    res.json({message: "OK"});
  }

  postValidationInit(req: Express.Request) {
    return FirstFactorValidator.validate(req, this.logger);
  }

  postValidationResponse(req: Express.Request, res: Express.Response)
    : BluebirdPromise<void> {
    const that = this;
    let secret: TOTPSecret;
    let userId: string;
    return new BluebirdPromise(function (resolve, reject) {
      const authSession = AuthenticationSessionHandler.get(req, that.logger);
      userId = authSession.userid;

      if (authSession.identity_check.challenge != Constants.CHALLENGE
        || !userId)
        return reject(new Error("Bad challenge."));

      resolve();
    })
      .then(function () {
        secret = that.totp.generate(userId,
          that.configuration.issuer);
        that.logger.debug(req, "Save the TOTP secret in DB");
        return that.userDataStore.saveTOTPSecret(userId, secret);
      })
      .then(function () {
        AuthenticationSessionHandler.reset(req);

        res.json({
          base32_secret: secret.base32,
          otpauth_url: secret.otpauth_url,
        });
      })
      .catch(ErrorReplies.replyWithError200(req, res, that.logger, UserMessages.OPERATION_FAILED));
  }

  mailSubject(): string {
    return "Set up Authelia's one-time password";
  }

  destinationPath(): string {
    return "/one-time-password-registration";
  }
}