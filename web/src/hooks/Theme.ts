import { useEmbeddedVariable } from "./Configuration";

export function useTheme() {
    return useEmbeddedVariable("theme-name");
}

export function usePrimaryColor() {
    return useEmbeddedVariable("theme-primarycolor");
}

export function useSecondaryColor() {
    return useEmbeddedVariable("theme-secondarycolor");
}
