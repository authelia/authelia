
import ExpressSession = require("express-session");
import { AppConfiguration } from "../../types/Configuration";
import { GlobalDependencies } from "../../types/Dependencies";

export default class SessionConfigurationBuilder {

    static build(configuration: AppConfiguration, deps: GlobalDependencies): ExpressSession.SessionOptions {
        const sessionOptions: ExpressSession.SessionOptions = {
            secret: configuration.session.secret,
            resave: false,
            saveUninitialized: true,
            cookie: {
                secure: false,
                maxAge: configuration.session.expiration,
                domain: configuration.session.domain
            },
        };

        if (configuration.session.redis) {
            let redisOptions;
            if (configuration.session.redis.host
                && configuration.session.redis.port) {
                redisOptions = {
                    host: configuration.session.redis.host,
                    port: configuration.session.redis.port
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