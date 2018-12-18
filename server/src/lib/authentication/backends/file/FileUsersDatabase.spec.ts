import Assert = require("assert");
import Bluebird = require("bluebird");
import Fs = require("fs");
import Sinon = require("sinon");
import Tmp = require("tmp");

import { FileUsersDatabase } from "./FileUsersDatabase";
import { FileUsersDatabaseConfiguration } from "../../../configuration/schema/FileUsersDatabaseConfiguration";
import { HashGenerator } from "../../../utils/HashGenerator";

const GOOD_DATABASE = `
users:
  john:
    password: "{CRYPT}$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev

  harry:
    password: "{CRYPT}$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    emails: harry.potter@authelia.com
    groups: []
`;

const BAD_HASH = `
users:
  john:
    password: "{CRYPT}$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
`;

const NO_PASSWORD_DATABASE = `
users:
  john:
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
`;

const NO_EMAIL_DATABASE = `
users:
  john:
    password: "{CRYPT}$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    groups:
      - admins
      - dev
`;

const SINGLE_USER_DATABASE = `
users:
  john:
    password: "{CRYPT}$6$rounds=500000$jgiCMRyGXzoqpxS3$w2pJeZnnH8bwW3zzvoMWtTRfQYsHbWbD/hquuQ5vUeIyl9gdwBIt6RWk2S6afBA0DPakbeWgD/4SZPiS0hYtU/"
    email: john.doe@authelia.com
    groups:
      - admins
      - dev
`

function createTmpFileFrom(yaml: string) {
  const tmpFileAsync = Bluebird.promisify(Tmp.file);
  return tmpFileAsync()
    .then((path: string) => {
      Fs.writeFileSync(path, yaml, "utf-8");
      return Bluebird.resolve(path);
    });
}

describe("authentication/backends/file/FileUsersDatabase", function() {
  let configuration: FileUsersDatabaseConfiguration;

  describe("checkUserPassword", () => {
    describe("good config", () => {
      beforeEach(() => {
        return createTmpFileFrom(GOOD_DATABASE)
          .then((path: string) => configuration = {
            path: path
          });
      });

      it("should succeed", () => {
          const usersDatabase = new FileUsersDatabase(configuration);
          return usersDatabase.checkUserPassword("john", "password")
            .then((groupsAndEmails) => {
              Assert.deepEqual(groupsAndEmails.groups, ["admins", "dev"]);
              Assert.deepEqual(groupsAndEmails.emails, ["john.doe@authelia.com"]);
            });
      });

      it("should fail when password is wrong", () => {
          const usersDatabase = new FileUsersDatabase(configuration);
          return usersDatabase.checkUserPassword("john", "bad_password")
            .then(() => Bluebird.reject(new Error("should not be here.")))
            .catch((err) => {
              return Bluebird.resolve();
            });
      });

      it("should fail when user does not exist", () => {
        const usersDatabase = new FileUsersDatabase(configuration);
        return usersDatabase.checkUserPassword("no_user", "password")
          .then(() => Bluebird.reject(new Error("should not be here.")))
          .catch((err) => {
            return Bluebird.resolve();
          });
      });
    });

    describe("bad hash", () => {
      beforeEach(() => {
        return createTmpFileFrom(GOOD_DATABASE)
          .then((path: string) => configuration = {
            path: path
          });
      });

      it("should fail when hash is wrong", () => {
        const usersDatabase = new FileUsersDatabase(configuration);
        return usersDatabase.checkUserPassword("john", "password")
          .then(() => Bluebird.reject(new Error("should not be here.")))
          .catch((err) => {
            return Bluebird.resolve();
          });
      });
    });

    describe("no password", () => {
      beforeEach(() => {
        return createTmpFileFrom(NO_PASSWORD_DATABASE)
          .then((path: string) => configuration = {
            path: path
          });
      });

      it("should fail", () => {
        const usersDatabase = new FileUsersDatabase(configuration);
        return usersDatabase.checkUserPassword("john", "password")
          .then(() => Bluebird.reject(new Error("should not be here.")))
          .catch((err) => {
            return Bluebird.resolve();
          });
      });
    });
  });

  describe("getEmails", () => {
    describe("good config", () => {
      beforeEach(() => {
        return createTmpFileFrom(GOOD_DATABASE)
          .then((path: string) => configuration = {
            path: path
          });
      });

      it("should succeed", () => {
        const usersDatabase = new FileUsersDatabase(configuration);
        return usersDatabase.getEmails("john")
          .then((emails) => {
            Assert.deepEqual(emails, ["john.doe@authelia.com"]);
          });
      });

      it("should fail when user does not exist", () => {
        const usersDatabase = new FileUsersDatabase(configuration);
        return usersDatabase.getEmails("no_user")
          .then(() => Bluebird.reject(new Error("should not be here.")))
          .catch((err) => {
            return Bluebird.resolve();
          });
      });
    });

    describe("no email provided", () => {
      beforeEach(() => {
        return createTmpFileFrom(NO_EMAIL_DATABASE)
          .then((path: string) => configuration = {
            path: path
          });
      });

      it("should fail", () => {
        const usersDatabase = new FileUsersDatabase(configuration);
        return usersDatabase.getEmails("john")
          .then(() => Bluebird.reject(new Error("should not be here.")))
          .catch((err) => {
            return Bluebird.resolve();
          });
      });
    });
  });

  describe("updatePassword", () => {
    beforeEach(() => {
      return createTmpFileFrom(SINGLE_USER_DATABASE)
        .then((path: string) => configuration = {
          path: path
        });
    });

    it("should succeed", () => {
      const usersDatabase = new FileUsersDatabase(configuration);
      const NEW_HASH = "{CRYPT}$6$rounds=500000$Qw6MhgADvLyYMEq9$ABCDEFGHIJKLMNOPQRSTUVWXYZ";
      const stub = Sinon.stub(HashGenerator, "ssha512").returns(Bluebird.resolve(NEW_HASH));
      return usersDatabase.updatePassword("john", "mypassword")
        .then(() => {
          const content = Fs.readFileSync(configuration.path, "utf-8");
          const matches = content.match(/password: '(.+)'/);
          Assert.equal(matches[1], NEW_HASH);
        })
        .finally(() => stub.restore());
    });

    it("should fail when user does not exist", () => {
      const usersDatabase = new FileUsersDatabase(configuration);
      return usersDatabase.updatePassword("bad_user", "mypassword")
        .then(() => Bluebird.reject(new Error("should not be here")))
        .catch(() => Bluebird.resolve());
    });
  });
});