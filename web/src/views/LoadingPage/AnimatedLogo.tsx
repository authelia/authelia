import { type CSSProperties } from "react";

interface AutheliaLogoProps {
    /**
     * Logo size in pixels.
     *
     * Default to `170`.
     */
    size?: number;
    /**
     * Animation duration in seconds.
     *
     * Default to `3s`.
     */
    duration?: number;
}

/**
 * Animated recreation of the Authelia keyhole logo.
 * - Center keyhole: static
 * - Middle ring: rotates clockwise
 * - Outer ring: same animation (speed curve + flow), mirrored counter-clockwise
 */
export function AnimatedLogo({ duration = 3, size = 170 }: AutheliaLogoProps) {
    return (
        <svg
            width={size}
            height={size}
            viewBox="0 0 200 200"
            xmlns="http://www.w3.org/2000/svg"
            role="img"
            aria-label="Authelia logo"
            style={{ "--duration": duration + "s" } as CSSProperties}
        >
            <defs>
                <linearGradient id="grad-center" x1="0%" y1="100%" x2="100%" y2="0%">
                    <stop offset="0%" stopColor="#3b4bb0" />
                    <stop offset="100%" stopColor="#19275b" />
                </linearGradient>
                <linearGradient id="grad-middle" x1="0%" y1="0%" x2="100%" y2="100%">
                    <stop offset="0%" stopColor="#4c5fd7" />
                    <stop offset="100%" stopColor="#16265a" />
                </linearGradient>
                <linearGradient id="grad-outer" x1="0%" y1="0%" x2="100%" y2="100%">
                    <stop offset="0%" stopColor="#4c5fd7" />
                    <stop offset="100%" stopColor="#0d1b3e" />
                </linearGradient>

                <mask id="keyhole-mask">
                    <circle cx="100" cy="100" r="40" fill="white" />
                    <circle cx="100" cy="86" r="10" fill="black" />
                    <path d="M96.5 95 L103.5 95 L110 118 C110 129 90 129 90 118 Z" fill="black" />
                </mask>
            </defs>

            {/* Middle broken ring - clockwise */}
            <circle
                className="origin-center not-motion-reduce:animate-[spin_var(--duration)_linear_infinite]"
                cx="100"
                cy="100"
                r="56"
                fill="none"
                stroke="url(#grad-middle)"
                strokeWidth="12"
                strokeLinecap="round"
                strokeDasharray="66 48"
            />

            {/* Outer broken ring - counter-clockwise */}
            <circle
                className="origin-center not-motion-reduce:animate-[spin_var(--duration)_linear_infinite] not-motion-reduce:[animation-direction:reverse]"
                cx="100"
                cy="100"
                r="78"
                fill="none"
                stroke="url(#grad-outer)"
                strokeWidth="12"
                strokeLinecap="round"
                strokeDasharray="88 62"
            />

            {/* Static center: filled circle with the keyhole cut out as a real hole */}
            <circle cx="100" cy="100" r="40" fill="url(#grad-center)" mask="url(#keyhole-mask)" />
        </svg>
    );
}
