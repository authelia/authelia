import { Level } from "./Level";
import { Subject } from "./Subject";
import { Object } from "./Object";

export interface IAuthorizer {
  authorization(object: Object, subject: Subject, ip: string): Level;
}