import * as BluebirdPromise from "bluebird";

export interface INotifier {
  notify(to: string, subject: string, link: string): BluebirdPromise<void>;
}