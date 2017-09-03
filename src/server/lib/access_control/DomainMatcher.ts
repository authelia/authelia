
export class DomainMatcher {
  static match(domain: string, allowedDomain: string): boolean {
    if (allowedDomain.startsWith("*") &&
      domain.endsWith(allowedDomain.substr(1))) {
      return true;
    }
    else if (domain == allowedDomain) {
      return true;
    }
  }
}