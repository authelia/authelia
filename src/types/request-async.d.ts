import * as Promise from "bluebird";
import * as request from "request";

declare module "request" {
    export interface RequestAsync extends RequestAPI<Request, CoreOptions, RequiredUriUrl> {
        getAsync(uri: string, options?: RequiredUriUrl): Promise<RequestResponse>;
        getAsync(uri: string): Promise<RequestResponse>;
        getAsync(options: RequiredUriUrl & CoreOptions): Promise<RequestResponse>;

        postAsync(uri: string, options?: CoreOptions): Promise<RequestResponse>;
        postAsync(uri: string): Promise<RequestResponse>;
        postAsync(options: RequiredUriUrl & CoreOptions): Promise<RequestResponse>;
    }
}