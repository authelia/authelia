
module.exports = {
  LdapSearchError: LdapSearchError,
  LdapBindError: LdapBindError,
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
