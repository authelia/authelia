
import * as winston from "winston";

export interface ILogger {
    debug: winston.LeveledLogMethod;
}

