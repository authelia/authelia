import React from "react";

import { render, screen } from "@testing-library/react";

import PasskeyIcon from "@components/PasskeyIcon";

it("renders without crashing", () => {
    render(<PasskeyIcon />);
});

it("renders svg icon", () => {
    render(<PasskeyIcon />);
    const svg = document.querySelector("svg");
    expect(svg).toBeInTheDocument();
});

it("has correct viewbox", () => {
    render(<PasskeyIcon />);
    const svg = document.querySelector("svg");
    expect(svg).toHaveAttribute("viewBox", "0 -960 960 960");
});

it("has correct fill", () => {
    render(<PasskeyIcon />);
    const svg = document.querySelector("svg");
    expect(svg).toHaveAttribute("fill", "#e8eaed");
});
