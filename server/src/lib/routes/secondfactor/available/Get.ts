import * as Express from "express";
import { ServerVariables } from "../../../ServerVariables";
import Method2FA from "../../../Method2FA";


export default function(vars: ServerVariables) {
  return async function(_: Express.Request, res: Express.Response) {
    const availableMethods: Method2FA[] = ["u2f", "totp"];
    if (vars.config.duo_api) {
      availableMethods.push("duo_push");
    }
    res.json(availableMethods);
  };
}