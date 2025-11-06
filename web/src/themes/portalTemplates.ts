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
        paddingTop?: string;
        paddingBottom?: string;
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
        rootInsetTop?: string;
        rootInsetBottom?: string;
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
    stylePath?: string;
    definitionPath?: string;
    effectPath?: string;
}

export interface PortalTemplateDefinition extends PortalTemplateSummary {
    style: PortalStyleConfig;
    effect?: PortalTemplateEffectDefinition;
}

const defaultTemplate: PortalTemplateDefinition = {
    name: "default",
    displayName: "Authelia Default",
    description: "Original Authelia login appearance using the stock Material UI theme.",
    style: {
        page: {
            background: "#ffffff",
            color: "rgba(0, 0, 0, 0.87)",
        },
        root: {
            padding: "3rem 1.5rem",
        },
        card: {
            background: "transparent",
            border: "none",
            borderRadius: "0px",
            padding: "0",
            color: "rgba(0, 0, 0, 0.87)",
            shadow: "none",
        },
        typography: {
            title: "rgba(0, 0, 0, 0.87)",
            subtitle: "rgba(0, 0, 0, 0.6)",
            body: "rgba(0, 0, 0, 0.87)",
            caption: "rgba(0, 0, 0, 0.6)",
            strong: "rgba(0, 0, 0, 0.87)",
            link: "#1976d2",
            linkHover: "#115293",
            brand: "#adb5bd",
        },
        buttons: {
            containedGradient: "#1976d2",
            containedHover: "#115293",
            text: "#ffffff",
            outlinedBorder: "rgba(25, 118, 210, 0.8)",
            outlinedHover: "rgba(17, 82, 147, 0.9)",
            outlinedBackground: "transparent",
            outlinedHoverBackground: "rgba(25, 118, 210, 0.04)",
            radius: "4px",
            padding: "0.75rem 1.5rem",
            shadow: "none",
            shadowHover: "none",
        },
        form: {
            background: "#ffffff",
            border: "1px solid rgba(0, 0, 0, 0.23)",
            borderHover: "rgba(25, 118, 210, 0.8)",
            borderFocus: "rgba(25, 118, 210, 1)",
            focusShadow: "none",
            input: "rgba(0, 0, 0, 0.87)",
            label: "rgba(0, 0, 0, 0.6)",
            labelFocus: "#1976d2",
        },
        surface: {
            backdrop: "rgba(255, 255, 255, 0.9)",
            shadow: "none",
        },
        status: {
            alertBackground: "rgba(229, 246, 253, 0.9)",
            alertBorder: "rgba(2, 136, 209, 0.3)",
            alertText: "rgba(2, 136, 209, 1)",
            progress: "#1976d2",
            divider: "rgba(0, 0, 0, 0.12)",
            chipBackground: "rgba(25, 118, 210, 0.08)",
            chipText: "#0d47a1",
            avatarGradient: "linear-gradient(135deg, rgba(38, 166, 154, 0.7), rgba(3, 155, 229, 0.7))",
            tooltipBackground: "rgba(97, 97, 97, 0.9)",
            tooltipBorder: "rgba(189, 189, 189, 0.4)",
            tableBorder: "rgba(224, 224, 224, 1)",
            tableText: "rgba(33, 33, 33, 0.87)",
            iconButton: "rgba(0, 0, 0, 0.54)",
            iconButtonHover: "rgba(0, 0, 0, 0.87)",
            adornment: "rgba(0, 0, 0, 0.12)",
        },
        layout: {
            cardVariant: "minimal",
            maxWidth: "xs",
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
