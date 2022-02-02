export function makeParams(params: Record<string, string | undefined | null>): string {
    const stringifiedParams = Object.entries(params)
        .filter(([, value]) => !!value)
        .map(([key, value]) => `${key}=${value}`)
        .join("&");
    return stringifiedParams ? `?${stringifiedParams}` : "";
}
