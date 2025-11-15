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

const globalScope: typeof globalThis | undefined = (() => {
    try {
        return globalThis;
    } catch {
        return undefined;
    }
})();

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

        if (sourceValue === null) {
            output[key] = null;
            continue;
        }

        const targetValue = output[key];
        if (isMergeableRecord(targetValue) && isMergeableRecord(sourceValue)) {
            output[key] = deepMerge(targetValue, sourceValue);
        } else {
            output[key] = sourceValue;
        }
    }

    return output as T;
}

async function fetchJSON<T>(path: string): Promise<T | null> {
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

        return JSON.parse(text) as T;
    } catch (error) {
        console.error(`Failed to load portal template configuration from ${path}`, error);
        return null;
    }
}

type PortalTemplateConfigEnvelope = {
    status: string;
    data?: unknown;
};

const isPortalTemplateConfig = (value: unknown): value is PortalTemplateConfig => {
    if (!value || typeof value !== "object") {
        return false;
    }

    const candidate = value as Record<string, unknown>;
    const templateValue = candidate.template;
    const switcherValue = candidate.enableTemplateSwitcher;

    if (templateValue !== undefined && typeof templateValue !== "string") {
        return false;
    }

    if (switcherValue !== undefined && typeof switcherValue !== "boolean") {
        return false;
    }

    return true;
};

const isPortalTemplateConfigEnvelope = (value: unknown): value is PortalTemplateConfigEnvelope => {
    if (!value || typeof value !== "object") {
        return false;
    }

    return typeof (value as { status?: unknown }).status === "string";
};

async function fetchPortalTemplateConfig(): Promise<PortalTemplateConfig | null> {
    const sources = ["./api/portal/template", "./static/branding/portal-template.json"];

    for (const source of sources) {
        const config = await fetchJSON<unknown>(source);
        if (!config) {
            continue;
        }

        if (isPortalTemplateConfigEnvelope(config) && isPortalTemplateConfig(config.data)) {
            return config.data;
        }

        if (isPortalTemplateConfig(config)) {
            return config;
        }
    }

    return null;
}

const sanitizeManifestEntry = (entry: unknown): PortalTemplateSummary | null => {
    if (!entry || typeof entry !== "object") {
        return null;
    }

    const record = entry as Record<string, unknown>;
    const rawName = typeof record.name === "string" ? record.name.trim() : "";
    const displayName = typeof record.displayName === "string" ? record.displayName.trim() : "";
    const description = typeof record.description === "string" ? record.description.trim() : "";

    if (!rawName || !displayName || !description) {
        return null;
    }

    const summary: PortalTemplateSummary = {
        name: rawName as PortalTemplateName,
        displayName,
        description,
    };

    if (record.interactive === "pointer") {
        summary.interactive = "pointer";
    }

    if (typeof record.stylePath === "string" && record.stylePath.trim().length > 0) {
        summary.stylePath = record.stylePath;
    }

    if (typeof record.definitionPath === "string" && record.definitionPath.trim().length > 0) {
        summary.definitionPath = record.definitionPath;
    }

    if (typeof record.effectPath === "string" && record.effectPath.trim().length > 0) {
        summary.effectPath = record.effectPath;
    }

    return summary;
};

function sanitizeManifest(manifest: unknown): PortalTemplateSummary[] | null {
    if (!Array.isArray(manifest)) {
        return null;
    }

    const summaries: PortalTemplateSummary[] = [];
    for (const entry of manifest) {
        const sanitized = sanitizeManifestEntry(entry);
        if (sanitized) {
            summaries.push(sanitized);
        }
    }

    return summaries.length > 0 ? summaries : null;
}

const getLocalStorage = (): Storage | null => globalScope?.localStorage ?? null;

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

const isPortalTemplateName = (value: string): value is PortalTemplateName =>
    Object.prototype.hasOwnProperty.call(portalTemplates, value);

