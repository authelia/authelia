
import * as speakeasy from "speakeasy";
import { Speakeasy } from "../types/Dependencies";
import BluebirdPromise = require("bluebird");

export default class TOTPGenerator {
  private speakeasy: Speakeasy;

  constructor(speakeasy: Speakeasy) {
    this.speakeasy = speakeasy;
  }

  generate(options: speakeasy.GenerateOptions): speakeasy.Key {
    return this.speakeasy.generateSecret(options);
  }
}