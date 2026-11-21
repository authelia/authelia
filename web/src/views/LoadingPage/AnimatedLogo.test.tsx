import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { AnimatedLogo } from "./AnimatedLogo";

describe("AutheliaLogo", () => {
    it("renders an svg with the default size", () => {
        render(<AnimatedLogo />);

        const svg = screen.getByRole("img", { name: /authelia logo/i });

        expect(svg).toHaveAttribute("width", "170");
        expect(svg).toHaveAttribute("height", "170");
    });

    it("applies a custom size and duration", () => {
        render(<AnimatedLogo size={50} duration={7} />);

        const svg = screen.getByRole("img", { name: /authelia logo/i });

        expect(svg).toHaveAttribute("width", "50");
        expect(svg.style.getPropertyValue("--duration")).toBe("7s");
    });
});
