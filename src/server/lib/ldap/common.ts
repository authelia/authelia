import util = require("util");

import { LdapConfiguration } from "../configuration/Configuration";


export function buildUserDN(username: string, options: LdapConfiguration): string {
  let userNameAttribute = options.user_name_attribute;
  // if not provided, default to cn
  if (!userNameAttribute) userNameAttribute = "cn";

  const additionalUserDN = options.additional_user_dn;
  const base_dn = options.base_dn;

  let userDN = util.format("%s=%s", userNameAttribute, username);
  if (additionalUserDN) userDN += util.format(",%s", additionalUserDN);
  userDN += util.format(",%s", base_dn);
  return userDN;
}