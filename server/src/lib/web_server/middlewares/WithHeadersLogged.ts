import Express = require("express");
import { IRequestLogger } from "../../logging/IRequestLogger";

export class WithHeadersLogged {
  static middleware(logger: IRequestLogger) {
    return function (req: Express.Request, res: Express.Response,
      next: Express.NextFunction): void {
      logger.debug(req, "Headers = %s", JSON.stringify(req.headers));
      next();
    };
  }
}