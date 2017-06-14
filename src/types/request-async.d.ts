import * as BluebirdPromise from "bluebird";

declare module "request" {
    export interface RequestAPI<TRequest extends Request,
        TOptions extends CoreOptions,
        TUriUrlOptions> {
        getAsync(uri: string, options?: RequiredUriUrl): BluebirdPromise<RequestResponse>;
        getAsync(uri: string): BluebirdPromise<RequestResponse>;
        getAsync(options: RequiredUriUrl & CoreOptions): BluebirdPromise<RequestResponse>;

        postAsync(uri: string, options?: CoreOptions): BluebirdPromise<RequestResponse>;
        postAsync(uri: string): BluebirdPromise<RequestResponse>;
        postAsync(options: RequiredUriUrl & CoreOptions): BluebirdPromise<RequestResponse>;
    }
}