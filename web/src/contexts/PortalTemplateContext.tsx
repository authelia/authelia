import React, { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";

import { PortalTemplateConfiguration } from "@models/PortalTemplateConfiguration";
import {
    PortalTemplateDefinition,
    PortalTemplateName,
    PortalTemplateSummary,
    defaultTemplateManifest,
    portalTemplates,
} from "@themes/portalTemplates";

interface PortalTemplateContextValue {
    template: PortalTemplateName;
    definition: PortalTemplateDefinition;
    allowSwitcher: boolean;
    availableTemplates: PortalTemplateSummary[];
    switchTemplate: (template: PortalTemplateName) => void;
}

interface PortalTemplateState {
    template: PortalTemplateName;
    definition: PortalTemplateDefinition;
    allowSwitcher: boolean;
    templates: PortalTemplateSummary[];
}

type PortalTemplateConfig = PortalTemplateConfiguration;

const DEFAULT_TEMPLATE: PortalTemplateName = "default";
const STORAGE_KEY = "authelia.portal.template";
const noop = () => undefined;
const templateStyleCache = new Map<string, HTMLStyleElement>();

const PortalTemplateContext = createContext<PortalTemplateContextValue>({
    template: DEFAULT_TEMPLATE,
    definition: portalTemplates[DEFAULT_TEMPLATE],
    allowSwitcher: false,
    availableTemplates: defaultTemplateManifest,
    switchTemplate: noop,
});

type MergeableRecord = Record<string, unknown>;

function isMergeableRecord(value: unknown): value is MergeableRecord {
    return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function deepMerge<T extends MergeableRecord>(target: T, source?: Partial<T>): T {
    if (!source) {
        return target;
    }

    const output: MergeableRecord = { ...target };
    for (const [key, sourceValue] of Object.entries(source)) {
        if (sourceValue === undefined) {
            continue;
        }

        const targetValue = output[key];
        if (isMergeableRecord(targetValue) && isMergeableRecord(sourceValue)) {
            output[key] = deepMerge(targetValue, sourceValue as MergeableRecord);
        } else {
            output[key] = sourceValue;
        }
    }

    return output as T;
}

async function fetchJSON(path: string): Promise<any | null> {
    try {
        const response = await fetch(path, { cache: "no-store" });
        if (!response.ok) {
            if (response.status !== 404) {
                console.warn(`Portal template configuration returned ${response.status} for ${path}`);
            }
            return null;
        }

        const text = await response.text();
        if (!text) {
            return null;
        }

        return JSON.parse(text);
    } catch (error) {
        console.error(`Failed to load portal template configuration from ${path}`, error);
        return null;
    }
}

async function fetchPortalTemplateConfig(): Promise<PortalTemplateConfig | null> {
    const sources = ["./api/portal/template", "./static/branding/portal-template.json"];

    for (const source of sources) {
        const config = await fetchJSON(source);
        if (config) {
            if (typeof config.status === "string" && config.data && typeof config.data === "object") {
                return config.data as PortalTemplateConfig;
            }

            return config as PortalTemplateConfig;
        }
    }

    return null;
}

function sanitizeManifest(manifest: unknown): PortalTemplateSummary[] | null {
    if (!Array.isArray(manifest)) {
        return null;
    }

    const summaries: PortalTemplateSummary[] = [];
    for (const entry of manifest) {
        if (entry === null || typeof entry !== "object") {
            continue;
        }
        const { name, displayName, description, interactive, stylePath, definitionPath, effectPath } =
            entry as PortalTemplateSummary;
        if (typeof name !== "string" || typeof displayName !== "string" || typeof description !== "string") {
            continue;
        }
        const sanitized: PortalTemplateSummary = { name, displayName, description };
        if (typeof interactive === "string") {
            sanitized.interactive = interactive;
        }
        if (typeof stylePath === "string" && stylePath.trim().length > 0) {
            sanitized.stylePath = stylePath;
        }
        if (typeof definitionPath === "string" && definitionPath.trim().length > 0) {
            sanitized.definitionPath = definitionPath;
        }
        if (typeof effectPath === "string" && effectPath.trim().length > 0) {
            sanitized.effectPath = effectPath;
        }
        summaries.push(sanitized);
    }

    return summaries.length > 0 ? summaries : null;
}

const getLocalStorage = (): Storage | null => {
    if (typeof globalThis === "undefined" || !("localStorage" in globalThis)) {
        return null;
    }
    return globalThis.localStorage ?? null;
};

const getStoredTemplate = (): string | null => {
    const storage = getLocalStorage();
    if (storage === null) {
        return null;
    }

    try {
        return storage.getItem(STORAGE_KEY);
    } catch (error) {
        console.warn("Failed to read stored portal template preference.", error);
        return null;
    }
};

const setStoredTemplate = (template: string | null) => {
    const storage = getLocalStorage();
    if (storage === null) {
        return;
    }

    try {
        if (template) {
            storage.setItem(STORAGE_KEY, template);
            return;
        }

        storage.removeItem(STORAGE_KEY);
    } catch (error) {
        console.warn("Failed to persist portal template preference.", error);
    }
};

const mergeManifestWithDefault = (manifest: PortalTemplateSummary[]): PortalTemplateSummary[] => {
    const merged: PortalTemplateSummary[] = [];
    const seen = new Set<string>();

    const addEntry = (entry: PortalTemplateSummary) => {
        const key = entry.name.toLowerCase();
        if (seen.has(key)) {
            return;
        }
        seen.add(key);
        merged.push(entry);
    };

    for (const entry of defaultTemplateManifest) {
        addEntry(entry);
    }
    for (const entry of manifest) {
        addEntry(entry);
    }

    return merged;
};

const getDocumentInstance = (): Document | null => {
    if (typeof globalThis === "undefined" || !("document" in globalThis) || !globalThis.document) {
        return null;
    }

    return globalThis.document;
};

const findSummary = (manifest: PortalTemplateSummary[], name: string) =>
    manifest.find((entry) => entry.name.toLowerCase() === name.toLowerCase());

const fetchTemplateCss = async (
    templateName: string,
    summary?: PortalTemplateSummary,
): Promise<{ css: string; path?: string } | null> => {
    const candidates = [summary?.stylePath, `./static/branding/templates/${templateName}/style.css`].filter(
        (value): value is string => Boolean(value && value.trim().length > 0),
    );

    for (const candidate of candidates) {
        try {
            const response = await fetch(candidate, { cache: "no-store" });
            if (!response.ok) {
                if (response.status !== 404) {
                    console.warn(
                        `Failed to load CSS for portal template '${templateName}' from '${candidate}' (${response.status}).`,
                    );
                }
                continue;
            }

            const css = await response.text();
            if (css) {
                return { css, path: candidate };
            }
        } catch (error) {
            console.warn(`Error fetching CSS for portal template '${templateName}' from '${candidate}'.`, error);
        }
    }

    return null;
};

const resolveCandidateFromManifest = (
    manifest: PortalTemplateSummary[],
    candidate?: string | null,
): PortalTemplateName | null => {
    if (candidate === null || candidate === undefined) {
        return null;
    }

    const normalized = candidate.trim().toLowerCase();
    if (!normalized) {
        return null;
    }
    if (normalized === "none") {
        return DEFAULT_TEMPLATE;
    }

    const match = manifest.find((entry) => entry.name.toLowerCase() === normalized);
    if (match) {
        return match.name as PortalTemplateName;
    }

    if (portalTemplates[normalized]) {
        return normalized as PortalTemplateName;
    }

    return null;
};

const chooseTemplate = (
    manifest: PortalTemplateSummary[],
    stored: string | null,
    configured?: string,
): { name: PortalTemplateName; persist: boolean } => {
    const storedResolved = resolveCandidateFromManifest(manifest, stored);
    if (storedResolved) {
        return { name: storedResolved, persist: storedResolved !== stored };
    }

    const configuredResolved = resolveCandidateFromManifest(manifest, configured ?? null);
    if (configuredResolved) {
        return { name: configuredResolved, persist: false };
    }

    return {
        name: resolveCandidateFromManifest(manifest, DEFAULT_TEMPLATE) ?? DEFAULT_TEMPLATE,
        persist: stored !== null,
    };
};

const applyTemplateStyle = async (templateName: string, summary?: PortalTemplateSummary) => {
    const documentInstance = getDocumentInstance();
    if (documentInstance === null) {
        return;
    }

    for (const element of templateStyleCache.values()) {
        element.disabled = true;
        element.media = "print";
    }

    const existing = templateStyleCache.get(templateName);

    try {
        const result = await fetchTemplateCss(templateName, summary);
        if (result === null) {
            if (existing) {
                existing.disabled = false;
                existing.media = "all";
                existing.textContent = "";
            }

            throw new Error(`CSS for portal template '${templateName}' could not be loaded from any candidate path.`);
        }

        let element = existing;

        if (element === undefined || element === null) {
            element = documentInstance.createElement("style");
            element.dataset.portalTemplateStyle = templateName;
            templateStyleCache.set(templateName, element);
            documentInstance.head.appendChild(element);
        }

        element.dataset.portalTemplateStyle = templateName;
        if (result.path) {
            element.dataset.portalTemplateStylePath = result.path;
        } else {
            delete element.dataset.portalTemplateStylePath;
        }

        element.textContent = result.css;
        element.disabled = false;
        element.media = "all";
    } catch (error) {
        if (existing) {
            existing.disabled = false;
            existing.media = "all";
        }
        console.warn(`Failed to apply CSS for portal template '${templateName}'`, error);
        throw error;
    }
};

const applyTemplateWithFallback = async (
    manifest: PortalTemplateSummary[],
    templateName: PortalTemplateName,
    loadDefinition: (name: PortalTemplateName, summary?: PortalTemplateSummary) => Promise<PortalTemplateDefinition>,
): Promise<{ template: PortalTemplateName; definition: PortalTemplateDefinition }> => {
    const summary = findSummary(manifest, templateName);
    const definition = await loadDefinition(templateName, summary);

    try {
        await applyTemplateStyle(templateName, summary);
        return { template: templateName, definition };
    } catch (error) {
        console.error(error);
        const fallbackSummary = findSummary(manifest, DEFAULT_TEMPLATE);
        const fallbackDefinition = await loadDefinition(DEFAULT_TEMPLATE, fallbackSummary);
        try {
            await applyTemplateStyle(DEFAULT_TEMPLATE, fallbackSummary);
        } catch (fallbackError) {
            console.error(fallbackError);
        }
        return { template: DEFAULT_TEMPLATE, definition: fallbackDefinition };
    }
};

export const PortalTemplateProvider = ({ children }: { children: React.ReactNode }) => {
    const [state, setState] = useState<PortalTemplateState>({
        template: DEFAULT_TEMPLATE,
        definition: portalTemplates[DEFAULT_TEMPLATE],
        allowSwitcher: false,
        templates: defaultTemplateManifest,
    });
    const mounted = useRef(true);
    useEffect(() => {
        return () => {
            mounted.current = false;
        };
    }, []);

    const loadDefinition = useCallback(
        async (
            templateName: PortalTemplateName,
            summary?: PortalTemplateSummary,
        ): Promise<PortalTemplateDefinition> => {
            const definitionPath =
                summary?.definitionPath ?? `./static/branding/templates/${templateName}/definition.json`;
            const definitionJson = definitionPath ? await fetchJSON(definitionPath) : null;
            const baseDefinition: PortalTemplateDefinition =
                definitionJson && definitionJson.style
                    ? {
                          ...definitionJson,
                          name: definitionJson.name ?? templateName,
                          interactive: definitionJson.interactive,
                      }
                    : (portalTemplates[templateName] ?? portalTemplates[DEFAULT_TEMPLATE]);

            if (summary?.effectPath) {
                return {
                    ...baseDefinition,
                    effect: {
                        module: summary.effectPath,
                    },
                };
            }

            return baseDefinition;
        },
        [],
    );

    useEffect(() => {
        let isMounted = true;

        const load = async () => {
            const manifestData = await fetchJSON("./static/branding/templates/manifest.json");
            const manifest = mergeManifestWithDefault(sanitizeManifest(manifestData) ?? defaultTemplateManifest);

            const baseConfig = await fetchPortalTemplateConfig();
            const storedTemplate = getStoredTemplate();
            const { name: templateName, persist } = chooseTemplate(manifest, storedTemplate, baseConfig?.template);
            const { template: resolvedTemplate, definition: resolvedDefinition } = await applyTemplateWithFallback(
                manifest,
                templateName,
                loadDefinition,
            );

            if (isMounted && mounted.current) {
                setState({
                    template: resolvedTemplate,
                    definition: resolvedDefinition,
                    allowSwitcher: Boolean(baseConfig?.enableTemplateSwitcher),
                    templates: manifest,
                });

                if (persist && resolvedTemplate === templateName) {
                    setStoredTemplate(templateName);
                }
            }
        };

        load();

        return () => {
            isMounted = false;
        };
    }, [loadDefinition]);

    const switchTemplate = useCallback(
        async (templateName: PortalTemplateName) => {
            const manifest =
                state.templates.length > 0 ? state.templates : mergeManifestWithDefault(defaultTemplateManifest);
            const resolved = resolveCandidateFromManifest(manifest, templateName) ?? DEFAULT_TEMPLATE;
            if (mounted.current === false) {
                return;
            }

            const { template: templateToApply, definition: definitionToApply } = await applyTemplateWithFallback(
                manifest,
                resolved,
                loadDefinition,
            );

            setState((prev) => ({
                ...prev,
                template: templateToApply,
                definition: definitionToApply,
            }));

            setStoredTemplate(templateToApply);
        },
        [loadDefinition, state.templates],
    );

    useEffect(() => {
        if (state.template === undefined || state.template === null) {
            return;
        }

        const summary = state.templates.find((entry) => entry.name.toLowerCase() === state.template.toLowerCase());
        void applyTemplateStyle(state.template, summary).catch((error) => {
            console.error(error);
        });
    }, [state.template, state.templates]);

    const value = useMemo<PortalTemplateContextValue>(
        () => ({
            template: state.template,
            definition: state.definition,
            allowSwitcher: state.allowSwitcher,
            availableTemplates: state.templates,
            switchTemplate,
        }),
        [state, switchTemplate],
    );

    return <PortalTemplateContext.Provider value={value}>{children}</PortalTemplateContext.Provider>;
};

export function usePortalTemplate(): PortalTemplateContextValue {
    return useContext(PortalTemplateContext);
}
