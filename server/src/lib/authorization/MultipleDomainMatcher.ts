
export class MultipleDomainMatcher {
  static match(domain: string, pattern: string): boolean {
    if (pattern.startsWith("*") &&
      domain.endsWith(pattern.substr(1))) {
      return true;
    }
    else if (domain == pattern) {
      return true;
    }
  }
}