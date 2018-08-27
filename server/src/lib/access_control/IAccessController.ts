
export interface IAccessController {
  isWhitelisted(domain: string, ip: string): boolean;
  isAccessAllowed(domain: string, resource: string, user: string, groups: string[]): boolean;
}