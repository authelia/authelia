
module.exports = {
  LdapSearchError: LdapSearchError,
  LdapBindError: LdapBindError,
  IdentityError: IdentityError,
  AccessDeniedError: AccessDeniedError
}

function LdapSearchError(message) {
  this.name = "LdapSearchError";
  this.message = (message || "");
}
LdapSearchError.prototype = Object.create(Error.prototype);

function LdapBindError(message) {
  this.name = "LdapBindError";
  this.message = (message || "");
}
LdapBindError.prototype = Object.create(Error.prototype);

function IdentityError(message) {
  this.name = "IdentityError";
  this.message = (message || "");
}
IdentityError.prototype = Object.create(Error.prototype);

function AccessDeniedError(message) {
  this.name = "AccessDeniedError";
  this.message = (message || "");
}
AccessDeniedError.prototype = Object.create(Error.prototype);
