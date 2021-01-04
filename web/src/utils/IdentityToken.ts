import queryString from "query-string";

export function extractIdentityToken(locationSearch: string) {
    const queryParams = queryString.parse(locationSearch);
    return queryParams && "token" in queryParams ? (queryParams["token"] as string) : null;
}
