export function useTheme() {
    const theme = document.body.getAttribute("data-theme-name");
    if (theme === null) {
        throw new Error("No theme detected");
    }

    return theme;
}
