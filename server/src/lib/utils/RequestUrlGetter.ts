import Constants = require("../../../../shared/constants");
import Express = require("express");
import GetHeader from "./GetHeader";
import HasHeader from "./HasHeader";

export class RequestUrlGetter {
  static getOriginalUrl(req: Express.Request): string {

    if (HasHeader(req, Constants.HEADER_X_ORIGINAL_URL)) {
      return GetHeader(req, Constants.HEADER_X_ORIGINAL_URL);
    }

    const proto = GetHeader(req, Constants.HEADER_X_FORWARDED_PROTO);
    const host = GetHeader(req, Constants.HEADER_X_FORWARDED_HOST);
    const port = GetHeader(req, Constants.HEADER_X_FORWARDED_PORT);
    const uri = GetHeader(req, Constants.HEADER_X_FORWARDED_URI);

    if (!proto || !host || !port) {
      throw new Error("Missing headers holding requested URL. Requires X-Original-Url or X-Forwarded-Proto, X-Forwarded-Host, and X-Forwarded-Port")
    }

    return "${proto}://${host}:${port}${uri}";

  }
}
