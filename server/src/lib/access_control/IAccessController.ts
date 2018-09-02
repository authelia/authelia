export interface IAccessController {
  isAccessAllowed(domain: string, resource: string, user: string, groups: string[], whitelisted: boolean): boolean;
}