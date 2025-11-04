import { CSSProperties } from "react";

export type PortalTemplateName = string;

type KeyframeMap = Record<string, CSSProperties>;

type AnimationConfig = {
    key: string;
    duration: string;
    timingFunction?: string;
    iterationCount?: string;
    direction?: string;
    delay?: string;
};

export interface PortalTemplateEffectDefinition {
    module: string;
}

export type BackgroundLayer = {
    background: string;
    opacity?: number;
    blur?: string;
    mixBlendMode?: CSSProperties["mixBlendMode"];
    position?: CSSProperties["position"];
    inset?: string;
    size?: string;
    zIndex?: number;
    pointerEvents?: CSSProperties["pointerEvents"];
    animation?: AnimationConfig;
    transform?: string;
};

export type PortalStyleConfig = {
    page: {
        background: string;
        color: string;
        before?: BackgroundLayer;
        after?: BackgroundLayer;
    };
    root: {
        padding?: string;
        background?: string;
        border?: string;
        before?: BackgroundLayer;
        after?: BackgroundLayer;
    };
    card: {
        background: string;
        border?: string;
        borderRadius: string;
        padding: string;
        shadow?: string;
        color: string;
        overlay?: BackgroundLayer;
        backdropFilter?: string;
        overflow?: string;
    };
    typography: {
        title: string;
        subtitle: string;
        body: string;
        caption: string;
        strong: string;
        link: string;
        linkHover: string;
        brand: string;
    };
    buttons: {
        containedGradient: string;
        containedHover: string;
        text: string;
        outlinedBorder: string;
        outlinedHover: string;
        outlinedBackground: string;
        outlinedHoverBackground: string;
        radius: string;
        padding: string;
        shadow: string;
        shadowHover: string;
    };
    form: {
        background: string;
        border: string;
        borderHover: string;
        borderFocus: string;
        focusShadow: string;
        input: string;
        label: string;
        labelFocus: string;
    };
    surface: {
        backdrop: string;
        shadow: string;
    };
    status: {
        alertBackground: string;
        alertBorder: string;
        alertText: string;
        progress: string;
        divider: string;
        chipBackground: string;
        chipText: string;
        avatarGradient: string;
        tooltipBackground: string;
        tooltipBorder: string;
        tableBorder: string;
        tableText: string;
        iconButton: string;
        iconButtonHover: string;
        adornment: string;
    };
    layout?: {
        pageInset?: string;
        rootInset?: string;
        cardVariant?: "default" | "minimal" | "panel";
        cardClipPath?: string;
        maxWidth?: "xs" | "sm" | "md" | "lg" | "xl" | false;
        rootJustify?: CSSProperties["justifyContent"];
        rootAlign?: CSSProperties["alignItems"];
        containerMaxWidth?: CSSProperties["maxWidth"];
        containerWidth?: CSSProperties["width"];
        containerMargin?: CSSProperties["margin"];
    };
    animations?: Record<string, KeyframeMap>;
};

export interface PortalTemplateSummary {
    name: PortalTemplateName;
    displayName: string;
    description: string;
    interactive?: "pointer";
}

export interface PortalTemplateDefinition extends PortalTemplateSummary {
    style: PortalStyleConfig;
    effect?: PortalTemplateEffectDefinition;
}

const defaultTemplate: PortalTemplateDefinition = {
    name: "default",
    displayName: "Authelia Default",
    description: "Baseline Authelia login layout with a neutral gradient backdrop.",
    style: {
        page: {
            background: "linear-gradient(135deg, #1d1f2f 0%, #212640 50%, #151728 100%)",
            color: "#f2f6ff",
        },
        root: {
            padding: "3rem 1.5rem",
        },
        card: {
            background: "rgba(18, 24, 48, 0.88)",
            borderRadius: "18px",
            padding: "2.5rem 2rem",
            color: "#f2f6ff",
        },
        typography: {
            title: "#ffffff",
            subtitle: "rgba(210, 220, 255, 0.82)",
            body: "rgba(204, 214, 244, 0.78)",
            caption: "rgba(180, 190, 230, 0.72)",
            strong: "#ffffff",
            link: "#8fc8ff",
            linkHover: "#d8f0ff",
            brand: "rgba(172, 186, 232, 0.7)",
        },
        buttons: {
            containedGradient: "linear-gradient(130deg, #5c7bff 0%, #6b8dff 50%, #6ec8ff 100%)",
            containedHover: "linear-gradient(130deg, #6b88ff 0%, #7c9cff 50%, #7bd6ff 100%)",
            text: "#0f182e",
            outlinedBorder: "rgba(136, 170, 255, 0.6)",
            outlinedHover: "rgba(136, 170, 255, 0.9)",
            outlinedBackground: "rgba(136, 170, 255, 0.14)",
            outlinedHoverBackground: "rgba(136, 170, 255, 0.24)",
            radius: "16px",
            padding: "0.85rem 1.4rem",
            shadow: "0 16px 36px -20px rgba(96, 140, 255, 0.6)",
            shadowHover: "0 24px 44px -22px rgba(96, 180, 255, 0.65)",
        },
        form: {
            background: "rgba(14, 20, 42, 0.9)",
            border: "1px solid rgba(116, 150, 228, 0.36)",
            borderHover: "rgba(136, 170, 240, 0.52)",
            borderFocus: "rgba(140, 200, 255, 0.86)",
            focusShadow: "0 0 0 3px rgba(120, 180, 255, 0.22)",
            input: "#f8fbff",
            label: "rgba(184, 198, 238, 0.74)",
            labelFocus: "#8fe0ff",
        },
        surface: {
            backdrop: "rgba(10, 14, 30, 0.72)",
            shadow: "0 32px 72px -36px rgba(8, 12, 30, 0.9)",
        },
        status: {
            alertBackground: "rgba(255, 148, 214, 0.18)",
            alertBorder: "rgba(255, 148, 214, 0.38)",
            alertText: "#ffe1f1",
            progress: "#7bccff",
            divider: "rgba(132, 176, 236, 0.32)",
            chipBackground: "rgba(136, 186, 255, 0.24)",
            chipText: "#e4f2ff",
            avatarGradient: "linear-gradient(135deg, rgba(128, 176, 255, 0.78), rgba(118, 216, 255, 0.7))",
            tooltipBackground: "rgba(9, 12, 28, 0.94)",
            tooltipBorder: "rgba(134, 182, 255, 0.34)",
            tableBorder: "rgba(132, 178, 236, 0.32)",
            tableText: "rgba(212, 226, 252, 0.84)",
            iconButton: "rgba(206, 220, 255, 0.84)",
            iconButtonHover: "#8ee6ff",
            adornment: "rgba(170, 198, 252, 0.74)",
        },
        layout: {
            pageInset: "clamp(1.6rem, 4vw, 3.2rem)",
            rootInset: "clamp(1.4rem, 3.2vw, 2.8rem)",
            cardVariant: "default",
            maxWidth: "md",
            rootJustify: "center",
            rootAlign: "center",
        },
    },
};

export const portalTemplates: Record<PortalTemplateName, PortalTemplateDefinition> = {
    default: defaultTemplate,
};

export const defaultTemplateManifest: PortalTemplateSummary[] = Object.values(portalTemplates).map(
    ({ name, displayName, description, interactive }) => ({
        name,
        displayName,
        description,
        interactive,
    }),
);
