import { Level } from "./Level";

export interface IAuthorizer {
  authorization(domain: string, resource: string, user: string, groups: string[]): Level;
}