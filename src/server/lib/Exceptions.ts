
export class LdapSearchError extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "LdapSearchError";
    Object.setPrototypeOf(this, LdapSearchError.prototype);
  }
}

export class LdapBindError extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "LdapBindError";
    Object.setPrototypeOf(this, LdapBindError.prototype);
  }
}

export class IdentityError extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "IdentityError";
    Object.setPrototypeOf(this, IdentityError.prototype);
  }
}

export class AccessDeniedError extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "AccessDeniedError";
    Object.setPrototypeOf(this, AccessDeniedError.prototype);
  }
}

export class AuthenticationRegulationError extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "AuthenticationRegulationError";
    Object.setPrototypeOf(this, AuthenticationRegulationError.prototype);
  }
}

export class InvalidTOTPError extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "InvalidTOTPError";
    Object.setPrototypeOf(this, InvalidTOTPError.prototype);
  }
}

export class DomainAccessDenied extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "DomainAccessDenied";
    Object.setPrototypeOf(this, DomainAccessDenied.prototype);
  }
}

export class FirstFactorValidationError extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "FirstFactorValidationError";
    Object.setPrototypeOf(this, FirstFactorValidationError.prototype);
  }
}

export class SecondFactorValidationError extends Error {
  constructor(message?: string) {
    super(message);
    this.name = "SecondFactorValidationError";
    Object.setPrototypeOf(this, FirstFactorValidationError.prototype);
  }
}