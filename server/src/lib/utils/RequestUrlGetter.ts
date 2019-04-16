import Constants = require("../constants");
import Express = require("express");
import GetHeader from "./GetHeader";
import HasHeader from "./HasHeader";

export default class RequestUrlGetter {
  static getOriginalUrl(req: Express.Request): string {

    if (HasHeader(req, Constants.HEADER_X_ORIGINAL_URL)) {
      return GetHeader(req, Constants.HEADER_X_ORIGINAL_URL);
    }

    // X-Forwarded-Port is not mandatory since the port is included in X-Forwarded-Host
    // at least in nginx and Traefik.
    const proto = GetHeader(req, Constants.HEADER_X_FORWARDED_PROTO);
    const host = GetHeader(req, Constants.HEADER_X_FORWARDED_HOST);
    const uri = GetHeader(req, Constants.HEADER_X_FORWARDED_URI);

    if (!proto || !host) {
      throw new Error("Missing headers holding requested URL. Requires either X-Original-Url or X-Forwarded-Proto, X-Forwarded-Host and X-Forwarded-Uri.");
    }

    if (!uri) {
      return `${proto}://${host}`;
    }

    return `${proto}://${host}${uri}`;
  }
}
