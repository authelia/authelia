export function useBasePath() {
    const basePath = document.body.getAttribute("data-basepath");
    if (basePath === null) {
        throw new Error("No base path detected");
    }

    return basePath;
}