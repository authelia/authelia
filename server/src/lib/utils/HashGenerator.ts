import BluebirdPromise = require("bluebird");
import RandomString = require("randomstring");
import Util = require("util");
const crypt = require("crypt3");

export class HashGenerator {
  static ssha512(
    password: string,
    rounds: number = 500000,
    salt?: string): BluebirdPromise<string> {
    // $6 means SHA512
    const _salt = Util.format("$6$rounds=%d$%s", rounds,
      (salt) ? salt : RandomString.generate(16));

    const cryptAsync = BluebirdPromise.promisify<string, string, string>(crypt);

    return cryptAsync(password, _salt)
      .then(function (hash: string) {
        return BluebirdPromise.resolve(Util.format("{CRYPT}%s", hash));
      });
  }
}