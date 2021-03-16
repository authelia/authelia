export function useRedirector() {
    return (url: string) => {
        window.location.href = url;
    };
}
