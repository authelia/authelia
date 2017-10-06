import BluebirdPromise = require("bluebird");
import { GroupsAndEmails } from "./IClient";

export interface IAuthenticator {
  authenticate(username: string, password: string): BluebirdPromise<GroupsAndEmails>;
}