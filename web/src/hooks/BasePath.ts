import { useEmbeddedVariable } from "./Configuration";

export function useBasePath() {
    return useEmbeddedVariable("basepath");
}