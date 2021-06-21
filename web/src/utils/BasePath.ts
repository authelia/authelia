import { getEmbeddedVariable } from "@utils/Configuration";

export function getBasePath() {
    return getEmbeddedVariable("basepath");
}
