import { useEmbeddedVariable } from "./Configuration";

export function useTheme() {
    return useEmbeddedVariable("theme-name");
}
