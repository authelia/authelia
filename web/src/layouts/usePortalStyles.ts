import { useEffect } from "react";

import { Theme } from "@mui/material";
import { makeStyles } from "tss-react/mui";

import { BackgroundLayer, PortalStyleConfig, PortalTemplateDefinition } from "@themes/portalTemplates";

type StyleParams = {
    config: PortalStyleConfig;
    template: PortalTemplateDefinition;
};

const useLegacyStyles = makeStyles({ name: "PortalLegacy" })((theme: Theme) => ({
    page: {
        minHeight: "100vh",
        background: theme.palette.background.default,
        color: theme.palette.text.primary,
        display: "flex",
        flexDirection: "column",
    },
    effectHost: {
        display: "none",
    },
    root: {
        minHeight: "90vh",
        textAlign: "center",
        padding: "3rem 1.5rem",
        background: "transparent",
        alignItems: "center",
        justifyContent: "center",
    },
    rootContainer: {
        paddingLeft: 32,
        paddingRight: 32,
        background: "transparent",
        border: "none",
        borderRadius: 0,
        boxShadow: "none",
    },
    icon: {
        margin: theme.spacing(),
        width: "64px",
        fill: theme.custom.icon,
    },
    body: {
        marginTop: theme.spacing(),
        paddingTop: theme.spacing(),
        paddingBottom: theme.spacing(),
    },
    typography: {
        "& .MuiTypography-h5, & .MuiTypography-h4, & .MuiTypography-h3": {
            color: theme.palette.text.primary,
            fontWeight: theme.typography.fontWeightMedium,
            letterSpacing: theme.typography.h5.letterSpacing,
            textShadow: "none",
        },
        "& .MuiTypography-h6": {
            color: theme.palette.text.secondary,
            fontWeight: theme.typography.fontWeightRegular,
            marginTop: theme.spacing(0.5),
        },
        "& .MuiTypography-root": {
            color: theme.palette.text.primary,
            letterSpacing: theme.typography.body1.letterSpacing,
        },
        "& .MuiTypography-root strong": {
            color: theme.palette.text.primary,
        },
    },
    links: {
        "& .MuiLink-root": {
            color: theme.palette.primary.main,
            fontWeight: theme.typography.fontWeightMedium,
            textShadow: "none",
        },
        "& .MuiLink-root:hover": {
            color: theme.palette.primary.dark,
            textShadow: "none",
        },
    },
    formElements: {
        "& .MuiOutlinedInput-root": {
            background: theme.palette.background.paper,
            borderRadius: theme.shape.borderRadius,
            backdropFilter: "none",
        },
        "& .MuiOutlinedInput-notchedOutline": {
            borderColor: theme.palette.divider,
        },
        "& .MuiOutlinedInput-root.Mui-focused .MuiOutlinedInput-notchedOutline": {
            borderColor: theme.palette.primary.main,
            boxShadow: "none",
        },
        "& .MuiOutlinedInput-input": {
            color: theme.palette.text.primary,
        },
        "& .MuiInputLabel-root": {
            color: theme.palette.text.secondary,
        },
        "& .MuiInputLabel-root.Mui-focused": {
            color: theme.palette.primary.main,
        },
    },
    buttons: {
        "& .MuiButton-contained": {
            background: theme.palette.primary.main,
            color: theme.palette.primary.contrastText,
            borderRadius: theme.shape.borderRadius,
            padding: theme.spacing(1, 2.5),
            boxShadow: theme.shadows[2],
        },
        "& .MuiButton-contained:hover": {
            background: theme.palette.primary.dark,
            boxShadow: theme.shadows[4],
        },
        "& .MuiButton-outlined": {
            color: theme.palette.primary.main,
            borderColor: theme.palette.primary.main,
            borderRadius: theme.shape.borderRadius,
            padding: theme.spacing(1, 2.5),
        },
        "& .MuiButton-outlined:hover": {
            borderColor: theme.palette.primary.dark,
        },
    },
    status: {
        "& .MuiAlert-root": {
            background: theme.palette.background.paper,
            border: `1px solid ${theme.palette.divider}`,
            color: theme.palette.text.primary,
        },
        "& .MuiLinearProgress-bar": {
            background: theme.palette.primary.main,
        },
        "& .MuiDivider-root": {
            background: theme.palette.divider,
        },
    },
}));

const buildLayerStyles = (layer: BackgroundLayer | undefined, templateName: string) => {
    if (!layer) {
        return undefined;
    }

    const {
        background,
        opacity,
        blur,
        mixBlendMode,
        animation,
        inset,
        position,
        size,
        zIndex,
        pointerEvents,
        transform,
    } = layer;
    const animationName = animation ? `portal-${templateName}-${animation.key}` : undefined;

    return {
        content: "''",
        position: position ?? "fixed",
        inset: inset ?? "-40vh -40vw",
        background,
        opacity,
        filter: blur ? `blur(${blur})` : undefined,
        mixBlendMode,
        pointerEvents: pointerEvents ?? "none",
        zIndex: zIndex ?? 0,
        backgroundSize: size,
        transform,
        animationName,
        animationDuration: animation?.duration,
        animationTimingFunction: animation?.timingFunction,
        animationIterationCount: animation?.iterationCount ?? (animation ? "infinite" : undefined),
        animationDirection: animation?.direction,
        animationDelay: animation?.delay,
    } as const;
};

