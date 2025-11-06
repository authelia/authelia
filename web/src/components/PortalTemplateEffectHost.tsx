import React, { useEffect, useRef } from "react";

import { usePortalTemplate } from "@contexts/PortalTemplateContext";
import { PortalTemplateDefinition } from "@themes/portalTemplates";
import { getBasePath } from "@utils/BasePath";

type EffectMountResult = void | (() => void) | Promise<void | (() => void)>;

type EffectModule = {
    mount?: (context: PortalTemplateEffectContext) => EffectMountResult;
    default?: (context: PortalTemplateEffectContext) => EffectMountResult;
};

export type PortalTemplateEffectContext = {
    container: HTMLDivElement;
    definition: PortalTemplateDefinition;
};

type Props = {
    className?: string;
};

function resolveModuleURL(modulePath: string): string {
    if (/^https?:\/\//i.test(modulePath)) {
        return modulePath;
    }

    if (typeof globalThis === "undefined" || !("location" in globalThis) || !globalThis.location) {
        return modulePath;
    }

    const basePath = getBasePath() ?? "";
    const baseURL = new URL(basePath.endsWith("/") ? basePath : `${basePath}/`, globalThis.location.origin);

    return new URL(modulePath, baseURL).toString();
}

const PortalTemplateEffectHost = ({ className }: Props) => {
    const { definition } = usePortalTemplate();
    const hostRef = useRef<HTMLDivElement>(null);
    const effectModulePath = definition.effect?.module ?? null;

    useEffect(() => {
        const host = hostRef.current;
        if (!host) {
            return undefined;
        }

        host.replaceChildren();

        if (!effectModulePath) {
            return () => {
                host.replaceChildren();
            };
        }

        let disposed = false;
        let cleanup: (() => void) | void;

        const resolvedURL = resolveModuleURL(effectModulePath);

        import(/* @vite-ignore */ resolvedURL)
            .then(async (module: EffectModule) => {
                if (disposed) {
                    return;
                }

                const mount = typeof module?.mount === "function" ? module.mount : module?.default;
                if (typeof mount !== "function") {
                    console.warn(`Portal template effect module '${resolvedURL}' does not export a mount function.`);
                    return;
                }

                const result = mount({
                    container: host,
                    definition,
                });

                if (result instanceof Promise) {
                    cleanup = await result;
                } else {
                    cleanup = result;
                }
            })
            .catch((error) => {
                if (!disposed) {
                    console.error(`Failed to load portal template effect module '${resolvedURL}'.`, error);
                }
            });

        return () => {
            disposed = true;
            if (typeof cleanup === "function") {
                try {
                    cleanup();
                } catch (error) {
                    console.error("An error occurred while cleaning up the portal template effect.", error);
                }
            }
            host.replaceChildren();
        };
    }, [definition, effectModulePath]);

    return <div ref={hostRef} className={className} aria-hidden="true" data-portal-effect-host />;
};

export default PortalTemplateEffectHost;