const getDocumentInstance = (): Document | null => globalScope?.document ?? null;

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
    if (candidate == null) {
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
        return match.name;
    }

    if (isPortalTemplateName(normalized)) {
        return normalized;
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

const applyTemplateStyle = async (
    templateName: string,
    summary: PortalTemplateSummary | undefined,
    signal: AbortSignal | null,
    ownerToken: number,
    ownerRef: React.MutableRefObject<number>,
) => {
    const documentInstance = getDocumentInstance();
    if (documentInstance === null || signal?.aborted || ownerToken !== ownerRef.current) {
        return;
    }

    for (const element of templateStyleCache.values()) {
        if (ownerToken !== ownerRef.current || signal?.aborted) {
            return;
        }
        element.disabled = true;
        element.media = "print";
    }

    const existing = templateStyleCache.get(templateName);

    try {
        const result = await fetchTemplateCss(templateName, summary);
        if (signal?.aborted || ownerToken !== ownerRef.current) {
            return;
        }

        if (result === null) {
            if (signal?.aborted || ownerToken !== ownerRef.current) {
                return;
            }
            throw new Error(`CSS for portal template '${templateName}' could not be loaded from any candidate path.`);
        }

        let element = existing;

        if (element == null) {
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

        if (signal?.aborted || ownerToken !== ownerRef.current) {
            return;
        }

        element.textContent = result.css;
        element.disabled = false;
        element.media = "all";
    } catch (error) {
        if (existing && ownerToken === ownerRef.current && !signal?.aborted) {
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
    signal: AbortSignal | null,
    ownerToken: number,
    ownerRef: React.MutableRefObject<number>,
): Promise<{ template: PortalTemplateName; definition: PortalTemplateDefinition }> => {
    const summary = findSummary(manifest, templateName);
    const definition = await loadDefinition(templateName, summary);

    try {
        await applyTemplateStyle(templateName, summary, signal, ownerToken, ownerRef);
        if (signal?.aborted || ownerToken !== ownerRef.current) {
            return { template: templateName, definition };
        }
        return { template: templateName, definition };
    } catch (error) {
        if (signal?.aborted || ownerToken !== ownerRef.current) {
            throw error;
        }
        console.error(error);
        const fallbackSummary = findSummary(manifest, DEFAULT_TEMPLATE);
        const fallbackDefinition = await loadDefinition(DEFAULT_TEMPLATE, fallbackSummary);
        try {
            await applyTemplateStyle(DEFAULT_TEMPLATE, fallbackSummary, signal, ownerToken, ownerRef);
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
    const styleOwnerRef = useRef(0);
    const styleAbortController = useRef<AbortController | null>(null);

    const loadDefinition = useCallback(
        async (
            templateName: PortalTemplateName,
            summary?: PortalTemplateSummary,
        ): Promise<PortalTemplateDefinition> => {
            const definitionPath =
                summary?.definitionPath ?? `./static/branding/templates/${templateName}/definition.json`;
            const definitionJson = definitionPath ? await fetchJSON<PortalTemplateDefinition>(definitionPath) : null;
            const baseDefinition: PortalTemplateDefinition = definitionJson?.style
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
            const controller = new AbortController();
            styleAbortController.current?.abort();
            styleAbortController.current = controller;
            const ownerToken = styleOwnerRef.current + 1;
            styleOwnerRef.current = ownerToken;
            try {
                const { template: resolvedTemplate, definition: resolvedDefinition } = await applyTemplateWithFallback(
                    manifest,
                    templateName,
                    loadDefinition,
                    controller.signal,
                    ownerToken,
                    styleOwnerRef,
                );

                if (controller.signal.aborted || ownerToken !== styleOwnerRef.current) {
                    return;
                }

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
            } catch (error) {
                if (!controller.signal.aborted && ownerToken === styleOwnerRef.current) {
                    console.error(error);
                }
            } finally {
                if (styleAbortController.current === controller) {
                    styleAbortController.current = null;
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

            const controller = new AbortController();
            styleAbortController.current?.abort();
            styleAbortController.current = controller;
            const ownerToken = styleOwnerRef.current + 1;
            styleOwnerRef.current = ownerToken;
            try {
                const { template: templateToApply, definition: definitionToApply } = await applyTemplateWithFallback(
                    manifest,
                    resolved,
                    loadDefinition,
                    controller.signal,
                    ownerToken,
                    styleOwnerRef,
                );

                if (controller.signal.aborted || ownerToken !== styleOwnerRef.current) {
                    return;
                }

                setState((prev) => ({
                    ...prev,
                    template: templateToApply,
                    definition: definitionToApply,
                }));

                setStoredTemplate(templateToApply);
            } catch (error) {
                if (!controller.signal.aborted && ownerToken === styleOwnerRef.current) {
                    console.error(error);
                }
            } finally {
                if (styleAbortController.current === controller) {
                    styleAbortController.current = null;
                }
            }
        },
        [loadDefinition, state.templates],
    );

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
