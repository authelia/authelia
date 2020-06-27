import { useEmbeddedVariable } from "./Configuration";

export function useTheme() {
    return useEmbeddedVariable("theme-name");
}

export function useMainColor() {
    return useEmbeddedVariable("theme-maincolor");
}

export function useSecondaryColor() {
    return useEmbeddedVariable("theme-secondarycolor");
}
