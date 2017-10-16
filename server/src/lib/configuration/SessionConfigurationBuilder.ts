
import ExpressSession = require("express-session");
import { AppConfiguration } from "./Configuration";
import { GlobalDependencies } from "../../../types/Dependencies";
import Redis = require("redis");

export class SessionConfigurationBuilder {

  static build(configuration: AppConfiguration, deps: GlobalDependencies): ExpressSession.SessionOptions {
    const sessionOptions: ExpressSession.SessionOptions = {
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
      let redisOptions;
      if (configuration.session.redis.host
        && configuration.session.redis.port) {
        const client = Redis.createClient({
          host: configuration.session.redis.host,
          port: configuration.session.redis.port
        });
        client.on("error", function (err: Error) {
          console.error("Redis error:", err);
        });
        redisOptions = {
          client: client,
          logErrors: true
        };
      }

      if (redisOptions) {
        const RedisStore = deps.ConnectRedis(deps.session);
        sessionOptions.store = new RedisStore(redisOptions);
      }
    }
    return sessionOptions;
  }
}