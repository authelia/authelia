import ExpressSession = require("express-session");
import Redis = require("redis");

import { Configuration } from "./schema/Configuration";
import { GlobalDependencies } from "../../../types/Dependencies";
import { RedisStoreOptions } from "connect-redis";

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
      let redisOptions;
      const options: Redis.ClientOpts = {
        host: configuration.session.redis.host,
        port: configuration.session.redis.port
      };

      if (configuration.session.redis.password) {
        options["password"] = configuration.session.redis.password;
      }
      const client = deps.Redis.createClient(options);

      client.on("error", function (err: Error) {
        console.error("Redis error:", err);
      });

      redisOptions = {
        client: client,
        logErrors: true
      };

      if (redisOptions) {
        const RedisStore = deps.ConnectRedis(deps.session);
        sessionOptions.store = new RedisStore(redisOptions);
      }
    }
    return sessionOptions;
  }
}