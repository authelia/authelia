import * as BluebirdPromise from "bluebird";
import * as request from "request";

declare module "request" {
    export interface RequestAsync extends RequestAPI<Request, CoreOptions, RequiredUriUrl> {
        getAsync(uri: string, options?: RequiredUriUrl): BluebirdPromise<RequestResponse>;
        getAsync(uri: string): BluebirdPromise<RequestResponse>;
        getAsync(options: RequiredUriUrl & CoreOptions): BluebirdPromise<RequestResponse>;

        postAsync(uri: string, options?: CoreOptions): BluebirdPromise<RequestResponse>;
        postAsync(uri: string): BluebirdPromise<RequestResponse>;
        postAsync(options: RequiredUriUrl & CoreOptions): BluebirdPromise<RequestResponse>;
    }
}