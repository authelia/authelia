import ExpressSession = require("express-session");
import { Configuration } from "./schema/Configuration";
import { GlobalDependencies } from "../../../types/Dependencies";

export class SessionConfigurationBuilder {
  static build(configuration: Configuration, deps: GlobalDependencies): ExpressSession.SessionOptions {
    const sessionOptions: ExpressSession.SessionOptions = {
      name: configuration.session.name,
      secret: configuration.session.secret,
      resave: false,
      saveUninitialized: true,
      cookie: {
        secure: true,
        httpOnly: true,
        maxAge: configuration.session.expiration,
        domain: configuration.session.domain
      },
    };

    if (configuration.session.redis) {
      const RedisStore = deps.ConnectRedis(deps.session);
      sessionOptions.store = new RedisStore({
        host: configuration.session.redis.host,
        port: configuration.session.redis.port,
        pass: configuration.session.redis.password,
        logErrors: true
      });
    }
    return sessionOptions;
  }
}