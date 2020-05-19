
export function useBasePath() {
    const bodyElements = document.getElementsByTagName("body");
    if (bodyElements.length !== 1) {
        throw new Error("No body detected");
    }

    const body = bodyElements[0];

    const basePath = body.getAttribute("data-basepath");
    if (basePath === null) {
        throw new Error("No base path detected");
    }

    return basePath;
}