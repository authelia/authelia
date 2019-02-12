import Express = require("express");
import { BelongToDomain } from "../../../../shared/BelongToDomain";


export class SafeRedirector {
  private domain: string;

  constructor(domain: string) {
    this.domain = domain;
  }

  redirectOrElse(
    res: Express.Response,
    url: string,
    defaultUrl: string): void {
    if (BelongToDomain(url, this.domain)) {
        res.redirect(url);
      }
      res.redirect(defaultUrl);
  }
}