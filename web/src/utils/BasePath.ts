import { getEmbeddedVariable } from "./Configuration";

export function getBasePath() {
    return getEmbeddedVariable("basepath");
}