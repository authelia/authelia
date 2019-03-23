
export class LdapSearchError extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "LdapSearchError";
    (<any>Object).setPrototypeOf(this, LdapSearchError.prototype);
  }
}

export class LdapBindError extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "LdapBindError";
    (<any>Object).setPrototypeOf(this, LdapBindError.prototype);
  }
}

export class LdapError extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "LdapError";
    (<any>Object).setPrototypeOf(this, LdapError.prototype);
  }
}

export class IdentityError extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "IdentityError";
    (<any>Object).setPrototypeOf(this, IdentityError.prototype);
  }
}

export class AccessDeniedError extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "AccessDeniedError";
    (<any>Object).setPrototypeOf(this, AccessDeniedError.prototype);
  }
}

export class AuthenticationRegulationError extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "AuthenticationRegulationError";
    (<any>Object).setPrototypeOf(this, AuthenticationRegulationError.prototype);
  }
}

export class InvalidTOTPError extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "InvalidTOTPError";
    (<any>Object).setPrototypeOf(this, InvalidTOTPError.prototype);
  }
}

export class NotAuthenticatedError extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "NotAuthenticatedError";
    (<any>Object).setPrototypeOf(this, NotAuthenticatedError.prototype);
  }
}

export class NotAuthorizedError extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "NotAuthenticatedError";
    (<any>Object).setPrototypeOf(this, NotAuthorizedError.prototype);
  }
}

export class FirstFactorValidationError extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "FirstFactorValidationError";
    (<any>Object).setPrototypeOf(this, FirstFactorValidationError.prototype);
  }
}

export class SecondFactorValidationError extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "SecondFactorValidationError";
    (<any>Object).setPrototypeOf(this, FirstFactorValidationError.prototype);
  }
}