const useStyles = makeStyles<StyleParams>({ name: "PortalTemplates" })((theme: Theme, { config, template }) => {
    const keyframes = Object.entries(config.animations ?? {}).reduce<Record<string, Record<string, string | number>>>(
        (accumulator, [key, frames]) => {
            accumulator[`@keyframes portal-${template.name}-${key}`] = frames as Record<string, string | number>;
            return accumulator;
        },
        {},
    );
    const variant = config.layout?.cardVariant ?? "default";
    const cardBackdrop = config.card.backdropFilter ?? (variant === "minimal" ? "none" : "blur(28px)");
    const cardOverflow = config.card.overflow ?? (variant === "minimal" ? "visible" : "hidden");
    const cardOverlay = variant === "minimal" ? undefined : config.card.overlay;
    const rootJustify = config.layout?.rootJustify ?? "flex-start";
    const rootAlign = config.layout?.rootAlign ?? "flex-start";
    const rootPadding = config.root.padding ?? config.layout?.rootInset ?? "0 1.5rem 3rem";
    const hasExplicitRootPadding = config.root.padding !== undefined;
    const rootPaddingTop =
        config.root.paddingTop ?? config.layout?.rootInsetTop ?? (hasExplicitRootPadding ? "0" : undefined) ?? "0";
    const rootPaddingBottom = config.root.paddingBottom ?? config.layout?.rootInsetBottom;
    const containerMaxWidth =
        config.layout?.containerMaxWidth ?? (variant === "minimal" ? "min(1080px, 100%)" : undefined);
    const containerMargin = config.layout?.containerMargin ?? (variant === "minimal" ? "0 auto" : undefined);
    const containerWidth = config.layout?.containerWidth;

    return {
        ...keyframes,
        page: {
            position: "relative",
            overflow: "hidden",
            background: config.page.background,
            color: config.page.color,
            display: "flex",
            flexDirection: "column",
            minHeight: "100vh",
            ...(config.layout?.pageInset ? { padding: config.layout.pageInset } : {}),
            "&::before": buildLayerStyles(config.page.before, template.name),
            "&::after": buildLayerStyles(config.page.after, template.name),
        },
        effectHost: {
            position: "absolute",
            inset: 0,
            pointerEvents: "none",
            zIndex: 0,
            overflow: "hidden",
        },
        root: {
            position: "relative",
            zIndex: 1,
            textAlign: variant === "minimal" ? "left" : "center",
            isolation: "isolate",
            padding: rootPadding,
            ...(rootPaddingTop === undefined ? {} : { paddingTop: rootPaddingTop }),
            ...(rootPaddingBottom === undefined ? {} : { paddingBottom: rootPaddingBottom }),
            display: "flex",
            alignItems: rootAlign,
            justifyContent: rootJustify,
            background: config.root.background,
            "&::before": buildLayerStyles(config.root.before, template.name),
            "&::after": buildLayerStyles(config.root.after, template.name),
        },
        rootContainer: {
            position: "relative",
            overflow: cardOverflow,
            background: config.card.background,
            border: config.card.border,
            borderRadius: config.card.borderRadius,
            padding: config.card.padding,
            color: config.card.color,
            boxShadow: config.card.shadow,
            backdropFilter: cardBackdrop,
            clipPath: config.layout?.cardClipPath,
            margin: containerMargin,
            maxWidth: containerMaxWidth,
            width: containerWidth,
            "&::before": buildLayerStyles(cardOverlay, template.name),
            ...(variant !== "minimal"
                ? {
                      [theme.breakpoints.down("md")]: {
                          padding: "clamp(1.8rem, 6vw, 2.2rem)",
                          borderRadius: config.card.borderRadius,
                      },
                      [theme.breakpoints.down("sm")]: {
                          padding: "1.75rem 1.4rem",
                      },
                  }
                : {}),
            ...(variant === "panel"
                ? {
                      display: "grid",
                      gridTemplateColumns: "minmax(0, 1.05fr) minmax(0, 0.85fr)",
                      gap: "clamp(1.6rem, 4vw, 2.6rem)",
                      alignItems: "stretch",
                      padding: config.card.padding ?? "clamp(2.4rem, 5vw, 3rem)",
                      position: "relative" as const,
                      overflow: cardOverflow,
                      "&::after": {
                          content: "''",
                          position: "absolute",
                          inset: "-25% -45% -25% 55%",
                          background:
                              "linear-gradient(165deg, rgba(0, 255, 170, 0.25) 0%, rgba(0, 211, 255, 0.22) 60%, rgba(0, 98, 255, 0) 100%)",
                          opacity: 0.45,
                          pointerEvents: "none",
                          mixBlendMode: "screen",
                      },
                  }
                : {}),
            ...(variant === "minimal"
                ? {
                      background: config.card.background ?? "transparent",
                      border: config.card.border ?? "none",
                      borderRadius: config.card.borderRadius ?? "0px",
                      boxShadow: config.card.shadow ?? "none",
                      padding: config.card.padding ?? "0",
                      backdropFilter: cardBackdrop,
                      overflow: cardOverflow,
                  }
                : {}),
        },
        icon: {
            margin: theme.spacing(),
            width: "64px",
            fill: config.typography.title,
            filter: "drop-shadow(0 14px 25px rgba(79, 126, 255, 0.45))",
        },
        body: {
            marginTop: theme.spacing(),
            paddingTop: theme.spacing(),
            paddingBottom: theme.spacing(),
            display: "flex",
            flexDirection: "column",
            gap: "1.1rem",
            alignItems: variant === "minimal" ? "flex-start" : "stretch",
        },
        typography: {
            "& .MuiTypography-h5, & .MuiTypography-h4, & .MuiTypography-h3": {
                color: config.typography.title,
                fontWeight: 700,
                letterSpacing: "0.03em",
                textShadow: "0 14px 35px rgba(26, 48, 111, 0.7)",
            },
            "& .MuiTypography-h6": {
                color: config.typography.subtitle,
                fontWeight: 500,
                marginTop: "0.3rem",
            },
            "& .MuiTypography-root": {
                color: config.typography.body,
                letterSpacing: "0.015em",
            },
            "& .MuiTypography-root strong": {
                color: config.typography.strong,
            },
        },
        links: {
            "& .MuiLink-root": {
                color: config.typography.link,
                fontWeight: 500,
                transition: "color 0.25s ease, text-shadow 0.3s ease",
            },
            "& .MuiLink-root:hover": {
                color: config.typography.linkHover,
                textShadow: "0 0 16px rgba(104, 192, 255, 0.6)",
            },
        },
        formElements: {
            "& .MuiOutlinedInput-root": {
                background: config.form.background,
                borderRadius: "18px",
                backdropFilter: "blur(18px)",
                transition: "transform 0.2s ease, box-shadow 0.3s ease, border-color 0.3s ease",
            },
            "& .MuiOutlinedInput-notchedOutline": {
                borderColor: config.form.border,
            },
            "& .MuiOutlinedInput-root:hover .MuiOutlinedInput-notchedOutline": {
                borderColor: config.form.borderHover,
            },
            "& .MuiOutlinedInput-root.Mui-focused .MuiOutlinedInput-notchedOutline": {
                borderColor: config.form.borderFocus,
                boxShadow: config.form.focusShadow,
            },
            "& .MuiOutlinedInput-input": {
                color: config.form.input,
            },
            "& .MuiInputLabel-root": {
                color: config.form.label,
            },
            "& .MuiInputLabel-root.Mui-focused": {
                color: config.form.labelFocus,
            },
        },
        buttons: {
            "& .MuiButton-contained": {
                background: config.buttons.containedGradient,
                color: config.buttons.text,
                borderRadius: config.buttons.radius,
                padding: config.buttons.padding,
                boxShadow: config.buttons.shadow,
                transition: "box-shadow 0.3s ease, transform 0.2s ease",
            },
            "& .MuiButton-contained:hover": {
                background: config.buttons.containedHover,
                boxShadow: config.buttons.shadowHover,
                transform: "translateY(-2px)",
            },
            "& .MuiButton-outlined": {
                color: config.buttons.text,
                borderColor: config.buttons.outlinedBorder,
                background: config.buttons.outlinedBackground,
                borderRadius: config.buttons.radius,
                padding: config.buttons.padding,
            },
            "& .MuiButton-outlined:hover": {
                borderColor: config.buttons.outlinedHover,
                background: config.buttons.outlinedHoverBackground,
            },
        },
        status: {
            "& .MuiAlert-root": {
                background: config.status.alertBackground,
                border: `1px solid ${config.status.alertBorder}`,
                color: config.status.alertText,
            },
            "& .MuiLinearProgress-bar": {
                background: config.status.progress,
            },
            "& .MuiDivider-root": {
                background: config.status.divider,
            },
        },
    };
});

export const usePortalStyles = (template: PortalTemplateDefinition) => {
    const { classes: templateClasses } = useStyles({ config: template.style, template });
    const { classes: legacyClasses } = useLegacyStyles();

    useEffect(() => {
        if (typeof document === "undefined") {
            return;
        }

        document.body.dataset.portalTemplate = template.name;

        if (template.name === "default") {
            document.body.style.background = "";
            document.body.style.color = "";
            return () => {
                document.body.style.background = "";
                document.body.style.color = "";
                delete document.body.dataset.portalTemplate;
            };
        }

        document.body.style.background = template.style.page.background;
        document.body.style.color = template.style.page.color;

        return () => {
            document.body.style.background = "";
            document.body.style.color = "";
            delete document.body.dataset.portalTemplate;
        };
    }, [template]);

    return template.name === "default" ? legacyClasses : templateClasses;
};
