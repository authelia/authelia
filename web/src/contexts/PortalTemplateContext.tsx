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

function deepMerge<T>(target: T, source: Partial<T>): T {
    if (source === undefined || source === null) {
        return target;
    }

    const output: any = Array.isArray(target) ? [...(target as any)] : { ...(target as any) };
    Object.keys(source).forEach((key) => {
        const typedKey = key as keyof T;
        const sourceValue = (source as any)[typedKey];
        if (sourceValue === undefined) {
            return;
        }

        const targetValue = (output as any)[typedKey];
        if (
            sourceValue &&
            typeof sourceValue === "object" &&
            !Array.isArray(sourceValue) &&
            targetValue &&
            typeof targetValue === "object" &&
            !Array.isArray(targetValue)
        ) {
            (output as any)[typedKey] = deepMerge(targetValue, sourceValue);
        } else {
            (output as any)[typedKey] = sourceValue;
        }
    });

    return output;
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
    manifest.forEach((entry) => {
        if (!entry || typeof entry !== "object") {
            return;
        }
        const { name, displayName, description, interactive } = entry as PortalTemplateSummary;
        if (typeof name !== "string" || typeof displayName !== "string" || typeof description !== "string") {
            return;
        }
        summaries.push({ name, displayName, description, interactive });
    });

    return summaries.length > 0 ? summaries : null;
}

const hasStorage = () => typeof window !== "undefined" && typeof window.localStorage !== "undefined";

const getStoredTemplate = (): string | null => {
    if (!hasStorage()) {
        return null;
    }

    try {
        return window.localStorage.getItem(STORAGE_KEY);
    } catch (error) {
        console.warn("Failed to read stored portal template preference.", error);
        return null;
    }
};

const setStoredTemplate = (template: string | null) => {
    if (!hasStorage()) {
        return;
    }

    try {
        if (!template) {
            window.localStorage.removeItem(STORAGE_KEY);
        } else {
            window.localStorage.setItem(STORAGE_KEY, template);
        }
    } catch (error) {
        console.warn("Failed to persist portal template preference.", error);
    }
};

const mergeManifestWithDefault = (manifest: PortalTemplateSummary[]): PortalTemplateSummary[] => {
    const merged: PortalTemplateSummary[] = [];
    const seen = new Set<string>();

    const addEntry = (entry: PortalTemplateSummary) => {
        const key = entry.name.toLowerCase();
        if (!seen.has(key)) {
            seen.add(key);
            merged.push(entry);
        }
    };

    defaultTemplateManifest.forEach(addEntry);
    manifest.forEach(addEntry);

    return merged;
};

const resolveCandidateFromManifest = (
    manifest: PortalTemplateSummary[],
    candidate?: string | null,
): PortalTemplateName | null => {
    if (!candidate) {
        return null;
    }

    const normalized = candidate.toLowerCase();
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

const applyTemplateStyle = async (templateName: string) => {
    if (typeof document === "undefined") {
        return;
    }

    const disableAll = () => {
        templateStyleCache.forEach((element) => {
            element.disabled = true;
            element.media = "print";
        });
    };

    disableAll();

    const existing = templateStyleCache.get(templateName);

    try {
        const response = await fetch(`./static/branding/templates/${templateName}/style.css`, {
            cache: "no-store",
        });

        if (!response.ok) {
            if (response.status !== 404) {
                console.warn(`Failed to load CSS for portal template '${templateName}' (${response.status}).`);
            }

            if (existing) {
                existing.disabled = false;
                existing.media = "all";
            }

            return;
        }

        const css = await response.text();
        let element = existing;

        if (!element) {
            element = document.createElement("style");
            element.type = "text/css";
            element.setAttribute("data-portal-template-style", templateName);
            templateStyleCache.set(templateName, element);
            document.head.appendChild(element);
        }

        element.textContent = css;
        element.disabled = false;
        element.media = "all";
    } catch (error) {
        if (existing) {
            existing.disabled = false;
            existing.media = "all";
        }
        console.warn(`Failed to apply CSS for portal template '${templateName}'`, error);
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

    const loadDefinition = useCallback(async (templateName: PortalTemplateName): Promise<PortalTemplateDefinition> => {
        const definitionJson = await fetchJSON(`./static/branding/templates/${templateName}/definition.json`);
        const baseDefinition: PortalTemplateDefinition =
            definitionJson && definitionJson.style
                ? {
                      ...definitionJson,
                      name: definitionJson.name ?? templateName,
                      interactive: definitionJson.interactive,
                  }
                : (portalTemplates[templateName] ?? portalTemplates[DEFAULT_TEMPLATE]);

        let mergedDefinition = baseDefinition;

        const overrides = await fetchJSON(`./static/branding/templates/${templateName}/config.json`);
        if (overrides) {
            try {
                const overrideStyle = overrides.style ?? overrides;
                mergedDefinition = {
                    ...baseDefinition,
                    style: deepMerge(baseDefinition.style, overrideStyle ?? {}),
                };
            } catch (error) {
                console.error(`Failed to merge overrides for template '${templateName}'`, error);
                mergedDefinition = baseDefinition;
            }
        }

        return mergedDefinition;
    }, []);

    useEffect(() => {
        let isMounted = true;

        const load = async () => {
            const manifestData = await fetchJSON("./static/branding/templates/manifest.json");
            const manifest = mergeManifestWithDefault(sanitizeManifest(manifestData) ?? defaultTemplateManifest);

            const baseConfig = await fetchPortalTemplateConfig();
            const storedTemplate = getStoredTemplate();
            const { name: templateName, persist } = chooseTemplate(manifest, storedTemplate, baseConfig?.template);
            const definition = await loadDefinition(templateName);

            if (isMounted && mounted.current) {
                setState({
                    template: templateName,
                    definition,
                    allowSwitcher: Boolean(baseConfig?.enableTemplateSwitcher),
                    templates: manifest,
                });

                if (persist) {
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
            const definition = await loadDefinition(resolved);

            if (!mounted.current) {
                return;
            }

            setState((prev) => ({
                ...prev,
                template: resolved,
                definition,
            }));

            setStoredTemplate(resolved);
        },
        [loadDefinition, state.templates],
    );

    useEffect(() => {
        if (!state.template) {
            return;
        }

        void applyTemplateStyle(state.template);
    }, [state.template]);

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
